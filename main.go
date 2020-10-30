package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dsbezerra/cinemais-jobs/fcm"

	"github.com/dsbezerra/cinemais"
	"github.com/dsbezerra/cinemais-jobs/job"
)

func main() {
	id := flag.String("id", "", "id of target theater to a specific job")
	jobname := flag.String("job", "", "which job to run")
	workers := flag.Int("workers", 5, "how many workers to allocate")
	all := flag.Bool("alltheaters", false, "whether the job should be executed for all theaters or not")
	notify := flag.Bool("notify", false, "whether the notification should be sent or not")
	fcmAuthKey := flag.String("fcmauthkey", fcm.FCMAuthKeyPlaceholder, "FCM authentication key used to send notifications")
	date := flag.String("date", time.Now().Format("2006-01-02"), "which date to check (YYYY-MM-DD)")

	flag.Parse()

	if *jobname == "" {
		log.Fatal("missing job")
	}

	if !job.IsJobValid(*jobname) {
		job.PrintJobs()
		log.Fatalf("job '%s' is not valid", *jobname)
	}

	if *notify && (*fcmAuthKey == "" || *fcmAuthKey == fcm.FCMAuthKeyPlaceholder) {
		log.Fatal("notify was set, but FCM auth key is missing")
	}
	os.Setenv(fcm.FCMAuthKey, *fcmAuthKey)

	d, err := time.Parse("2006-01-02", *date)
	if err != nil {
		log.Fatal(fmt.Sprintf("invalid date format. expected YYYY-MM-DD, but got %s", *date))
	}
	input := job.JobInput{
		Date:   d,
		Notify: *notify,
	}

	if *all {
		theaters, err := cinemais.GetTheaters()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for _, theater := range theaters {
				sval := strconv.Itoa(theater.ID)
				j := job.NewJob(sval, *jobname)
				jobs <- j
			}
			close(jobs)
		}()

		done := make(chan bool)
		go result(done)

		createWorkerPool(*workers, input)

		<-done

	} else {
		if *id == "" {
			log.Fatal("missing target theater")
		}

		j := job.NewJob(*id, *jobname)
		if j == nil {
			log.Fatal("couldn't find job to run")
		}

		j.Run(input)
	}
}
