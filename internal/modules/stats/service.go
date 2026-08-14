package stats

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type OverviewStats struct {
	GrowthJourney    GrowthJourney     `json:"growthJourney"`
	SubjectStrengths []SubjectStrength `json:"subjectStrengths"`
	AttendanceTrend  []AttendanceTrend `json:"attendanceTrend"`
	WaysToLevelUp    []LevelUpAdvice   `json:"waysToLevelUp"`
}

type GrowthJourney struct {
	RisingLevel          string `json:"risingLevel"`
	StreakDays           int    `json:"streakDays"`
	StreakTrend          string `json:"streakTrend"`
	GPA                  string `json:"gpa,omitempty"` // Omitted for now as requested
	GPATrend             string `json:"gpaTrend,omitempty"`
	AttendancePercentage int    `json:"attendancePercentage"`
	AttendanceTrend      string `json:"attendanceTrend"`
	TasksDone            int    `json:"tasksDone"`
	TasksTrend           string `json:"tasksTrend"`
	FocusScore           int    `json:"focusScore"`
	FocusTrend           string `json:"focusTrend"`
}

type SubjectStrength struct {
	SubjectName string `json:"subjectName"`
	Color       string `json:"color"`
	Percentage  int    `json:"percentage"`
}

type AttendanceTrend struct {
	WeekLabel  string `json:"weekLabel"`
	Percentage int    `json:"percentage"`
}

type LevelUpAdvice struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// predefined colors for charts
var colors = []string{"#3b82f6", "#8b5cf6", "#10b981", "#f59e0b", "#ef4444", "#ec4899"}

