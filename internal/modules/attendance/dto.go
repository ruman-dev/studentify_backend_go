package attendance

type MarkRequest struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	SessionDate string `json:"session_date" validate:"required"` // YYYY-MM-DD
	Status      string `json:"status" validate:"required,oneof=present absent late excused"`
	Note        string `json:"note"`
}

type UpdateRequest struct {
	Status *string `json:"status" validate:"omitempty,oneof=present absent late excused"`
	Note   *string `json:"note"`
}

type RecordResponse struct {
	ID          string `json:"id"`
	SubjectID   string `json:"subjectId"`
	SubjectName string `json:"subjectName,omitempty"`
	SubjectCode string `json:"subjectCode,omitempty"`
	SessionDate string `json:"sessionDate"` // YYYY-MM-DD
	Status      string `json:"status"`
	Note        string `json:"note"`
	MarkedAt    string `json:"markedAt"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SubjectSummary struct {
	SubjectID   string  `json:"subjectId"`
	SubjectName string  `json:"subjectName"`
	SubjectCode string  `json:"subjectCode"`
	Present     int     `json:"present"`
	Absent      int     `json:"absent"`
	Late        int     `json:"late"`
	Excused     int     `json:"excused"`
	Attended    int     `json:"attended"` // present + late
	Total       int     `json:"total"`    // present + absent + late (excused excluded)
	Percentage  float64 `json:"percentage"`
}

type OverviewResponse struct {
	OverallPercentage float64          `json:"overallPercentage"`
	TotalAttended     int              `json:"totalAttended"`
	TotalClasses      int              `json:"totalClasses"`
	Subjects          []SubjectSummary `json:"subjects"`
}

type SubjectDetailResponse struct {
	Summary SubjectSummary   `json:"summary"`
	Records []RecordResponse `json:"records"`
}
