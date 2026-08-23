package service

import (
	"sync"

	"fire-equipment-control/internal/dispatch"
	"fire-equipment-control/internal/domain"
)

type ConfirmationResult struct {
	JobID string
	Actor string
	Error error
}

func (s *Service) ConfirmAfterBarrier(code, actor string, barrier <-chan struct{}) error {
	<-barrier
	s.mu.Lock()
	if s.consumed[code] > 0 {
		s.mu.Unlock()
		return ErrDuplicateConsumption
	}
	s.consumed[code]++
	s.mu.Unlock()
	if _, err := s.GetEquipment(code); err != nil {
		return err
	}
	return s.writeLifecycle(code, actor, "confirm-pending", "barrier confirmation consumed")
}

func (s *Service) RunBarrierConfirmations(code string, actors []string, barrier <-chan struct{}) []ConfirmationResult {
	batch := dispatch.NewBatch(code, actors)
	jobs := batch.SortedJobs()
	results := make([]ConfirmationResult, len(jobs))
	var wait sync.WaitGroup
	for index, job := range jobs {
		wait.Add(1)
		go func(index int, job dispatch.Job) {
			defer wait.Done()
			results[index] = ConfirmationResult{JobID: job.ID, Actor: job.Actor, Error: s.ConfirmAfterBarrier(job.EquipmentCode, job.Actor, barrier)}
		}(index, job)
	}
	wait.Wait()
	return results
}

func CountSuccessfulConfirmations(results []ConfirmationResult) int {
	count := 0
	for _, result := range results {
		if result.Error == nil {
			count++
		}
	}
	return count
}

func ReadyBarrier() chan struct{} {
	return make(chan struct{})
}

func ReleaseBarrier(barrier chan struct{}) {
	close(barrier)
}

func PendingConfirmationRecord(code string) domain.EquipmentRecord {
	return domain.EquipmentRecord{ID: "equipment-" + code, Code: code, Type: "灭火器", Building: "A", Floor: 1, InspectionDate: "2026-08-23", Owner: "安全员", Status: domain.StatusActive}
}
