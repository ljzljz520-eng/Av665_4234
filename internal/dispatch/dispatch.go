package dispatch

import (
	"fmt"
	"sort"
)

type Job struct {
	ID            string
	EquipmentCode string
	Actor         string
	Ordinal       int
}

type Batch struct {
	code string
	jobs []Job
}

func NewBatch(code string, actors []string) Batch {
	jobs := make([]Job, 0, len(actors))
	for index, actor := range actors {
		jobs = append(jobs, Job{ID: fmt.Sprintf("pending-%s-%02d", code, index+1), EquipmentCode: code, Actor: actor, Ordinal: index + 1})
	}
	return Batch{code: code, jobs: jobs}
}

func (b Batch) Code() string {
	return b.code
}

func (b Batch) Jobs() []Job {
	return append([]Job(nil), b.jobs...)
}

func (b Batch) Len() int {
	return len(b.jobs)
}

func (b Batch) Empty() bool {
	return len(b.jobs) == 0
}

func (b Batch) JobIDs() []string {
	ids := make([]string, 0, len(b.jobs))
	for _, job := range b.jobs {
		ids = append(ids, job.ID)
	}
	return ids
}

func (b Batch) Actors() []string {
	actors := make([]string, 0, len(b.jobs))
	for _, job := range b.jobs {
		actors = append(actors, job.Actor)
	}
	return actors
}

func (b Batch) SortedJobs() []Job {
	jobs := b.Jobs()
	sort.SliceStable(jobs, func(i, j int) bool { return jobs[i].Ordinal < jobs[j].Ordinal })
	return jobs
}

func (b Batch) ContainsActor(actor string) bool {
	for _, job := range b.jobs {
		if job.Actor == actor {
			return true
		}
	}
	return false
}

func (b Batch) DuplicateActors() []string {
	seen := map[string]bool{}
	duplicates := []string{}
	for _, job := range b.jobs {
		if seen[job.Actor] {
			duplicates = append(duplicates, job.Actor)
		}
		seen[job.Actor] = true
	}
	return duplicates
}

func WorkerCount(batch Batch) int {
	if batch.Len() <= 1 {
		return 1
	}
	if batch.Len() > 8 {
		return 8
	}
	return batch.Len()
}
