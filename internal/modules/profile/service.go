package profile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"softixa-solutions.com/studentify/internal/models"
	"softixa-solutions.com/studentify/internal/utils"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Get(ctx context.Context, userID string) (*Response, error) {
	p, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toResponse(p), nil
}

func (s *Service) Update(ctx context.Context, userID string, req UpdateRequest) (*Response, error) {
	p, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.FullName != nil {
		p.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Phone != nil {
		p.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.ProfileImage != nil {
		p.ProfileImage = strings.TrimSpace(*req.ProfileImage)
	}
	if req.DateOfBirth != nil {
		raw := strings.TrimSpace(*req.DateOfBirth)
		if raw == "" {
			p.DateOfBirth = nil
		} else {
			dob, err := time.Parse("2006-01-02", raw)
			if err != nil {
				return nil, fmt.Errorf("invalid date_of_birth: %w", err)
			}
			p.DateOfBirth = &dob
		}
	}
	if req.InstituteName != nil {
		p.InstituteName = strings.TrimSpace(*req.InstituteName)
	}
	if req.DegreeOrClass != nil {
		p.DegreeOrClass = strings.TrimSpace(*req.DegreeOrClass)
	}
	if req.Section != nil {
		p.Section = strings.TrimSpace(*req.Section)
	}
	if req.StudentIDNumber != nil {
		p.StudentIDNumber = strings.TrimSpace(*req.StudentIDNumber)
	}
	if req.Address != nil {
		p.Address = strings.TrimSpace(*req.Address)
	}
	if req.City != nil {
		p.City = strings.TrimSpace(*req.City)
	}
	if req.Country != nil {
		p.Country = strings.TrimSpace(*req.Country)
	}
	if req.Bio != nil {
		p.Bio = strings.TrimSpace(*req.Bio)
	}
	if req.Others != nil {
		p.Others = strings.TrimSpace(*req.Others)
	}

	now := time.Now()
	p.UpdatedAt = now

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET name = $2, phone = $3, profile_image = $4, updated_at = $5
		WHERE id = $1`,
		userID, p.FullName, p.Phone, p.ProfileImage, now,
	)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO student_profiles (
			user_id, date_of_birth, institute_name, degree_or_class, section,
			student_id_number, address, city, country, bio, others, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (user_id) DO UPDATE SET
			date_of_birth = EXCLUDED.date_of_birth,
			institute_name = EXCLUDED.institute_name,
			degree_or_class = EXCLUDED.degree_or_class,
			section = EXCLUDED.section,
			student_id_number = EXCLUDED.student_id_number,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			country = EXCLUDED.country,
			bio = EXCLUDED.bio,
			others = EXCLUDED.others,
			updated_at = EXCLUDED.updated_at`,
		userID, nullTime(p.DateOfBirth), p.InstituteName, p.DegreeOrClass, p.Section,
		p.StudentIDNumber, p.Address, p.City, p.Country, p.Bio, p.Others, now,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert profile: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return toResponse(p), nil
}

func (s *Service) load(ctx context.Context, userID string) (*models.StudentProfile, error) {
	var p models.StudentProfile
	var dob sql.NullTime
	var (
		institute, degree, section, studentID sql.NullString
		address, city, country, bio, others   sql.NullString
		updatedAt                             sql.NullTime
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.name, u.email, u.phone, u.profile_image,
			p.date_of_birth, p.institute_name, p.degree_or_class, p.section,
			p.student_id_number, p.address, p.city, p.country, p.bio, p.others, p.updated_at
		FROM users u
		LEFT JOIN student_profiles p ON p.user_id = u.id
		WHERE u.id = $1`, userID,
	).Scan(
		&p.UserID, &p.FullName, &p.Email, &p.Phone, &p.ProfileImage,
		&dob, &institute, &degree, &section,
		&studentID, &address, &city, &country, &bio, &others, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load profile: %w", err)
	}

	if dob.Valid {
		t := dob.Time
		p.DateOfBirth = &t
	}
	p.InstituteName = institute.String
	p.DegreeOrClass = degree.String
	p.Section = section.String
	p.StudentIDNumber = studentID.String
	p.Address = address.String
	p.City = city.String
	p.Country = country.String
	p.Bio = bio.String
	p.Others = others.String
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	} else {
		p.UpdatedAt = time.Now()
	}

	return &p, nil
}

func toResponse(p *models.StudentProfile) *Response {
	return &Response{
		UserID:          p.UserID,
		FullName:        p.FullName,
		Email:           p.Email,
		Phone:           p.Phone,
		ProfileImage:    p.ProfileImage,
		DateOfBirth:     formatDate(p.DateOfBirth),
		InstituteName:   p.InstituteName,
		DegreeOrClass:   p.DegreeOrClass,
		Section:         p.Section,
		StudentIDNumber: p.StudentIDNumber,
		Address:         p.Address,
		City:            p.City,
		Country:         p.Country,
		Bio:             p.Bio,
		Others:          p.Others,
		UpdatedAt:       p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
