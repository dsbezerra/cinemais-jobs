package main

import (
	"fmt"
	"sync"

	"github.com/dsbezerra/cinemais-jobs/job"
)

var (
	jobs    = make(chan job.Job, 10)
	results = make(chan job.Result, 10)
)

func createWorkerPool(noOfWorkers int, input job.JobInput) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg, input)
	}
	wg.Wait()
	close(results)
}

func worker(wg *sync.WaitGroup, input job.JobInput) {
	for job := range jobs {
		output := job.Run(input)
		results <- output
	}
	wg.Done()
}

func result(done chan bool) {
	for result := range results {
		if result == nil {
			continue
		}
		fmt.Printf("Job #%s - Finished.\n", result.JobName())
	}
	done <- true
}
