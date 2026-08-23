package audit

import (
	"fmt"
	"sort"
	"strings"

	"fire-equipment-control/internal/domain"
)

func NewLifecycleEvent(code, actor, action, detail string, sequence int) domain.AuditEvent {
	return domain.AuditEvent{
		ID:            fmt.Sprintf("audit-%04d", sequence),
		EquipmentCode: strings.TrimSpace(code),
		Kind:          "lifecycle",
		Actor:         strings.TrimSpace(actor),
		Action:        strings.TrimSpace(action),
		Detail:        strings.TrimSpace(detail),
		OccurredAt:    fmt.Sprintf("sequence-%04d", sequence),
		Sequence:      sequence,
	}
}

func NewChangeEvent(code, actor, before, after string, sequence int) domain.AuditEvent {
	detail := fmt.Sprintf("owner:%s=>%s", strings.TrimSpace(before), strings.TrimSpace(after))
	return domain.AuditEvent{
		ID:            fmt.Sprintf("audit-%04d", sequence),
		EquipmentCode: strings.TrimSpace(code),
		Kind:          "change",
		Actor:         strings.TrimSpace(actor),
		Action:        "owner-change",
		Detail:        detail,
		OccurredAt:    fmt.Sprintf("sequence-%04d", sequence),
		Sequence:      sequence,
	}
}

func NewReviewEvent(code, actor, stage, detail string, sequence int) domain.AuditEvent {
	return domain.AuditEvent{
		ID:            fmt.Sprintf("audit-%04d", sequence),
		EquipmentCode: strings.TrimSpace(code),
		Kind:          "review",
		Actor:         strings.TrimSpace(actor),
		Action:        "review-" + strings.TrimSpace(stage),
		Detail:        strings.TrimSpace(detail),
		OccurredAt:    fmt.Sprintf("sequence-%04d", sequence),
		Sequence:      sequence,
	}
}

func FilterByAction(events []domain.AuditEvent, action string) []domain.AuditEvent {
	filtered := make([]domain.AuditEvent, 0, len(events))
	for _, event := range events {
		if action == "" || event.Action == action {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func CountActions(events []domain.AuditEvent) map[string]int {
	counts := map[string]int{}
	for _, event := range events {
		counts[event.Action]++
	}
	return counts
}

func Timeline(events []domain.AuditEvent) []domain.AuditEvent {
	ordered := append([]domain.AuditEvent(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Sequence != ordered[j].Sequence {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

func Describe(event domain.AuditEvent) string {
	parts := []string{event.OccurredAt, event.Actor, event.Action}
	if event.Detail != "" {
		parts = append(parts, event.Detail)
	}
	return strings.Join(parts, " | ")
}

func IsLifecycleEvent(event domain.AuditEvent) bool {
	return event.Kind == "lifecycle" || domain.IsLifecycleAction(event.Action)
}
