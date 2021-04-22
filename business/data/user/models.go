package user

import (
	"time"

	"github.com/lib/pq"
)

type Info struct {
	ID           string         `bson:"_id"`
	Name         string         `json:"name" validate:"required,min=2,max=100"`
	Email        string         `json:"email" validate:"email,required"`
	Roles        pq.StringArray `json:"roles"`
	Password     string         `json:"password"`
	PasswordHash []byte         `json:"password_hash"`
	Created_at   time.Time      `json:"created_at"`
	Updated_at   time.Time      `json:"updated_at"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// UpdateUser defines what information may be provided to modify an existing
// User.
type UpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email" validate:"omitempty,email"`
	Roles           []string `json:"roles"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}
