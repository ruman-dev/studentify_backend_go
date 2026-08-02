package subjects

type CreateRequest struct {
	Name         string   `json:"name" validate:"required"`
	Code         string   `json:"code"`
	Description  string   `json:"description"`
	TeacherID    *string  `json:"teacher_id"`
	CreditHours  *float64 `json:"credit_hours"`
	ScheduleDays []string `json:"schedule_days" validate:"required,min=1,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    string   `json:"start_time" validate:"required,timeofday"`
	EndTime      string   `json:"end_time" validate:"required,timeofday"`
	Room         string   `json:"room"`
	MeetingLink  string   `json:"meeting_link"`
}

type UpdateRequest struct {
	Name         *string  `json:"name" validate:"omitempty,min=1"`
	Code         *string  `json:"code"`
	Description  *string  `json:"description"`
	TeacherID    *string  `json:"teacher_id"`
	CreditHours  *float64 `json:"credit_hours"`
	ScheduleDays []string `json:"schedule_days" validate:"omitempty,min=1,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    *string  `json:"start_time" validate:"omitempty,timeofday"`
	EndTime      *string  `json:"end_time" validate:"omitempty,timeofday"`
	Room         *string  `json:"room"`
	MeetingLink  *string  `json:"meeting_link"`
}

type Response struct {
	ID           string   `json:"id"`
	TeacherID    *string  `json:"teacherId,omitempty"`
	Name         string   `json:"name"`
	Code         string   `json:"code"`
	Description  string   `json:"description"`
	CreditHours  *float64 `json:"creditHours,omitempty"`
	ScheduleDays []string `json:"scheduleDays"`
	StartTime    string   `json:"startTime"`
	EndTime      string   `json:"endTime"`
	Room         string   `json:"room"`
	MeetingLink  string   `json:"meetingLink"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}
