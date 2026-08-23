package storage

import (
	"fmt"

	"fire-equipment-control/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveReview(review domain.Review) error {
	if err := domain.ValidateReview(review); err != nil {
		return err
	}
	if review.ID == "" {
		return fmt.Errorf("review id is required")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketReview, []byte(review.ID), review)
	})
}

func (s *Store) GetReview(id string) (domain.Review, error) {
	var review domain.Review
	err := s.db.View(func(tx *bbolt.Tx) error {
		return getJSON(tx, bucketReview, []byte(id), &review)
	})
	return review, err
}

func (s *Store) ListReviews(code string) ([]domain.Review, error) {
	items := []domain.Review{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketReview).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var item domain.Review
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

func (s *Store) CountReviews(code string) (int, error) {
	items, err := s.ListReviews(code)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
