package attachment

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"

	"fire-equipment-control/internal/domain"
)

func BuildPhoto(code, id, name, path, mediaType, content string) (domain.Attachment, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(id) == "" {
		return domain.Attachment{}, errors.New("photo code and id are required")
	}
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || cleanPath == ".." {
		return domain.Attachment{}, errors.New("photo path is invalid")
	}
	checksum := Checksum(content)
	item := domain.Attachment{
		ID: id, EquipmentCode: strings.TrimSpace(code), Name: strings.TrimSpace(name),
		MediaType: strings.ToLower(strings.TrimSpace(mediaType)), Path: cleanPath,
		Checksum: checksum, Verified: false,
	}
	if err := domain.ValidateAttachment(item); err != nil {
		return domain.Attachment{}, err
	}
	return item, nil
}

func Checksum(content string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:])
}

func Verify(item domain.Attachment, content string) bool {
	if item.Checksum == "" || item.MediaType == "" {
		return false
	}
	return strings.EqualFold(item.Checksum, Checksum(content)) && strings.HasPrefix(strings.ToLower(item.MediaType), "image/")
}

func MarkVerified(item domain.Attachment, content string) (domain.Attachment, error) {
	if !Verify(item, content) {
		return domain.Attachment{}, errors.New("photo verification failed")
	}
	item.Verified = true
	return item, nil
}

func IsPhotoMedia(mediaType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(mediaType)), "image/")
}

func ExtensionAllowed(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func GroupByEquipment(items []domain.Attachment) map[string][]domain.Attachment {
	grouped := map[string][]domain.Attachment{}
	for _, item := range items {
		grouped[item.EquipmentCode] = append(grouped[item.EquipmentCode], item)
	}
	return grouped
}
