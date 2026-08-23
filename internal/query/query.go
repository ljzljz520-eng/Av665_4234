package query

import (
	"sort"
	"strings"

	"fire-equipment-control/internal/domain"
)

func Matches(record domain.EquipmentRecord, filter domain.SearchFilter) bool {
	if filter.Code != "" && !strings.EqualFold(record.Code, filter.Code) {
		return false
	}
	if filter.Type != "" && !strings.EqualFold(record.Type, filter.Type) {
		return false
	}
	if filter.Building != "" && !strings.EqualFold(record.Building, filter.Building) {
		return false
	}
	if filter.Floor > 0 && record.Floor != filter.Floor {
		return false
	}
	if filter.Owner != "" && !strings.Contains(strings.ToLower(record.Owner), strings.ToLower(filter.Owner)) {
		return false
	}
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	return true
}

func Filter(records []domain.EquipmentRecord, filter domain.SearchFilter) []domain.EquipmentRecord {
	result := make([]domain.EquipmentRecord, 0, len(records))
	for _, record := range records {
		if Matches(record, filter) {
			result = append(result, domain.CloneRecord(record))
		}
	}
	return result
}

func SortByCode(records []domain.EquipmentRecord) []domain.EquipmentRecord {
	result := append([]domain.EquipmentRecord(nil), records...)
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})
	return result
}

func SortByInspection(records []domain.EquipmentRecord) []domain.EquipmentRecord {
	result := append([]domain.EquipmentRecord(nil), records...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].InspectionDate != result[j].InspectionDate {
			return result[i].InspectionDate < result[j].InspectionDate
		}
		return result[i].Code < result[j].Code
	})
	return result
}

func Paginate(records []domain.EquipmentRecord, offset, limit int) []domain.EquipmentRecord {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || offset >= len(records) {
		return []domain.EquipmentRecord{}
	}
	end := offset + limit
	if end > len(records) {
		end = len(records)
	}
	return append([]domain.EquipmentRecord(nil), records[offset:end]...)
}

func StatusCounts(records []domain.EquipmentRecord) map[domain.Status]int {
	counts := map[domain.Status]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	return counts
}

func LocationCounts(records []domain.EquipmentRecord) map[string]int {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.DisplayLocation()]++
	}
	return counts
}

func BuildFilter(code, typ, building, owner string, floor int, status domain.Status) domain.SearchFilter {
	return domain.SearchFilter{Code: strings.TrimSpace(code), Type: strings.TrimSpace(typ), Building: strings.TrimSpace(building), Owner: strings.TrimSpace(owner), Floor: floor, Status: status}
}
