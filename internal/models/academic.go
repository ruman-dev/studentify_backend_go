package models

import "time"

type StudentProfile struct {
	UserID          string
	FullName        string
	Email           string
	Phone           string
	ProfileImage    string
	DateOfBirth     *time.Time
	InstituteName   string
	DegreeOrClass   string
	Section         string
	StudentIDNumber string
	Address         string
	City            string
	Country         string
	Bio             string
	Others          string
	UpdatedAt       time.Time
}

type Teacher struct {
	ID          string
	UserID      string
	FullName    string
	Email       string
	Phone       string
	Department  string
	Designation string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Subject struct {
	ID          string
	UserID      string
	TeacherID   *string
	Name        string
	Code        string
	Description string
	CreditHours *float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Assignment struct {
	ID          string
	UserID      string
	SubjectID   string
	Title       string
	Description string
	DueDate     time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Exam struct {
	ID               string
	UserID           string
	SubjectID        string
	Title            string
	ExamDate         time.Time
	DurationMinutes  *int
	Venue            string
	TotalMarks       *float64
	ObtainedMarks    *float64
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Event struct {
	ID          string
	UserID      string
	SubjectID   *string
	Title       string
	Description string
	Location    string
	StartsAt    time.Time
	EndsAt      *time.Time
	EventType   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
