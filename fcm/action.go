package fcm

type Action interface {
}

const (
	ActionViewTrailer  = "view_trailer"
	ActionMovieDetails = "movie_details"
)

type ViewTrailerAction struct {
	Type    string `json:"type"`
	VideoID string `json:"video_id"`
}

type MovieDetailsAction struct {
	Type    string `json:"type"`
	MovieID int    `json:"movie_id"`
}

func NewMovieDetailsAction(ID int) *MovieDetailsAction {
	return &MovieDetailsAction{
		Type:    ActionMovieDetails,
		MovieID: ID,
	}
}

func NewViewTrailerAction(ID string) *ViewTrailerAction {
	return &ViewTrailerAction{
		Type:    ActionViewTrailer,
		VideoID: ID,
	}
}
