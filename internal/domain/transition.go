package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrInvalidTransition = errors.New("invalid equipment status transition")

func CanTransition(from, to Status) bool {
	switch from {
	case StatusDraft:
		return to == StatusReview
	case StatusReview:
		return to == StatusApproved || to == StatusDraft
	case StatusApproved:
		return to == StatusActive
	case StatusActive:
		return to == StatusDisabled || to == StatusRetired || to == StatusArchived
	case StatusDisabled:
		return to == StatusActive || to == StatusRetired
	case StatusRetired:
		return to == StatusArchived
	case StatusArchived:
		return false
	default:
		return false
	}
}

func Transition(record EquipmentRecord, to Status) (EquipmentRecord, error) {
	if !CanTransition(record.Status, to) {
		return EquipmentRecord{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, record.Status, to)
	}
	record.Status = to
	record.Revision++
	return record, nil
}

func RecordKey(code string) []byte {
	return []byte(strings.TrimSpace(code))
}

func CloneRecord(record EquipmentRecord) EquipmentRecord {
	return EquipmentRecord{
		ID: record.ID, Code: record.Code, Type: record.Type, Building: record.Building,
		Floor: record.Floor, InspectionDate: record.InspectionDate, Owner: record.Owner,
		Status: record.Status, Notes: record.Notes, Revision: record.Revision,
	}
}

func SortRecords(records []EquipmentRecord) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].Building != records[j].Building {
			return records[i].Building < records[j].Building
		}
		if records[i].Floor != records[j].Floor {
			return records[i].Floor < records[j].Floor
		}
		return records[i].Code < records[j].Code
	})
}

func SortEvents(events []AuditEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].EquipmentCode != events[j].EquipmentCode {
			return events[i].EquipmentCode < events[j].EquipmentCode
		}
		return events[i].Sequence < events[j].Sequence
	})
}

func IsLifecycleAction(action string) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "disable", "retire", "archive":
		return true
	default:
		return false
	}
}
