package storage

import (
	"fmt"

	"fire-equipment-control/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAttachment(attachment domain.Attachment) error {
	if err := domain.ValidateAttachment(attachment); err != nil {
		return err
	}
	if attachment.ID == "" {
		return fmt.Errorf("attachment id is required")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketAttachment, []byte(attachment.ID), attachment)
	})
}

func (s *Store) GetAttachment(id string) (domain.Attachment, error) {
	var attachment domain.Attachment
	err := s.db.View(func(tx *bbolt.Tx) error {
		return getJSON(tx, bucketAttachment, []byte(id), &attachment)
	})
	return attachment, err
}

func (s *Store) ListAttachments(code string) ([]domain.Attachment, error) {
	items := []domain.Attachment{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAttachment).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var item domain.Attachment
			if err := decode(value, &item); err != nil {
				return err
			}
			if code == "" || item.EquipmentCode == code {
				items = append(items, item)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) CountAttachments(code string) (int, error) {
	items, err := s.ListAttachments(code)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
