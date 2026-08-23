package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"fire-equipment-control/internal/attachment"
	"fire-equipment-control/internal/audit"
	"fire-equipment-control/internal/dashboard"
	"fire-equipment-control/internal/domain"
	"fire-equipment-control/internal/importexport"
	"fire-equipment-control/internal/inspection"
	"fire-equipment-control/internal/query"
	"fire-equipment-control/internal/registry"
	"fire-equipment-control/internal/report"
	"fire-equipment-control/internal/storage"
)

var (
	ErrAlreadyExists        = errors.New("equipment already exists")
	ErrDuplicateConsumption = errors.New("pending confirmation consumed twice")
	ErrOwnerProposalEmpty   = errors.New("owner proposal is empty")
	ErrEquipmentMissing     = errors.New("equipment does not exist")
)

type Service struct {
	store    *storage.Store
	mu       sync.Mutex
	consumed map[string]int
}

func New(store *storage.Store) *Service {
	return &Service{store: store, consumed: map[string]int{}}
}

func (s *Service) Store() *storage.Store {
	return s.store
}

func (s *Service) RegisterEquipment(record domain.EquipmentRecord) (domain.EquipmentRecord, error) {
	record = record.Normalized()
	if record.Status != domain.StatusDraft {
		return domain.EquipmentRecord{}, errors.New("new equipment must start as draft")
	}
	record.Type = registry.NormalizeType(record.Type)
	if err := registry.ValidateRecord(record); err != nil {
		return domain.EquipmentRecord{}, err
	}
	if _, err := s.store.GetEquipment(record.Code); err == nil {
		return domain.EquipmentRecord{}, ErrAlreadyExists
	} else if !errors.Is(err, storage.ErrNotFound) {
		return domain.EquipmentRecord{}, err
	}
	if record.ID == "" {
		record.ID = "equipment-" + record.Code
	}
	if err := s.store.SaveEquipment(record); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return record, nil
}

func (s *Service) GetEquipment(code string) (domain.EquipmentRecord, error) {
	record, err := s.store.GetEquipment(strings.TrimSpace(code))
	if errors.Is(err, storage.ErrNotFound) {
		return domain.EquipmentRecord{}, ErrEquipmentMissing
	}
	return record, err
}

func (s *Service) SubmitForReview(code, actor, comment string) (domain.Review, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.Review{}, err
	}
	updated, err := domain.Transition(record, domain.StatusReview)
	if err != nil {
		return domain.Review{}, err
	}
	if err := s.store.SaveEquipment(updated); err != nil {
		return domain.Review{}, err
	}
	sequence, err := s.store.NextSequence()
	if err != nil {
		return domain.Review{}, err
	}
	review := domain.Review{ID: fmt.Sprintf("review-%04d", sequence), EquipmentCode: record.Code, Stage: "submitted", Actor: strings.TrimSpace(actor), Comment: strings.TrimSpace(comment), CreatedAt: fmt.Sprintf("sequence-%04d", sequence)}
	if err := s.store.SaveReview(review); err != nil {
		return domain.Review{}, err
	}
	if err := s.store.SaveAuditEvent(audit.NewReviewEvent(record.Code, actor, "submitted", comment, sequence)); err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Service) ApproveReview(code, actor, comment string) (domain.Review, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.Review{}, err
	}
	updated, err := domain.Transition(record, domain.StatusApproved)
	if err != nil {
		return domain.Review{}, err
	}
	if err := s.store.SaveEquipment(updated); err != nil {
		return domain.Review{}, err
	}
	sequence, err := s.store.NextSequence()
	if err != nil {
		return domain.Review{}, err
	}
	review := domain.Review{ID: fmt.Sprintf("review-%04d", sequence), EquipmentCode: record.Code, Stage: "approved", Actor: strings.TrimSpace(actor), Comment: strings.TrimSpace(comment), CreatedAt: fmt.Sprintf("sequence-%04d", sequence)}
	if err := s.store.SaveReview(review); err != nil {
		return domain.Review{}, err
	}
	if err := s.store.SaveAuditEvent(audit.NewReviewEvent(record.Code, actor, "approved", comment, sequence)); err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Service) ConfirmEquipment(code, actor string) (domain.EquipmentRecord, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	updated, err := domain.Transition(record, domain.StatusActive)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.store.SaveEquipment(updated); err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.writeLifecycle(record.Code, actor, "confirm", "approved equipment confirmed"); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return updated, nil
}

func (s *Service) ArchiveEquipment(code, actor string) (domain.EquipmentRecord, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	updated, err := domain.Transition(record, domain.StatusArchived)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.store.SaveEquipment(updated); err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.writeLifecycle(record.Code, actor, "archive", "record archived after lifecycle close"); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return updated, nil
}

func (s *Service) DeactivateEquipment(code, actor, reason string) (domain.EquipmentRecord, error) {
	return s.changeLifecycle(code, actor, domain.StatusDisabled, "disable", reason)
}

