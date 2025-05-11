package dto

type MovieVariance struct {
	ID       uint    `json:"id"`
	Title    string  `json:"title"`
	Year     int32   `json:"year"`
	Runtime  int32   `json:"runtime"`
	Variance float64 `json:"variance"`
}
