package utils

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ParseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
}

func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func UUIDToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := UUIDToString(u)
	return &s
}

func StringPtrToUUID(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	return ParseUUID(*s)
}

func TimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func NumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return 0
	}
	return f.Float64
}

func Float64ToNumeric(f float64) pgtype.Numeric {
	// Convert to an integer-based representation to avoid floating-point issues.
	// We store 2 decimal places, so multiply by 100.
	scaled := int64(math.Round(f * 100))
	return pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -2,
		Valid: true,
	}
}

func TextFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func CurrentDate() pgtype.Date {
	now := time.Now()
	return pgtype.Date{Time: now, Valid: true}
}

func DateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func StringToDate(s string) pgtype.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: t, Valid: true}
}

func TextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
