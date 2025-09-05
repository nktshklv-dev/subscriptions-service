package subscriptions

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateDTO struct {
	UserID      string  `json:"user_id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	Start       string  `json:"start"`
	End         *string `json:"end,omitempty"`
}

type UpdateDTO struct {
	UserID      string  `json:"user_id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	Start       string  `json:"start"`
	End         *string `json:"end,omitempty"`
}

func ParseMonth(s string) (time.Time, error) {
	if len(s) != 7 {
		return time.Time{}, fmt.Errorf("bad month format: %q", s)
	}
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("bad month format: %q", s)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func (in CreateDTO) Validate() (uuid.UUID, string, int, time.Time, *time.Time, error) {
	uid, err := uuid.Parse(in.UserID)
	if err != nil {
		return uuid.Nil, "", 0, time.Time{}, nil, errors.New("user_id must be UUID")
	}
	if in.ServiceName == "" {
		return uuid.Nil, "", 0, time.Time{}, nil, errors.New("service_name is required")
	}
	if in.Price < 0 {
		return uuid.Nil, "", 0, time.Time{}, nil, errors.New("price must be >= 0")
	}
	start, err := ParseMonth(in.Start)
	if err != nil {
		return uuid.Nil, "", 0, time.Time{}, nil, errors.New("start must be MM-YYYY")
	}
	var endPtr *time.Time
	if in.End != nil && *in.End != "" {
		et, err := ParseMonth(*in.End)
		if err != nil {
			return uuid.Nil, "", 0, time.Time{}, nil, errors.New("end must be MM-YYYY")
		}
		if et.Before(start) {
			return uuid.Nil, "", 0, time.Time{}, nil, errors.New("end must be >= start")
		}
		endPtr = &et
	}
	return uid, in.ServiceName, in.Price, start, endPtr, nil
}

func (in UpdateDTO) Validate() (uuid.UUID, string, int, time.Time, *time.Time, error) {
	return CreateDTO(in).Validate()
}

type SummaryDTO struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
}

type SummaryResult struct {
	Total int `json:"total"`
}

func (in SummaryDTO) Parse() (time.Time, time.Time, *uuid.UUID, *string, error) {
	from, err := ParseMonth(in.From)
	if err != nil {
		return time.Time{}, time.Time{}, nil, nil, fmt.Errorf("from must be MM-YYYY")
	}
	to, err := ParseMonth(in.To)
	if err != nil {
		return time.Time{}, time.Time{}, nil, nil, fmt.Errorf("to must be MM-YYYY")
	}
	var uidPtr *uuid.UUID
	if in.UserID != nil && *in.UserID != "" {
		uid, err := uuid.Parse(*in.UserID)
		if err != nil {
			return time.Time{}, time.Time{}, nil, nil, fmt.Errorf("user_id must be uuid")
		}
		uidPtr = &uid
	}
	var svcPtr *string
	if in.ServiceName != nil && *in.ServiceName != "" {
		s := *in.ServiceName
		svcPtr = &s
	}
	return from, to, uidPtr, svcPtr, nil
}
