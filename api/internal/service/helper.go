package service

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

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

// NumericToInt64 rounds to the nearest whole unit, for Midtrans's integer gross_amount.
func NumericToInt64(num pgtype.Numeric) (int64, error) {
	f, err := strconv.ParseFloat(NumericToString(num), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %w", err)
	}
	return int64(math.Round(f)), nil
}

func StringToNumeric(str string) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	if err := num.Scan(str); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric value: %w", err)
	}
	return num, nil
}

// multiplyNumericByInt multiplies a NUMERIC by an integer factor without
// going through float64, preserving exact decimal precision for money math.
func multiplyNumericByInt(n pgtype.Numeric, factor int32) pgtype.Numeric {
	if !n.Valid || n.Int == nil {
		return pgtype.Numeric{Int: big.NewInt(0), Valid: true}
	}
	return pgtype.Numeric{
		Int:   new(big.Int).Mul(n.Int, big.NewInt(int64(factor))),
		Exp:   n.Exp,
		Valid: true,
	}
}

// addNumeric adds two NUMERIC values without going through float64,
// aligning them to their common (more precise) exponent first since Numeric
// stores value as Int * 10^Exp and different rows may carry different Exp.
func addNumeric(a, b pgtype.Numeric) pgtype.Numeric {
	if !a.Valid || a.Int == nil {
		a = pgtype.Numeric{Int: big.NewInt(0), Exp: b.Exp, Valid: true}
	}
	if !b.Valid || b.Int == nil {
		return a
	}

	exp := a.Exp
	if b.Exp < exp {
		exp = b.Exp
	}

	scale := func(n *big.Int, from int32) *big.Int {
		if from == exp {
			return n
		}
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(from-exp)), nil)
		return new(big.Int).Mul(n, factor)
	}

	return pgtype.Numeric{
		Int:   new(big.Int).Add(scale(a.Int, a.Exp), scale(b.Int, b.Exp)),
		Exp:   exp,
		Valid: true,
	}
}
