package models

import (
	"time"
	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int        `json:"price" db:"price"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       int    `json:"price" validate:"required,min=0"`
	UserID      string `json:"user_id" validate:"required,uuid"`
	StartDate   string `json:"start_date" validate:"required"` // Format: "MM-YYYY"
	EndDate     string `json:"end_date,omitempty"`            // Format: "MM-YYYY" or empty
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name,omitempty"`
	Price       int    `json:"price,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	ServiceName *string    `json:"service_name,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type AggregationRequest struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	ServiceName *string    `json:"service_name,omitempty"`
	StartDate   string     `json:"start_date" validate:"required"` // Format: "MM-YYYY"
	EndDate     string     `json:"end_date" validate:"required"`   // Format: "MM-YYYY"
}

type AggregationResponse struct {
	TotalCost int64      `json:"total_cost"`
	Period    string     `json:"period"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
}
