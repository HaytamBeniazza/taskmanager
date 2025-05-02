package main

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Role represents user permission level
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	FirstName    string    `json:"firstName,omitempty"`
	LastName     string    `json:"lastName,omitempty"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	LastLogin    time.Time `json:"lastLogin,omitempty"`
	Preferences  UserPreferences `json:"preferences"`
}

// UserPreferences stores user-specific settings
type UserPreferences struct {
	DefaultView     string `json:"defaultView"` // list, kanban, calendar
	DefaultCategory string `json:"defaultCategory,omitempty"`
	DefaultPriority string `json:"defaultPriority,omitempty"`
	Theme           string `json:"theme,omitempty"` // light, dark, system
	TimeFormat      string `json:"timeFormat,omitempty"` // 12h, 24h
	DateFormat      string `json:"dateFormat,omitempty"` // MM/DD/YYYY, DD/MM/YYYY, etc.
	WeekStart       int    `json:"weekStart,omitempty"` // 0=Sunday, 1=Monday
	Notifications   bool   `json:"notifications"`
}

// NewUser creates a new user with the given details
func NewUser(username, email, password string) (*User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         RoleUser,
		CreatedAt:    time.Now(),
		Preferences: UserPreferences{
			DefaultView:   "list",
			Theme:         "system",
			TimeFormat:    "24h",
			DateFormat:    "YYYY-MM-DD",
			WeekStart:     1, // Monday
			Notifications: true,
		},
	}, nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword changes the user's password
func (u *User) UpdatePassword(currentPassword, newPassword string) error {
	// Verify current password
	if !u.CheckPassword(currentPassword) {
		return errors.New("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	u.LastLogin = time.Now()
}
