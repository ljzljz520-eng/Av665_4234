package storage

import (
	"fmt"

	"fire-equipment-control/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveEquipment(record domain.EquipmentRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketEquipment, domain.RecordKey(record.Code), record)
	})
}

func (s *Store) GetEquipment(code string) (domain.EquipmentRecord, error) {
	var record domain.EquipmentRecord
	err := s.db.View(func(tx *bbolt.Tx) error {
		var lookupErr error
		record, lookupErr = s.GetEquipmentInTx(tx, code)
		return lookupErr
	})
	if err != nil {
		return record, err
	}
	return record, nil
}

func (s *Store) ListEquipment() ([]domain.EquipmentRecord, error) {
	items := []domain.EquipmentRecord{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketEquipment).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var record domain.EquipmentRecord
			if err := decode(value, &record); err != nil {
				return err
			}
			items = append(items, record)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("list equipment: %w", err)
	}
	domain.SortRecords(items)
	return items, nil
}

func (s *Store) DeleteEquipment(code string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketEquipment).Get(domain.RecordKey(code)) == nil {
			return ErrNotFound
		}
		return tx.Bucket(bucketEquipment).Delete(domain.RecordKey(code))
	})
}

func (s *Store) CountEquipment() (int, error) {
	count := 0
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketEquipment).ForEach(func(_, value []byte) error {
			if value != nil {
				count++
			}
			return nil
		})
	})
	return count, err
}
