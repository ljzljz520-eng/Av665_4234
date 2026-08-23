package domain

import "testing"

func TestEquipmentValidation(t *testing.T) {
	record := EquipmentRecord{Code: "V-122", Type: "灭火器", Building: "A栋", Floor: 2, InspectionDate: "2026-08-23", Owner: "张三", Status: StatusDraft}
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	if !ValidateDateShape(record.InspectionDate) {
		t.Fatal("expected valid date shape")
	}
	if StatusLabel(StatusActive) != "在用" {
		t.Fatal("unexpected status label")
	}
}

func TestStatusTransitions(t *testing.T) {
	record := EquipmentRecord{Code: "V-122", Status: StatusDraft}
	for _, status := range []Status{StatusReview, StatusApproved, StatusActive, StatusArchived} {
		var err error
		record, err = Transition(record, status)
		if err != nil {
			t.Fatalf("transition to %s: %v", status, err)
		}
	}
	if !IsTerminal(record.Status) {
		t.Fatal("archived record should be terminal")
	}
}
