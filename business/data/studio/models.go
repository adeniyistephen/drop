package studio

import "time"

type Info struct {
	ID           string    `bson:"_id"`
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"email,required"`
	SocialHandle string    `json:"socials"`
	Description  string    `json:"description"`
	Created_at   time.Time `json:"created_at"`
	City         string    `json:"city" validate:"required"`
	State        string    `json:"state" validate:"required"`
	Country      string    `json:"country" validate:"required"`
	Updated_at   time.Time `json:"updated_at"`
}

// NewUser contains information needed to create a new User.
type NewStudio struct {
	Name         string    `json:"name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	SocialHandle string    `json:"socials"`
	City         string    `json:"city" validate:"required"`
	Description  string    `json:"description"`
	State        string    `json:"state" validate:"required"`
	Country      string    `json:"country" validate:"required"`
	Created_at   time.Time `json:"created_at"`
}

// UpdateUser defines what information may be provided to modify an existing
// User.
type UpdateStudio struct {
	Name         *string `json:"name"`
	Email        *string `json:"email" validate:"omitempty,email"`
	SocialHandle *string `json:"socials"`
	Description  *string `json:"description"`
	City         *string `json:"city" validate:"required"`
	State        *string `json:"state" validate:"required"`
	Country      *string `json:"country" validate:"required"`
}
