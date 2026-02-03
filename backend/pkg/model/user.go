// Package model provides data models for the application
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255)" json:"-"` // LDAP users may have empty password
	UserType     string    `gorm:"type:varchar(50);not null;default:'local'" json:"user_type"`
	DisplayName  string    `gorm:"type:varchar(255)" json:"display_name"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeUpdate hook updates the UpdatedAt timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// UserType constants
const (
	UserTypeLocal = "local"
	UserTypeLDAP  = "ldap"
)
