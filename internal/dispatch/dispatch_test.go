package dispatch

import "testing"

func TestDispatchBatch(t *testing.T) {
	batch := NewBatch("V-122", []string{"甲", "乙"})
	if batch.Empty() || batch.Len() != 2 || WorkerCount(batch) != 2 {
		t.Fatal("batch metadata is wrong")
	}
	if !batch.ContainsActor("甲") || len(batch.JobIDs()) != 2 || len(batch.DuplicateActors()) != 0 {
		t.Fatal("batch jobs are wrong")
	}
}
