package model

type Assessment struct {
	ID            int     `json:"id"`
	ApplicationID int     `json:"application_id"`
	LecturerID    int     `json:"lecturer_id"`
	Score         float64 `json:"score"`
	Comment       string  `json:"comment"`
}
