package inspection

import (
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestInspectionAssessment(t *testing.T) {
	record := domain.EquipmentRecord{Code: "V-122", InspectionDate: "2024-01-01"}
	assessment := Assess(record, "2026-08-23")
	if assessment.Grade != GradeLate || !assessment.FollowUp {
		t.Fatalf("unexpected assessment: %+v", assessment)
	}
	if _, _, _, err := ParseDate("bad"); err == nil {
		t.Fatal("invalid date should fail")
	}
}

func TestInspectionPlan(t *testing.T) {
	records := []domain.EquipmentRecord{{Code: "V-2", Building: "B", Floor: 2, Owner: "乙", InspectionDate: "2025-01-01"}, {Code: "V-1", Building: "A", Floor: 1, Owner: "甲", InspectionDate: "2026-08-01"}}
	plan := BuildPlan(records, "2026-08-23")
	if len(plan) != 1 || plan[0].Code != "V-2" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if PlanSummary(plan) != "late=0,due=1,invalid=0,total=1" {
		t.Fatal("unexpected plan summary")
	}
}
