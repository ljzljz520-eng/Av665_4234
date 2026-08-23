package dashboard

import (
	"strings"
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestDashboardBuildAndRender(t *testing.T) {
	records := []domain.EquipmentRecord{{Code: "V-122", Building: "A栋", Floor: 2, Owner: "甲", Status: domain.StatusActive}, {Code: "V-123", Building: "A栋", Floor: 2, Owner: "乙", Status: domain.StatusDraft}}
	events := []domain.AuditEvent{{ID: "audit-1", EquipmentCode: "V-122", Actor: "甲", Action: "confirm", Sequence: 1}}
	view := Build(records, events)
	if err := ValidateView(view); err != nil {
		t.Fatal(err)
	}
	if view.Total != 2 || StatusCount(view, domain.StatusActive) != 1 || LocationCount(view) != 1 || !HasAuditAction(view, "confirm") {
		t.Fatal("dashboard counts are wrong")
	}
	if !strings.Contains(Render(view), "消防器材管理台账") || len(OwnerLeaderboard(view)) != 2 {
		t.Fatal("dashboard render is incomplete")
	}
	if ActionCount(view, "confirm") != 1 || len(StatusLegend()) != 7 || len(SummaryRows(view)) < 3 {
		t.Fatal("dashboard summary is incomplete")
	}
}

func TestDashboardEmptyAndLocation(t *testing.T) {
	view := EmptyView()
	if !IsEmpty(view) || len(EquipmentCodes(view)) != 0 {
		t.Fatal("empty dashboard should be empty")
	}
	if _, ok := FindLocation(view, "A栋", 1); ok {
		t.Fatal("empty dashboard has a location")
	}
}
