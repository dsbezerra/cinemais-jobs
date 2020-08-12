package job

import "fmt"

// Result is a interface of a job result
type Result interface {
	JobName() string
	// Notify sends a notification. Returns true if notification was successfully sent.
	Notify() bool
}

type Job interface {
	Run(notify bool) Result
}

// Job represents a job.
type job struct {
	ID        uint
	TheaterID string
}

type jobInfo struct {
	name, description string
}

var (
	// JobCounter is used only to get job identifiers
	JobCounter uint = 0
)

// NewJob ...
func NewJob(theater string, jobname string) Job {
	switch jobname {
	case JobWeekReleases:
		return &CheckWeekReleases{
			ID:        getJobID(),
			TheaterID: theater,
		}
	default:
		return nil
	}
}

func getJobID() uint {
	defer func() {
		JobCounter++
	}()
	return JobCounter
}

// IsJobValid check if the given job is valid.
func IsJobValid(j string) bool {
	switch j {
	case JobWeekReleases:
		return true
	default:
		return false
	}
}

// PrintJobs print information about all available jobs to run.
func PrintJobs() {
	jobInfos := []jobInfo{
		jobInfo{
			name: JobWeekReleases, description: "Check for releases in the current week",
		},
	}

	fmt.Printf("\nAvailable jobs:\n")
	for _, i := range jobInfos {
		fmt.Printf("\t%s\t %s\n", i.name, i.description)
	}
	fmt.Println()
}
