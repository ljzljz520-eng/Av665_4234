package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"fire-equipment-control/internal/domain"
)

func InventoryCSV(records []domain.EquipmentRecord, writer io.Writer) error {
	ordered := append([]domain.EquipmentRecord(nil), records...)
	domain.SortRecords(ordered)
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"code", "type", "building", "floor", "inspection_date", "owner", "status", "notes"}); err != nil {
		return err
	}
	for _, record := range ordered {
		row := []string{record.Code, record.Type, record.Building, strconv.Itoa(record.Floor), record.InspectionDate, record.Owner, string(record.Status), record.Notes}
		if err := output.Write(row); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}

func AuditCSV(events []domain.AuditEvent, writer io.Writer) error {
	ordered := append([]domain.AuditEvent(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Sequence < ordered[j].Sequence })
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"sequence", "equipment_code", "kind", "actor", "action", "detail", "occurred_at"}); err != nil {
		return err
	}
	for _, event := range ordered {
		if err := output.Write([]string{strconv.Itoa(event.Sequence), event.EquipmentCode, event.Kind, event.Actor, event.Action, event.Detail, event.OccurredAt}); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}

func Summary(records []domain.EquipmentRecord) string {
	counts := map[domain.Status]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	statuses := []domain.Status{domain.StatusActive, domain.StatusApproved, domain.StatusArchived, domain.StatusDisabled, domain.StatusDraft, domain.StatusRetired, domain.StatusReview}
	parts := []string{fmt.Sprintf("total=%d", len(records))}
	for _, status := range statuses {
		if counts[status] > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", status, counts[status]))
		}
	}
	return strings.Join(parts, ",")
}

func AuditSummary(events []domain.AuditEvent) string {
	counts := map[string]int{}
	for _, event := range events {
		counts[event.Action]++
	}
	actions := make([]string, 0, len(counts))
	for action := range counts {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	parts := make([]string, 0, len(actions))
	for _, action := range actions {
		parts = append(parts, fmt.Sprintf("%s=%d", action, counts[action]))
	}
	return strings.Join(parts, ",")
}

func RenderDashboard(records []domain.EquipmentRecord, events []domain.AuditEvent) string {
	return "inventory{" + Summary(records) + "} audit{" + AuditSummary(events) + "}"
}
