package service

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func Float64ToNumeric(val float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%.2f", val))
	return num
}

// NumericToString renders a pgtype.Numeric as a plain decimal string.
// Numeric stores the value as Int * 10^Exp, so reading .Int directly
// (e.g. "19.99" -> Int=1999, Exp=-2) drops the decimal point.
func NumericToString(num pgtype.Numeric) string {
	if !num.Valid {
		return "0"
	}
	val, err := num.Value()
	if err != nil {
		return "0"
	}
	str, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return str
}

func StringToNumeric(str string) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	if err := num.Scan(str); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric value: %w", err)
	}
	return num, nil
}
