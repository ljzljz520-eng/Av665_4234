package storage

import (
	"path/filepath"
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "firecontrol.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := domain.EquipmentRecord{ID: "equipment-V-122", Code: "V-122", Type: "灭火器", Building: "A栋", Floor: 2, InspectionDate: "2026-08-23", Owner: "张三", Status: domain.StatusActive}
	if err := store.SaveEquipment(record); err != nil {
		t.Fatal(err)
	}
	event := domain.AuditEvent{ID: "audit-0001", EquipmentCode: "V-122", Kind: "lifecycle", Actor: "张三", Action: "confirm", Detail: "ok", OccurredAt: "sequence-0001", Sequence: 1}
	if err := store.SaveAuditEvent(event); err != nil {
		t.Fatal(err)
	}
	attachment := domain.Attachment{ID: "photo-1", EquipmentCode: "V-122", Name: "front.jpg", MediaType: "image/jpeg", Path: "photos/front.jpg", Checksum: "checksum", Verified: true}
	if err := store.SaveAttachment(attachment); err != nil {
		t.Fatal(err)
	}
	review := domain.Review{ID: "review-1", EquipmentCode: "V-122", Stage: "approved", Actor: "李四", Comment: "通过", CreatedAt: "sequence-0002"}
	if err := store.SaveReview(review); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.GetEquipment("V-122")
	if err != nil || loaded.Code != "V-122" {
		t.Fatalf("equipment did not survive reopen: %+v %v", loaded, err)
	}
	events, err := reopened.ListAuditEvents("V-122")
	if err != nil || len(events) != 1 {
		t.Fatalf("audit did not survive reopen: %d %v", len(events), err)
	}
	photos, err := reopened.ListAttachments("V-122")
	if err != nil || len(photos) != 1 {
		t.Fatalf("attachment did not survive reopen: %d %v", len(photos), err)
	}
	reviews, err := reopened.ListReviews("V-122")
	if err != nil || len(reviews) != 1 {
		t.Fatalf("review did not survive reopen: %d %v", len(reviews), err)
	}
}

func TestStorageCountsAndDelete(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "count.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	record := domain.EquipmentRecord{ID: "equipment-A", Code: "A", Type: "栓", Building: "B", Floor: 1, InspectionDate: "2026-01-01", Owner: "甲", Status: domain.StatusDraft}
	if err := store.SaveEquipment(record); err != nil {
		t.Fatal(err)
	}
	count, err := store.CountEquipment()
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if err := store.DeleteEquipment("A"); err != nil {
		t.Fatal(err)
	}
	count, err = store.CountEquipment()
	if err != nil || count != 0 {
		t.Fatalf("count after delete=%d err=%v", count, err)
	}
}
