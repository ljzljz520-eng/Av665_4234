package registry

import (
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestRegistryCatalog(t *testing.T) {
	if NormalizeType(" 灭火器 ") != "灭火器" || !RequiresPhoto("灭火器") || !IsPortable("灭火器") {
		t.Fatal("catalog normalization failed")
	}
	if err := ValidateType("未知器材"); err == nil {
		t.Fatal("unknown type should fail")
	}
	if len(TypeNames()) < 4 || len(CatalogSummary()) < 2 {
		t.Fatal("catalog is incomplete")
	}
}

func TestRegistryRecordClassification(t *testing.T) {
	record := domain.EquipmentRecord{Type: "消火栓"}
	if Classify(record) != "fixed-供水设备" || PhotoRequirement(record) != "照片必填" {
		t.Fatal("record classification failed")
	}
}
