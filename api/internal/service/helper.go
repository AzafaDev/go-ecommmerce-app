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
