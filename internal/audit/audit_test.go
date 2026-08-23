package audit

import (
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestAuditLifecycle(t *testing.T) {
	events := []domain.AuditEvent{NewLifecycleEvent("V-122", "甲", "retire", "到期", 2), NewChangeEvent("V-122", "乙", "甲", "丙", 1)}
	timeline := Timeline(events)
	if timeline[0].Action != "owner-change" {
		t.Fatal("timeline did not sort by sequence")
	}
	if !IsLifecycleEvent(timeline[1]) {
		t.Fatal("retire event should be lifecycle")
	}
	if CountActions(events)["retire"] != 1 {
		t.Fatal("action count missing")
	}
}

func TestAuditDescriptionAndFilter(t *testing.T) {
	event := NewLifecycleEvent("V-8", "甲", "disable", "漏检", 4)
	if Describe(event) == "" {
		t.Fatal("description should not be empty")
	}
	if len(FilterByAction([]domain.AuditEvent{event}, "retire")) != 0 {
		t.Fatal("filter returned an unrelated event")
	}
}
