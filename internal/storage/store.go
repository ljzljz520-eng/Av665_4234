package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"fire-equipment-control/internal/domain"
	"go.etcd.io/bbolt"
)

var (
	ErrNotFound      = errors.New("storage record not found")
	bucketEquipment  = []byte("equipment")
	bucketAudit      = []byte("audit_events")
	bucketAttachment = []byte("attachments")
	bucketReview     = []byte("reviews")
	bucketMeta       = []byte("meta")
)

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}
	database, err := bbolt.Open(filepath.Clean(path), 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}
	store := &Store{db: database, path: filepath.Clean(path)}
	if err := store.initialize(); err != nil {
		database.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{bucketEquipment, bucketAudit, bucketAttachment, bucketReview, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		if tx.Bucket(bucketMeta).Get([]byte("sequence")) == nil {
			return tx.Bucket(bucketMeta).Put([]byte("sequence"), []byte("0"))
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode storage value: %w", err)
	}
	return data, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode storage value: %w", err)
	}
	return nil
}

func putJSON(tx *bbolt.Tx, bucket, key []byte, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	return tx.Bucket(bucket).Put(key, data)
}

func getJSON(tx *bbolt.Tx, bucket, key []byte, target any) error {
	value := tx.Bucket(bucket).Get(key)
	if value == nil {
		return ErrNotFound
	}
	return decode(value, target)
}

func listJSON(tx *bbolt.Tx, bucket []byte, factory func() any) ([]any, error) {
	items := make([]any, 0)
	return items, tx.Bucket(bucket).ForEach(func(_, value []byte) error {
		if value == nil {
			return nil
		}
		item := factory()
		if err := decode(value, item); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
}

func (s *Store) NextSequence() (int, error) {
	sequence := 0
	err := s.db.Update(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketMeta).Get([]byte("sequence"))
		if _, err := fmt.Sscanf(string(value), "%d", &sequence); err != nil {
			return err
		}
		sequence++
		return tx.Bucket(bucketMeta).Put([]byte("sequence"), []byte(fmt.Sprintf("%d", sequence)))
	})
	return sequence, err
}

func (s *Store) Snapshot() (map[string]int, error) {
	snapshot := map[string]int{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{bucketEquipment, bucketAudit, bucketAttachment, bucketReview} {
			count := 0
			err := tx.Bucket(bucket).ForEach(func(_, value []byte) error {
				if value != nil {
					count++
				}
				return nil
			})
			if err != nil {
				return err
			}
			snapshot[string(bucket)] = count
		}
		return nil
	})
	return snapshot, err
}

func (s *Store) GetEquipmentInTx(tx *bbolt.Tx, code string) (domain.EquipmentRecord, error) {
	var record domain.EquipmentRecord
	err := getJSON(tx, bucketEquipment, domain.RecordKey(code), &record)
	return record, err
}
