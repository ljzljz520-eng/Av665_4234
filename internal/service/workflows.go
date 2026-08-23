package service

import (
	"errors"
	"fmt"
	"io"

	"fire-equipment-control/internal/domain"
	"fire-equipment-control/internal/importexport"
)

type ArchiveWorkflow struct {
	Record      domain.EquipmentRecord
	ReviewCount int
	AuditCount  int
}

type CollaborationWorkflow struct {
	Record  domain.EquipmentRecord
	Summary string
}

type ImportWorkflow struct {
	Outcome domain.ImportOutcome
	Report  string
}

func (s *Service) CreateReviewArchive(record domain.EquipmentRecord, actor string) (ArchiveWorkflow, error) {
	created, err := s.RegisterEquipment(record)
	if err != nil {
		return ArchiveWorkflow{}, err
	}
	if _, err := s.SubmitForReview(created.Code, actor, "登记资料已提交"); err != nil {
		return ArchiveWorkflow{}, err
	}
	if _, err := s.ApproveReview(created.Code, actor, "现场复核通过"); err != nil {
		return ArchiveWorkflow{}, err
	}
	if _, err := s.ConfirmEquipment(created.Code, actor); err != nil {
		return ArchiveWorkflow{}, err
	}
	archived, err := s.ArchiveEquipment(created.Code, actor)
	if err != nil {
		return ArchiveWorkflow{}, err
	}
	reviews, err := s.Reviews(created.Code)
	if err != nil {
		return ArchiveWorkflow{}, err
	}
	events, err := s.AuditTrail(created.Code)
	if err != nil {
		return ArchiveWorkflow{}, err
	}
	return ArchiveWorkflow{Record: archived, ReviewCount: len(reviews), AuditCount: len(events)}, nil
}

func (s *Service) SearchUpdatePublish(filter domain.SearchFilter, code, actor, owner string) (CollaborationWorkflow, error) {
	records, err := s.SearchEquipment(filter)
	if err != nil {
		return CollaborationWorkflow{}, err
	}
	if len(records) == 0 {
		return CollaborationWorkflow{}, ErrEquipmentMissing
	}
	selected, err := s.GetEquipment(code)
	if err != nil {
		return CollaborationWorkflow{}, err
	}
	if selected.Code != records[0].Code && code == "" {
		code = records[0].Code
	}
	updated, err := s.CollaborateOwnerChange(code, actor, owner)
	if err != nil {
		return CollaborationWorkflow{}, err
	}
	summary, err := s.PublishChangeSummary(updated.Code)
	if err != nil {
		return CollaborationWorkflow{}, err
	}
	return CollaborationWorkflow{Record: updated, Summary: summary}, nil
}

func (s *Service) ImportReport(reader io.Reader) (ImportWorkflow, error) {
	outcome, err := s.ImportCSV(reader)
	if err != nil {
		return ImportWorkflow{}, err
	}
	report, err := importexport.ExportReport(outcome)
	if err != nil {
		return ImportWorkflow{}, err
	}
	if len(outcome.Accepted) == 0 && len(outcome.Issues) == 0 {
		return ImportWorkflow{}, errors.New("import contained no data")
	}
	return ImportWorkflow{Outcome: outcome, Report: report}, nil
}

func (s *Service) DeactivateAndRetire(code, actor string) (domain.EquipmentRecord, error) {
	if _, err := s.DeactivateEquipment(code, actor, "检查发现需要更换"); err != nil {
		return domain.EquipmentRecord{}, err
	}
	return s.RetireEquipment(code, actor, "设备达到报废条件")
}

func (s *Service) LifecycleSummary(code string) (string, error) {
	record, err := s.GetEquipment(code)
	if err != nil {
		return "", err
	}
	events, err := s.AuditTrail(code)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %d events", record.Code, domain.StatusLabel(record.Status), len(events)), nil
}

func (s *Service) ValidateImportHeader(reader io.Reader) error {
	rows, err := importexport.ParseCSV(reader)
	if err != nil {
		return err
	}
	if rows == nil {
		return errors.New("import header parsed without rows")
	}
	return nil
}