func (s *Service) GetOverview(ctx context.Context, userID string) (*OverviewStats, error) {
	stats := &OverviewStats{
		GrowthJourney: GrowthJourney{
			RisingLevel:     "Level",
			StreakTrend:     "+14% This week", // hardcoded trend for now
			AttendanceTrend: "+3%",
			TasksTrend:      "+6",
			FocusTrend:      "-4",
		},
		SubjectStrengths: []SubjectStrength{},
		AttendanceTrend:  []AttendanceTrend{},
		WaysToLevelUp:    []LevelUpAdvice{},
	}

	// 1. Attendance stats
	var totalAttendance, presentCount int
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'present' THEN 1 ELSE 0 END), 0)
		FROM attendance_records 
		WHERE user_id = $1
	`, userID).Scan(&totalAttendance, &presentCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch attendance stats: %w", err)
	}

	if totalAttendance > 0 {
		stats.GrowthJourney.AttendancePercentage = int(math.Round(float64(presentCount) / float64(totalAttendance) * 100))
	} else {
		stats.GrowthJourney.AttendancePercentage = 0 // Default if no records
	}

	// 2. Tasks Done (Assignments with obtained marks or status completed)
	// We'll count ones with obtained_marks IS NOT NULL as 'done' for simplicity
	var tasksDone int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM assignments 
		WHERE user_id = $1 AND obtained_marks IS NOT NULL
	`, userID).Scan(&tasksDone)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch tasks stats: %w", err)
	}
	stats.GrowthJourney.TasksDone = tasksDone

	// 3. Subject Strengths
	rows, err := s.db.QueryContext(ctx, `
		WITH subject_scores AS (
			SELECT 
				s.id, s.name,
				COALESCE(SUM(a.obtained_marks), 0) + COALESCE(SUM(e.obtained_marks), 0) as obtained,
				COALESCE(SUM(a.total_marks), 0) + COALESCE(SUM(e.total_marks), 0) as total
			FROM subjects s
			LEFT JOIN assignments a ON s.id = a.subject_id AND a.user_id = $1 AND a.obtained_marks IS NOT NULL AND a.total_marks IS NOT NULL
			LEFT JOIN exams e ON s.id = e.subject_id AND e.user_id = $1 AND e.obtained_marks IS NOT NULL AND e.total_marks IS NOT NULL
			WHERE s.user_id = $1
			GROUP BY s.id, s.name
		)
		SELECT name, obtained, total FROM subject_scores WHERE total > 0
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subject strengths: %w", err)
	}
	defer rows.Close()

	var lowestSubject string
	lowestScore := 101.0
	highestSubject := ""
	highestScore := -1.0

	colorIdx := 0
	for rows.Next() {
		var name string
		var obtained, total float64
		if err := rows.Scan(&name, &obtained, &total); err != nil {
			return nil, err
		}
		percentage := int(math.Round((obtained / total) * 100))
		stats.SubjectStrengths = append(stats.SubjectStrengths, SubjectStrength{
			SubjectName: name,
			Percentage:  percentage,
			Color:       colors[colorIdx%len(colors)],
		})
		colorIdx++

		if float64(percentage) < lowestScore {
			lowestScore = float64(percentage)
			lowestSubject = name
		}
		if float64(percentage) > highestScore {
			highestScore = float64(percentage)
			highestSubject = name
		}
	}

	// 4. Attendance Trend (Last 6 weeks)
	// We'll generate the last 6 weeks, counting attendance
	// Group by week using date_trunc
	trendRows, err := s.db.QueryContext(ctx, `
		SELECT 
			date_trunc('week', session_starts_at) as week_start,
			COUNT(*) as total_sessions,
			SUM(CASE WHEN status = 'present' THEN 1 ELSE 0 END) as present_sessions
		FROM attendance_records
		WHERE user_id = $1 AND session_starts_at >= NOW() - INTERVAL '6 weeks'
		GROUP BY date_trunc('week', session_starts_at)
		ORDER BY week_start ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance trend: %w", err)
	}
	defer trendRows.Close()

	weekCounter := 1
	for trendRows.Next() {
		var weekStart time.Time
		var totalSessions, presentSessions int
		if err := trendRows.Scan(&weekStart, &totalSessions, &presentSessions); err != nil {
			return nil, err
		}
		perc := 0
		if totalSessions > 0 {
			perc = int(math.Round(float64(presentSessions) / float64(totalSessions) * 100))
		}
		label := fmt.Sprintf("W%d", weekCounter)
		stats.AttendanceTrend = append(stats.AttendanceTrend, AttendanceTrend{
			WeekLabel:  label,
			Percentage: perc,
		})
		weekCounter++
	}

	// Pad with empty weeks if less than 6
	for len(stats.AttendanceTrend) < 6 {
		stats.AttendanceTrend = append(stats.AttendanceTrend, AttendanceTrend{
			WeekLabel:  fmt.Sprintf("W%d", len(stats.AttendanceTrend)+1),
			Percentage: stats.GrowthJourney.AttendancePercentage, // default to overall avg
		})
	}

	// Synthetic streak & focus score
	stats.GrowthJourney.StreakDays = presentCount / 2 // Just a placeholder formula
	if stats.GrowthJourney.StreakDays == 0 {
		stats.GrowthJourney.StreakDays = 12 // Fallback to design mockup for demo
	}
	stats.GrowthJourney.FocusScore = (stats.GrowthJourney.AttendancePercentage + 90) / 2 // Placeholder

	// 5. Ways to Level Up
	if lowestSubject != "" {
		strongText := ""
		if highestSubject != "" && highestSubject != lowestSubject {
			strongText = fmt.Sprintf("You're strongest in %s. ", highestSubject)
		}
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       fmt.Sprintf("Boost %s this week", lowestSubject),
			Description: fmt.Sprintf("%sTwo 30-min reading sessions can lift %s by ~8%%.", strongText, lowestSubject),
			Icon:        "book", // Map to a book icon on frontend
		})
	} else {
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       "Keep up the great work!",
			Description: "Your subjects are looking good. Maintain your study habits.",
			Icon:        "book",
		})
	}

	// Check peak study day
	var peakDay string
	err = s.db.QueryRowContext(ctx, `
		SELECT to_char(session_starts_at, 'Day') as day_name
		FROM attendance_records
		WHERE user_id = $1 AND status = 'present'
		GROUP BY day_name
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, userID).Scan(&peakDay)
	if err == nil && peakDay != "" {
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       fmt.Sprintf("Protect your %s streak", peakDay),
			Description: fmt.Sprintf("%s is your peak study day. Keep that rhythm — it's driving your weekly gain.", peakDay),
			Icon:        "fire", // Map to a flame icon on frontend
		})
	} else {
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       "Protect your daily streak",
			Description: "Consistency is key. Keep logging in and completing tasks.",
			Icon:        "fire",
		})
	}

	// Check pending assignments
	var pendingCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM assignments
		WHERE user_id = $1 AND obtained_marks IS NULL
	`, userID).Scan(&pendingCount)
	if err == nil && pendingCount > 0 {
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       "Close the potential gap",
			Description: fmt.Sprintf("Finish %d pending assignments to boost your scores closer to your potential.", pendingCount),
			Icon:        "flag", // Map to a flag icon
		})
	} else {
		stats.WaysToLevelUp = append(stats.WaysToLevelUp, LevelUpAdvice{
			Title:       "You're all caught up",
			Description: "No pending assignments. Review your notes to stay ahead.",
			Icon:        "flag",
		})
	}

	return stats, nil
}
