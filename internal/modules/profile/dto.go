package profile

import "time"

type UpdateRequest struct {
	FullName        *string `json:"full_name"`
	Phone           *string `json:"phone"`
	ProfileImage    *string `json:"profile_image"`
	DateOfBirth     *string `json:"date_of_birth"`
	InstituteName   *string `json:"institute_name"`
	DegreeOrClass   *string `json:"degree_or_class"`
	Section         *string `json:"section"`
	StudentIDNumber *string `json:"student_id_number"`
	Address         *string `json:"address"`
	City            *string `json:"city"`
	Country         *string `json:"country"`
	Bio             *string `json:"bio"`
	Others          *string `json:"others"`
}

type Response struct {
	UserID          string  `json:"userId"`
	FullName        string  `json:"fullName"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	ProfileImage    string  `json:"profileImage"`
	DateOfBirth     *string `json:"dateOfBirth,omitempty"`
	InstituteName   string  `json:"instituteName"`
	DegreeOrClass   string  `json:"degreeOrClass"`
	Section         string  `json:"section"`
	StudentIDNumber string  `json:"studentIdNumber"`
	Address         string  `json:"address"`
	City            string  `json:"city"`
	Country         string  `json:"country"`
	Bio             string  `json:"bio"`
	Others          string  `json:"others"`
	UpdatedAt       string  `json:"updatedAt"`
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