func (s *Service) RetireEquipment(code, actor, reason string) (domain.EquipmentRecord, error) {
	return s.changeLifecycle(code, actor, domain.StatusRetired, "retire", reason)
}

func (s *Service) changeLifecycle(code, actor string, target domain.Status, action, detail string) (domain.EquipmentRecord, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	updated, err := domain.Transition(record, target)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.store.SaveEquipment(updated); err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.writeLifecycle(record.Code, actor, action, detail); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return updated, nil
}

func (s *Service) writeLifecycle(code, actor, action, detail string) error {
	sequence, err := s.store.NextSequence()
	if err != nil {
		return err
	}
	return s.store.SaveAuditEvent(audit.NewLifecycleEvent(code, actor, action, detail, sequence))
}

func (s *Service) AttachPhoto(code, id, name, path, mediaType, content string) (domain.Attachment, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.Attachment{}, err
	}
	if err := registry.ValidateType(record.Type); err != nil {
		return domain.Attachment{}, err
	}
	item, err := attachment.BuildPhoto(record.Code, id, name, path, mediaType, content)
	if err != nil {
		return domain.Attachment{}, err
	}
	if err := s.store.SaveAttachment(item); err != nil {
		return domain.Attachment{}, err
	}
	return item, nil
}

func (s *Service) SearchEquipment(filter domain.SearchFilter) ([]domain.EquipmentRecord, error) {
	records, err := s.store.ListEquipment()
	if err != nil {
		return nil, err
	}
	return query.Filter(records, filter), nil
}

func (s *Service) CollaborateOwnerChange(code, actor, proposedOwner string) (domain.EquipmentRecord, error) {
	if strings.TrimSpace(proposedOwner) == "" {
		return domain.EquipmentRecord{}, ErrOwnerProposalEmpty
	}
	record, err := s.GetEquipment(code)
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	before := record.Owner
	record.Owner = strings.TrimSpace(proposedOwner)
	record.Revision++
	if err := registry.ValidateRecord(record); err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.store.SaveEquipment(record); err != nil {
		return domain.EquipmentRecord{}, err
	}
	sequence, err := s.store.NextSequence()
	if err != nil {
		return domain.EquipmentRecord{}, err
	}
	if err := s.store.SaveAuditEvent(audit.NewChangeEvent(code, actor, before, record.Owner, sequence)); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return record, nil
}

func (s *Service) PublishChangeSummary(code string) (string, error) {
	records, err := s.SearchEquipment(query.BuildFilter(code, "", "", "", 0, ""))
	if err != nil {
		return "", err
	}
	events, err := s.store.ListAuditEvents(code)
	if err != nil {
		return "", err
	}
	return dashboard.Render(dashboard.Build(records, events)), nil
}

func (s *Service) ImportCSV(reader io.Reader) (domain.ImportOutcome, error) {
	rows, err := importexport.ParseCSV(reader)
	if err != nil {
		return domain.ImportOutcome{}, err
	}
	existingRecords, err := s.store.ListEquipment()
	if err != nil {
		return domain.ImportOutcome{}, err
	}
	existing := map[string]bool{}
	for _, record := range existingRecords {
		existing[record.Code] = true
	}
	outcome := importexport.ValidateRows(rows, existing)
	for _, record := range outcome.Accepted {
		if err := s.store.SaveEquipment(record); err != nil {
			return outcome, err
		}
	}
	return outcome, nil
}

func (s *Service) ExportInventory() (string, error) {
	records, err := s.store.ListEquipment()
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := report.InventoryCSV(records, &output); err != nil {
		return "", err
	}
	return output.String(), nil
}

func (s *Service) ExportAudit(code string) (string, error) {
	events, err := s.store.ListAuditEvents(code)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := report.AuditCSV(events, &output); err != nil {
		return "", err
	}
	return output.String(), nil
}

func (s *Service) AuditTrail(code string) ([]domain.AuditEvent, error) {
	return s.store.ListAuditEvents(code)
}

func (s *Service) Attachments(code string) ([]domain.Attachment, error) {
	return s.store.ListAttachments(code)
}

func (s *Service) Reviews(code string) ([]domain.Review, error) {
	return s.store.ListReviews(code)
}

func (s *Service) InspectEquipment(code, asOf string) (inspection.Assessment, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return inspection.Assessment{}, err
	}
	return inspection.Assess(record, asOf), nil
}

func (s *Service) InspectionPlan(asOf string) ([]inspection.PlanItem, error) {
	records, err := s.store.ListEquipment()
	if err != nil {
		return nil, err
	}
	return inspection.BuildPlan(records, asOf), nil
}

func (s *Service) CatalogTypes() []registry.TypeDescriptor {
	return registry.Types()
}

func (s *Service) CatalogCategory(category string) ([]domain.EquipmentRecord, error) {
	records, err := s.store.ListEquipment()
	if err != nil {
		return nil, err
	}
	return registry.FilterByCategory(records, category), nil
}
