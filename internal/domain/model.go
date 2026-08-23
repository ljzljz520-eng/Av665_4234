package domain

import "strings"

type Status string

const (
	StatusDraft    Status = "draft"
	StatusReview   Status = "review"
	StatusApproved Status = "approved"
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusRetired  Status = "retired"
	StatusArchived Status = "archived"
)

type EquipmentRecord struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Type           string `json:"type"`
	Building       string `json:"building"`
	Floor          int    `json:"floor"`
	InspectionDate string `json:"inspection_date"`
	Owner          string `json:"owner"`
	Status         Status `json:"status"`
	Notes          string `json:"notes"`
	Revision       int    `json:"revision"`
}

type AuditEvent struct {
	ID            string `json:"id"`
	EquipmentCode string `json:"equipment_code"`
	Kind          string `json:"kind"`
	Actor         string `json:"actor"`
	Action        string `json:"action"`
	Detail        string `json:"detail"`
	OccurredAt    string `json:"occurred_at"`
	Sequence      int    `json:"sequence"`
}

type Attachment struct {
	ID            string `json:"id"`
	EquipmentCode string `json:"equipment_code"`
	Name          string `json:"name"`
	MediaType     string `json:"media_type"`
	Path          string `json:"path"`
	Checksum      string `json:"checksum"`
	Verified      bool   `json:"verified"`
}

type Review struct {
	ID            string `json:"id"`
	EquipmentCode string `json:"equipment_code"`
	Stage         string `json:"stage"`
	Actor         string `json:"actor"`
	Comment       string `json:"comment"`
	CreatedAt     string `json:"created_at"`
}

type SearchFilter struct {
	Code     string
	Type     string
	Building string
	Floor    int
	Owner    string
	Status   Status
}

type ImportRow struct {
	Code           string
	Type           string
	Building       string
	Floor          int
	InspectionDate string
	Owner          string
	Status         Status
	Notes          string
}

type ImportIssue struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ImportOutcome struct {
	Accepted []EquipmentRecord `json:"accepted"`
	Issues   []ImportIssue     `json:"issues"`
}

func (r EquipmentRecord) DisplayLocation() string {
	return strings.TrimSpace(r.Building) + " / " + formatFloor(r.Floor)
}

func formatFloor(floor int) string {
	if floor <= 0 {
		return "未标注楼层"
	}
	return "F" + integerText(floor)
}

func integerText(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 12)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
