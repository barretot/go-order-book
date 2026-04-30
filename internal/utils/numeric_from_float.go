package utils

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

func NumericFromFloat(value float64) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(int64(value * 100000000)),
		Exp:   -8,
		Valid: true,
	}
}
