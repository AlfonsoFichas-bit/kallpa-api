package domain

import (
	"time"
)

// UserRole represents the access level of a user.
type UserRole string

const (
	// RoleSuperadmin has complete access across all stores.
	RoleSuperadmin UserRole = "superadmin"
	// RoleStoreOwner has full access to their specific store.
	RoleStoreOwner UserRole = "store_owner"
	// RoleSalesAgent has limited access for sales operations.
	RoleSalesAgent UserRole = "sales_agent"
)

// User represents a system user with authentication and role details.
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID      *int64    `json:"store_id" gorm:"index"`
	Name         string    `json:"name" gorm:"type:varchar(150);not null"`
	Email        string    `json:"email" gorm:"type:varchar(150);unique;not null"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
	Role         UserRole  `json:"role" gorm:"type:varchar(50);default:'store_owner'"`
	Phone        *string   `json:"phone" gorm:"type:varchar(20)"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
