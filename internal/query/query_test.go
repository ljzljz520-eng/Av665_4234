package query

import (
	"testing"

	"fire-equipment-control/internal/domain"
)

func TestSearchFilters(t *testing.T) {
	records := []domain.EquipmentRecord{
		{Code: "V-2", Type: "灭火器", Building: "B栋", Floor: 2, Owner: "张三", Status: domain.StatusActive},
		{Code: "V-1", Type: "消火栓", Building: "A栋", Floor: 1, Owner: "李四", Status: domain.StatusDraft},
	}
	filter := BuildFilter("", "灭火器", "B栋", "张", 2, domain.StatusActive)
	matched := Filter(records, filter)
	if len(matched) != 1 || matched[0].Code != "V-2" {
		t.Fatalf("unexpected search result: %+v", matched)
	}
	if len(Paginate(SortByCode(records), 0, 1)) != 1 {
		t.Fatal("pagination failed")
	}
}

func TestQuerySummaries(t *testing.T) {
	records := []domain.EquipmentRecord{{Code: "V-1", Building: "A", Floor: 1, Status: domain.StatusActive}, {Code: "V-2", Building: "A", Floor: 1, Status: domain.StatusDraft}}
	if StatusCounts(records)[domain.StatusActive] != 1 || LocationCounts(records)["A / F1"] != 2 {
		t.Fatal("summary counts failed")
	}
	if len(Paginate(records, 9, 2)) != 0 {
		t.Fatal("out-of-range pagination should be empty")
	}
}
