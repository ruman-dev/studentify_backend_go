package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"softixa-solutions.com/studentify/internal/config"
	"softixa-solutions.com/studentify/internal/models"
	"softixa-solutions.com/studentify/utils"
)

type User struct {
	ID           string
	Phone        string
	FullName     string
	Email        string
	ProfileImage string
	Role         string
	PasswordHash string
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	ProfileImage string `json:"profile_image"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var sources []utils.ErrorSource
	if !utils.IsValidEmail(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if req.Password == "" {
		sources = append(sources, utils.ErrorSource{Path: "password", Message: "Password is required"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	var user User
	err := config.DB.QueryRow(`SELECT id, phone, name, email, profile_image, role, password 
        FROM users WHERE email=$1`, req.Email).Scan(
		&user.ID, &user.Phone, &user.FullName, &user.Email, &user.ProfileImage, &user.Role, &user.PasswordHash,
	)

	if err == sql.ErrNoRows {
		utils.Error(w, http.StatusUnauthorized, "User not found", err)
		return
	} else if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Incorrect password", err)
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Phone, user.Role)

	utils.Success(w, http.StatusOK, "Login successful", models.AuthResponse{
		UserID:       user.ID,
		Phone:        user.Phone,
		FullName:     user.FullName,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
		Role:         user.Role,
		Token:        token,
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	var sources []utils.ErrorSource
	if req.FullName == "" {
		sources = append(sources, utils.ErrorSource{Path: "full_name", Message: "Full name is required"})
	}
	if !utils.IsValidEmail(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if !utils.IsValidPhone(req.Phone) {
		sources = append(sources, utils.ErrorSource{Path: "phone", Message: "Invalid phone number. Use international format e.g. +1234567890"})
	}
	if !utils.IsValidPassword(req.Password) {
		sources = append(sources, utils.ErrorSource{Path: "password", Message: "Password must be at least 6 characters and include a letter and a number"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Password hashing failed", err)
		return
	}

	user := models.User{
		ID:           uuid.New().String(),
		Name:         req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     string(hashedPassword),
		ProfileImage: req.ProfileImage,
		IsVerified:   false,
		IsActive:     true,
		IsDeleted:    false,
		IsBanned:     false,
		IsSuspended:  false,
		IsLocked:     false,
		IsExpired:    false,
		Role:         "USER",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = config.DB.Exec(`INSERT INTO users 
        (id, name, email, phone, password, profile_image, is_verified, is_active, is_deleted, is_banned, is_suspended, is_locked, is_expired, role, created_at, updated_at) 
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		user.ID, user.Name, user.Email, user.Phone, user.Password, user.ProfileImage,
		user.IsVerified, user.IsActive, user.IsDeleted, user.IsBanned, user.IsSuspended,
		user.IsLocked, user.IsExpired, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Database insert failed", err)
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Phone, user.Role)

	utils.Success(w, http.StatusCreated, "Registration successful", models.AuthResponse{
		UserID:       user.ID,
		Phone:        user.Phone,
		FullName:     user.Name,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
		Role:         user.Role,
		Token:        token,
	})
}
