package service

import (
	"path/filepath"
	"testing"

	"fire-equipment-control/internal/storage"
)

func TestBusiness022Regression(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "concurrency.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	manager := New(store)
	if err := store.SaveEquipment(PendingConfirmationRecord("V-122")); err != nil {
		t.Fatal(err)
	}
	barrier := ReadyBarrier()
	resultsDone := make(chan []ConfirmationResult, 1)
	go func() {
		resultsDone <- manager.RunBarrierConfirmations("V-122", []string{"操作员甲", "操作员乙"}, barrier)
	}()
	ReleaseBarrier(barrier)
	results := <-resultsDone
	if CountSuccessfulConfirmations(results) != 2 {
		t.Fatalf("both barrier confirmations must be recorded: %+v", results)
	}
	count, err := store.CountAuditEvents("V-122")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected two persisted confirmations, got %d", count)
	}
}
