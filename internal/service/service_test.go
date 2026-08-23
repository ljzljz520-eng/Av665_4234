package service

import (
	"path/filepath"
	"strings"
	"testing"

	"fire-equipment-control/internal/domain"
	"fire-equipment-control/internal/storage"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	store, err := storage.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return New(store)
}

func sampleRecord(code string) domain.EquipmentRecord {
	return domain.EquipmentRecord{Code: code, Type: "灭火器", Building: "A栋", Floor: 2, InspectionDate: "2026-08-23", Owner: "张三", Status: domain.StatusDraft}
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	manager := newTestService(t)
	workflow, err := manager.CreateReviewArchive(sampleRecord("V-122"), "复核员")
	if err != nil {
		t.Fatal(err)
	}
	if workflow.Record.Status != domain.StatusArchived || workflow.ReviewCount != 2 || workflow.AuditCount != 4 {
		t.Fatalf("unexpected archive workflow: %+v", workflow)
	}
	photo, err := manager.AttachPhoto("V-122", "photo-1", "front.jpg", "photos/front.jpg", "image/jpeg", "front")
	if err != nil || photo.EquipmentCode != "V-122" {
		t.Fatalf("photo attach failed: %+v %v", photo, err)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	manager := newTestService(t)
	if _, err := manager.RegisterEquipment(sampleRecord("V-122")); err != nil {
		t.Fatal(err)
	}
	workflow, err := manager.SearchUpdatePublish(domain.SearchFilter{Building: "A栋"}, "V-122", "协作员", "李四")
	if err != nil {
		t.Fatal(err)
	}
	if workflow.Record.Owner != "李四" || !strings.Contains(workflow.Summary, "owner-change") {
		t.Fatalf("unexpected collaboration workflow: %+v", workflow)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	manager := newTestService(t)
	input := "code,type,building,floor,inspection_date,owner,status,notes\nV-200,灭火器,B栋,3,2026-08-23,王五,draft,新登记\nV-201,消火栓,B栋,1,2026-08-24,赵六,draft,\n"
	workflow, err := manager.ImportReport(strings.NewReader(input))
	if err != nil || len(workflow.Outcome.Accepted) != 2 || !strings.Contains(workflow.Report, "accepted") {
		t.Fatalf("unexpected import workflow: %+v %v", workflow, err)
	}
	records, err := manager.SearchEquipment(domain.SearchFilter{Building: "B栋"})
	if err != nil || len(records) != 2 {
		t.Fatalf("imported records missing: %d %v", len(records), err)
	}
}

func TestLifecycleAuditAndExport(t *testing.T) {
	manager := newTestService(t)
	if _, err := manager.RegisterEquipment(sampleRecord("V-300")); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SubmitForReview("V-300", "甲", "资料齐全"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ApproveReview("V-300", "乙", "通过"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ConfirmEquipment("V-300", "乙"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.DeactivateAndRetire("V-300", "丙"); err != nil {
		t.Fatal(err)
	}
	auditCSV, err := manager.ExportAudit("V-300")
	if err != nil || !strings.Contains(auditCSV, "disable") || !strings.Contains(auditCSV, "retire") {
		t.Fatalf("audit export missing lifecycle events: %q %v", auditCSV, err)
	}
	inventoryCSV, err := manager.ExportInventory()
	if err != nil || !strings.Contains(inventoryCSV, "V-300") {
		t.Fatalf("inventory export missing record: %q %v", inventoryCSV, err)
	}
}
