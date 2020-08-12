package job

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dsbezerra/cinemais"
	"github.com/dsbezerra/cinemais-jobs/fcm"
)

const (
	// JobWeekReleases ...
	JobWeekReleases = "week_releases"

	// TypeReleases ...
	TypeReleases = "releases"
	// TypeRelease ...
	TypeRelease = "release"
)

// CheckWeekReleases ...
type CheckWeekReleases struct {
	ID        uint
	TheaterID string
}

// CheckWeekReleasesResult ...
type CheckWeekReleasesResult struct {
	Job          *CheckWeekReleases
	Theater      cinemais.Theater
	WeekReleases []cinemais.Movie
}

// Date ...
type Date struct {
	Day   int
	Month time.Month
	Year  int
}

// Today always store the current day.
var Today = today()

var collecting bool

var theaters = make(map[int]cinemais.Theater, 0)
var movies = make(map[int]cinemais.Movie, 0)

// Run checks for week release for the given theater
func (j *CheckWeekReleases) Run(notify bool) Result {
	if j.TheaterID == "" {
		return nil
	}

	fmt.Printf("Job #%d - Running %s for theater %s...\n", j.ID, JobWeekReleases, j.TheaterID)

	if !collecting {
		if len(movies) == 0 && len(theaters) == 0 {
			collecting = true

			fmt.Printf("Job #%d - Collecting aux data...\n", j.ID)
			go collectAuxData()
		}
	}

	for collecting {
		time.Sleep(100 * time.Millisecond)
	}

	sval, _ := strconv.Atoi(j.TheaterID)
	result := &CheckWeekReleasesResult{
		Job: j,
	}

	fmt.Printf("Job #%d - Aux data collected.\n", j.ID)

	// Get current schedule
	sched, err := cinemais.GetSchedule(j.TheaterID)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	// Check which movies is releasing at the day of execution
	playing := make(map[int]bool, 0)
	for _, sess := range sched.Sessions {
		_, ok := playing[sess.Movie.ID]
		if !ok {
			m := movies[sess.Movie.ID]
			playing[sess.Movie.ID] = IsToday(m.ReleaseDate)
		}
	}

	// Add them to our result if any
	for k, release := range playing {
		if release {
			result.WeekReleases = append(result.WeekReleases, movies[k])
		}
	}

	result.Theater = theaters[sval]
	if notify {
		result.Notify()
	}

	return result
}

// JobName ...
func (r *CheckWeekReleasesResult) JobName() string {
	return fmt.Sprintf("%d (%s)", r.Job.ID, JobWeekReleases)
}

// Notify ...
func (r *CheckWeekReleasesResult) Notify() bool {
	if len(r.WeekReleases) == 0 {
		return true
	}

	type ReleasePayload struct {
		Type         string       `json:"type,omitempty"`
		Image        string       `json:"image,omitempty"`
		Text         string       `json:"text,omitempty"`
		ExpandedText string       `json:"expanded_text,omitempty"`
		MovieID      int          `json:"movie_id,omitempty"`
		Actions      []fcm.Action `json:"actions,omitempty"`
	}

	type ReleasesPayload struct {
		Type         string `json:"type,omitempty"`
		Text         string `json:"text,omitempty"`
		ExpandedText string `json:"expanded_text,omitempty"`
		BigText      string `json:"big_text,omitempty"`
		TheaterID    int    `json:"theater_id,omitempty"`
	}

	var payload interface{}

	if len(r.WeekReleases) == 1 {
		release := r.WeekReleases[0]

		getImage := func(arr []cinemais.Image, index int) (string, bool) {
			if index >= 0 && index < len(arr) {
				return arr[index].URL, true
			}
			index = len(arr) - 1
			if index < 0 {
				return "", false
			}
			return arr[index].URL, true
		}

		getPicture := func() string {
			fallback := release.PosterURLs[cinemais.PosterSizeLarge]
			if release.Images == nil || len(release.Images.Backdrops) == 0 && len(release.Images.Posters) == 0 {
				return fallback
			}
			weekday := int(time.Now().Weekday())
			backdrop, ok := getImage(release.Images.Backdrops, weekday)
			if ok {
				return backdrop
			}
			poster, ok := getImage(release.Images.Posters, weekday)
			if ok {
				return poster
			}
			return fallback
		}

		srelease := ReleasePayload{
			Type:         TypeRelease,
			Text:         r.Theater.Name,
			Image:        getPicture(),
			ExpandedText: release.Title,
			MovieID:      release.ID,
			Actions: []fcm.Action{
				fcm.NewMovieDetailsAction(release.ID),
			},
		}

		if release.Trailer != nil {
			srelease.Actions = append(srelease.Actions, fcm.NewViewTrailerAction(release.Trailer.ID))
		}

		payload = srelease

	} else {
		getBigText := func() string {
			b := strings.Builder{}
			for index, wr := range r.WeekReleases {
				b.WriteString(wr.Title)
				if index < len(r.WeekReleases)-1 {
					b.WriteString("\n")
				}
			}
			return b.String()
		}

		payload = ReleasesPayload{
			Type:         TypeReleases,
			TheaterID:    r.Theater.ID,
			Text:         r.Theater.Name,
			ExpandedText: r.Theater.Name,
			BigText:      getBigText(),
		}
	}

	notification := fcm.NewNotification(fcm.NewTheaterTopic(r.Theater.ID), payload)
	ok, err := fcm.SendNotification(notification)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	fmt.Printf("Job #%s - Notification sent: %v\n", r.JobName(), ok)

	return ok
}

func collectAuxData() {
	t, err := cinemais.GetTheaters()
	if err != nil {
		return
	}
	for _, tt := range t {
		theaters[tt.ID] = tt
	}

	playing, err := cinemais.GetNowPlaying()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	collect := func(id int) {
		defer wg.Done()

		client := cinemais.New(
			cinemais.IncludeTrailer(true),
			cinemais.IncludeImages(true),
		)
		movie, err := client.GetMovie(id)
		if err == nil {
			movies[id] = *movie
		}
	}
	for _, m := range playing {
		wg.Add(1)
		go collect(m.ID)
	}
	wg.Wait()

	collecting = false
}

func today() Date {
	now := time.Now()
	return Date{
		Day:   now.Day(),
		Month: now.Month(),
		Year:  now.Year(),
	}
}

// IsToday check if a given time equals to current day
func IsToday(t *time.Time) bool {
	if t == nil {
		return false
	}
	return t.Day() == Today.Day && t.Month() == Today.Month && t.Year() == Today.Year
}
