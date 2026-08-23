package storage

import (
	"fmt"

	"fire-equipment-control/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAuditEvent(event domain.AuditEvent) error {
	if event.ID == "" {
		return fmt.Errorf("audit event id is required")
	}
	if event.EquipmentCode == "" {
		return fmt.Errorf("audit equipment code is required")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketAudit, []byte(event.ID), event)
	})
}

func (s *Store) ListAuditEvents(code string) ([]domain.AuditEvent, error) {
	events := []domain.AuditEvent{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAudit).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var event domain.AuditEvent
			if err := decode(value, &event); err != nil {
				return err
			}
			if code == "" || event.EquipmentCode == code {
				events = append(events, event)
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	domain.SortEvents(events)
	return events, nil
}

func (s *Store) CountAuditEvents(code string) (int, error) {
	events, err := s.ListAuditEvents(code)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

func (s *Store) DeleteAuditEvent(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketAudit).Get([]byte(id)) == nil {
			return ErrNotFound
		}
		return tx.Bucket(bucketAudit).Delete([]byte(id))
	})
}
