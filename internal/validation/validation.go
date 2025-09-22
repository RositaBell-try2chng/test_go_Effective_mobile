package validation

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"subscription-aggregator/internal/models"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateCreateSubscription validates CreateSubscriptionRequest
func ValidateCreateSubscription(req models.CreateSubscriptionRequest) error {
	var errors ValidationErrors

	// Validate service_name
	if req.ServiceName == "" {
		errors = append(errors, ValidationError{Field: "service_name", Message: "название сервиса обязательно"})
	} else if len(req.ServiceName) > 255 {
		errors = append(errors, ValidationError{Field: "service_name", Message: "название сервиса не должно превышать 255 символов"})
	}

	// Validate price
	if req.Price < 0 {
		errors = append(errors, ValidationError{Field: "price", Message: "стоимость не может быть отрицательной"})
	}

	// Validate user_id
	if req.UserID == "" {
		errors = append(errors, ValidationError{Field: "user_id", Message: "ID пользователя обязателен"})
	} else if _, err := uuid.Parse(req.UserID); err != nil {
		errors = append(errors, ValidationError{Field: "user_id", Message: "ID пользователя должен быть в формате UUID"})
	}

	// Validate start_date
	if req.StartDate == "" {
		errors = append(errors, ValidationError{Field: "start_date", Message: "дата начала обязательна"})
	} else if _, err := ParseMonthYear(req.StartDate); err != nil {
		errors = append(errors, ValidationError{Field: "start_date", Message: "дата начала должна быть в формате MM-YYYY"})
	}

	// Validate end_date if provided
	if req.EndDate != "" {
		if _, err := ParseMonthYear(req.EndDate); err != nil {
			errors = append(errors, ValidationError{Field: "end_date", Message: "дата окончания должна быть в формате MM-YYYY"})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateUpdateSubscription validates UpdateSubscriptionRequest
func ValidateUpdateSubscription(req models.UpdateSubscriptionRequest) error {
	var errors ValidationErrors

	// Validate service_name if provided
	if req.ServiceName != "" && len(req.ServiceName) > 255 {
		errors = append(errors, ValidationError{Field: "service_name", Message: "название сервиса не должно превышать 255 символов"})
	}

	// Validate price if provided
	if req.Price < 0 {
		errors = append(errors, ValidationError{Field: "price", Message: "стоимость не может быть отрицательной"})
	}

	// Validate start_date if provided
	if req.StartDate != "" {
		if _, err := ParseMonthYear(req.StartDate); err != nil {
			errors = append(errors, ValidationError{Field: "start_date", Message: "дата начала должна быть в формате MM-YYYY"})
		}
	}

	// Validate end_date if provided
	if req.EndDate != "" {
		if req.EndDate == "null" {
			// Allow null value for clearing end_date
		} else if _, err := ParseMonthYear(req.EndDate); err != nil {
			errors = append(errors, ValidationError{Field: "end_date", Message: "дата окончания должна быть в формате MM-YYYY"})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateAggregationRequest validates AggregationRequest
func ValidateAggregationRequest(req models.AggregationRequest) error {
	var errors ValidationErrors

	// Validate start_date
	if req.StartDate == "" {
		errors = append(errors, ValidationError{Field: "start_date", Message: "дата начала обязательна"})
	} else if _, err := ParseMonthYear(req.StartDate); err != nil {
		errors = append(errors, ValidationError{Field: "start_date", Message: "дата начала должна быть в формате MM-YYYY"})
	}

	// Validate end_date
	if req.EndDate == "" {
		errors = append(errors, ValidationError{Field: "end_date", Message: "дата окончания обязательна"})
	} else if _, err := ParseMonthYear(req.EndDate); err != nil {
		errors = append(errors, ValidationError{Field: "end_date", Message: "дата окончания должна быть в формате MM-YYYY"})
	}

	// Validate user_id if provided
	if req.UserID != nil {
		if req.UserID.String() == "" {
			errors = append(errors, ValidationError{Field: "user_id", Message: "ID пользователя не может быть пустым"})
		}
	}

	// Validate service_name if provided
	if req.ServiceName != nil && len(*req.ServiceName) > 255 {
		errors = append(errors, ValidationError{Field: "service_name", Message: "название сервиса не должно превышать 255 символов"})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ValidationError{Field: "id", Message: "ID должен быть в формате UUID"}
	}
	return nil
}

// ValidateJSONFields validates that only allowed fields are present in JSON
func ValidateJSONFields(data map[string]interface{}, allowedFields []string) error {
	allowedFieldsMap := make(map[string]bool)
	for _, field := range allowedFields {
		allowedFieldsMap[field] = true
	}

	var errors ValidationErrors
	for field := range data {
		if !allowedFieldsMap[field] {
			errors = append(errors, ValidationError{Field: field, Message: "поле не разрешено"})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ParseMonthYear parses date in MM-YYYY format
func ParseMonthYear(dateStr string) (time.Time, error) {
	// Expected format: "MM-YYYY"
	parts := strings.Split(dateStr, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid date format")
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month")
	}

	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 1900 || year > 2100 {
		return time.Time{}, fmt.Errorf("invalid year")
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

// GetAllowedFieldsForCreate returns allowed fields for creating subscription
func GetAllowedFieldsForCreate() []string {
	return []string{"service_name", "price", "user_id", "start_date", "end_date"}
}

// GetAllowedFieldsForUpdate returns allowed fields for updating subscription
func GetAllowedFieldsForUpdate() []string {
	return []string{"service_name", "price", "start_date", "end_date"}
}

// GetAllowedFieldsForAggregation returns allowed fields for aggregation
func GetAllowedFieldsForAggregation() []string {
	return []string{"user_id", "service_name", "start_date", "end_date"}
}
