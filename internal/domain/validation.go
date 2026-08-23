package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCodeRequired       = errors.New("equipment code is required")
	ErrTypeRequired       = errors.New("equipment type is required")
	ErrBuildingRequired   = errors.New("building is required")
	ErrOwnerRequired      = errors.New("owner is required")
	ErrInspectionRequired = errors.New("inspection date is required")
	ErrInvalidFloor       = errors.New("floor must be positive")
	ErrInvalidStatus      = errors.New("invalid equipment status")
)

func (r EquipmentRecord) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return ErrCodeRequired
	}
	if strings.TrimSpace(r.Type) == "" {
		return ErrTypeRequired
	}
	if strings.TrimSpace(r.Building) == "" {
		return ErrBuildingRequired
	}
	if r.Floor <= 0 {
		return ErrInvalidFloor
	}
	if strings.TrimSpace(r.InspectionDate) == "" {
		return ErrInspectionRequired
	}
	if strings.TrimSpace(r.Owner) == "" {
		return ErrOwnerRequired
	}
	if !ValidStatus(r.Status) {
		return fmt.Errorf("%w: %q", ErrInvalidStatus, r.Status)
	}
	return nil
}

func (r EquipmentRecord) Normalized() EquipmentRecord {
	r.Code = strings.TrimSpace(r.Code)
	r.Type = strings.TrimSpace(r.Type)
	r.Building = strings.TrimSpace(r.Building)
	r.InspectionDate = strings.TrimSpace(r.InspectionDate)
	r.Owner = strings.TrimSpace(r.Owner)
	r.Notes = strings.TrimSpace(r.Notes)
	if r.Status == "" {
		r.Status = StatusDraft
	}
	return r
}

func ValidStatus(status Status) bool {
	switch status {
	case StatusDraft, StatusReview, StatusApproved, StatusActive, StatusDisabled, StatusRetired, StatusArchived:
		return true
	default:
		return false
	}
}

func IsTerminal(status Status) bool {
	return status == StatusRetired || status == StatusArchived
}

func StatusLabel(status Status) string {
	labels := map[Status]string{
		StatusDraft:    "草稿",
		StatusReview:   "待审核",
		StatusApproved: "已批准",
		StatusActive:   "在用",
		StatusDisabled: "停用",
		StatusRetired:  "报废",
		StatusArchived: "已归档",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return "未知状态"
}

func ValidateDateShape(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	return len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
}

func ValidateAttachment(attachment Attachment) error {
	if strings.TrimSpace(attachment.EquipmentCode) == "" {
		return ErrCodeRequired
	}
	if strings.TrimSpace(attachment.Name) == "" || strings.TrimSpace(attachment.Path) == "" {
		return errors.New("attachment name and path are required")
	}
	if !strings.HasPrefix(strings.ToLower(attachment.MediaType), "image/") {
		return errors.New("attachment must be an image")
	}
	if strings.TrimSpace(attachment.Checksum) == "" {
		return errors.New("attachment checksum is required")
	}
	return nil
}

func ValidateReview(review Review) error {
	if strings.TrimSpace(review.EquipmentCode) == "" || strings.TrimSpace(review.Actor) == "" {
		return errors.New("review equipment and actor are required")
	}
	if review.Stage != "submitted" && review.Stage != "approved" && review.Stage != "rejected" {
		return errors.New("review stage is invalid")
	}
	return nil
}
