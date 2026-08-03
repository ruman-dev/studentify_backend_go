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

// Weekday values used in Subject.ScheduleDays.
const (
	WeekdayMonday    = "monday"
	WeekdayTuesday   = "tuesday"
	WeekdayWednesday = "wednesday"
	WeekdayThursday  = "thursday"
	WeekdayFriday    = "friday"
	WeekdaySaturday  = "saturday"
	WeekdaySunday    = "sunday"
)

type Subject struct {
	ID           string
	UserID       string
	TeacherID    *string
	Name         string
	Code         string
	Description  string
	CreditHours  *float64
	ScheduleDays []string // e.g. monday, wednesday
	StartTime    string   // HH:MM, required
	EndTime      string   // HH:MM, required
	Room         string
	MeetingLink  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

// Default exam type suggestions. Custom types are auto-saved when used on POST/PUT /exams.
var DefaultExamTypes = []string{
	"Mid-term",
	"Final",
	"Lab test",
	"Quiz",
	"Presentation",
	"Practical",
}

type ExamType struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt time.Time
}

type Exam struct {
	ID              string
	UserID          string
	SubjectID       string
	Title           string // optional
	ExamType        string // optional; defaults suggestions in DefaultExamTypes
	ExamDate        time.Time
	DurationMinutes *int
	Venue           string
	TotalMarks      *float64
	ObtainedMarks   *float64
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
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

// Attendance status values for self-tracked class sessions.
const (
	AttendancePresent = "present"
	AttendanceAbsent  = "absent"
	AttendanceLate    = "late"
	AttendanceExcused = "excused"
)

type AttendanceRecord struct {
	ID          string
	UserID      string
	SubjectID   string
	SessionDate time.Time // date-only (UTC midnight)
	Status      string
	Note        string
	MarkedAt    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
