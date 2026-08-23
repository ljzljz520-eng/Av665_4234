package registry

import (
	"errors"
	"sort"
	"strings"

	"fire-equipment-control/internal/domain"
)

type TypeDescriptor struct {
	Name           string
	Category       string
	Portable       bool
	RequiresPhoto  bool
	InspectionDays int
}

var catalog = map[string]TypeDescriptor{
	"灭火器": {Name: "灭火器", Category: "灭火设备", Portable: true, RequiresPhoto: true, InspectionDays: 365},
	"消火栓": {Name: "消火栓", Category: "供水设备", Portable: false, RequiresPhoto: true, InspectionDays: 365},
	"防火门": {Name: "防火门", Category: "分隔设备", Portable: false, RequiresPhoto: true, InspectionDays: 180},
	"应急灯": {Name: "应急灯", Category: "疏散设备", Portable: false, RequiresPhoto: false, InspectionDays: 180},
	"喷淋头": {Name: "喷淋头", Category: "供水设备", Portable: false, RequiresPhoto: false, InspectionDays: 365},
}

func NormalizeType(value string) string {
	trimmed := strings.TrimSpace(value)
	for name := range catalog {
		if strings.EqualFold(name, trimmed) {
			return name
		}
	}
	return trimmed
}

func Lookup(value string) (TypeDescriptor, bool) {
	descriptor, ok := catalog[NormalizeType(value)]
	return descriptor, ok
}

func ValidateType(value string) error {
	if _, ok := Lookup(value); !ok {
		return errors.New("equipment type is not in the catalog")
	}
	return nil
}

func Types() []TypeDescriptor {
	items := make([]TypeDescriptor, 0, len(catalog))
	for _, descriptor := range catalog {
		items = append(items, descriptor)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func TypeNames() []string {
	items := Types()
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func IsPortable(value string) bool {
	descriptor, ok := Lookup(value)
	return ok && descriptor.Portable
}

func RequiresPhoto(value string) bool {
	descriptor, ok := Lookup(value)
	return ok && descriptor.RequiresPhoto
}

func InspectionDays(value string) int {
	descriptor, ok := Lookup(value)
	if !ok {
		return 0
	}
	return descriptor.InspectionDays
}

func Category(value string) string {
	descriptor, ok := Lookup(value)
	if !ok {
		return "未分类"
	}
	return descriptor.Category
}

func CatalogSummary() map[string]int {
	counts := map[string]int{}
	for _, descriptor := range catalog {
		counts[descriptor.Category]++
	}
	return counts
}

func Classify(record domain.EquipmentRecord) string {
	descriptor, ok := Lookup(record.Type)
	if !ok {
		return "unknown"
	}
	if descriptor.Portable {
		return "portable-" + descriptor.Category
	}
	return "fixed-" + descriptor.Category
}

func PhotoRequirement(record domain.EquipmentRecord) string {
	if RequiresPhoto(record.Type) {
		return "照片必填"
	}
	return "照片可选"
}

func FilterByCategory(records []domain.EquipmentRecord, category string) []domain.EquipmentRecord {
	filtered := make([]domain.EquipmentRecord, 0, len(records))
	for _, record := range records {
		if strings.EqualFold(Category(record.Type), strings.TrimSpace(category)) {
			filtered = append(filtered, domain.CloneRecord(record))
		}
	}
	return filtered
}

func ValidateRecord(record domain.EquipmentRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	return ValidateType(record.Type)
}
