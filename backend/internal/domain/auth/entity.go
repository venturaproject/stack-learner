package auth

import "time"

// User is the aggregate root for authentication. Pure domain type — no json /
// gorm tags. HTTP-facing shapes live in internal/application/auth (views).
type User struct {
	ID           int32
	Name         string
	Email        string
	Username     *string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID          int32
	Name        string
	GuardName   string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Permission struct {
	ID        int32
	Name      string
	GuardName string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefreshToken struct {
	ID        int32
	JTI       string
	UserID    int32
	Revoked   bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserSettings struct {
	UserID int32
	Avatar *string
}
