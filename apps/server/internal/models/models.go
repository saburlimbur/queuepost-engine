package models

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type SocialAccount struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	Platform       string    `json:"platform" db:"platform"`
	PlatformUserID string    `json:"platform_user_id" db:"platform_user_id"`
	AccessToken    string    `json:"access_token" db:"access_token"`
	RefreshToken   *string   `json:"refresh_token,omitempty" db:"refresh_token"`
	TokenExpiresAt time.Time `json:"token_expires_at" db:"token_expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Post struct {
	ID              int        `json:"id" db:"id"`
	UserID          int        `json:"user_id" db:"user_id"`
	SocialAccountID int        `json:"social_account_id" db:"social_account_id"`
	Content         string     `json:"content" db:"content"`
	MediaUrls       []string   `json:"media_urls" db:"media_urls"`
	Status          string     `json:"status" db:"status"` // PENDING, SCHEDULED, PUBLISHED, FAILED
	ScheduledAt     time.Time  `json:"scheduled_at" db:"scheduled_at"`
	PublishedAt     *time.Time `json:"published_at,omitempty" db:"published_at"`
	ErrorMessage    *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
