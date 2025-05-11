package dto

type MovieWithoutReviews struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Year    int32  `json:"year"`
	Runtime int32  `json:"runtime"`
}
