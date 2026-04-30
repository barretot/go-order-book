package utils_test

import (
	"testing"

	"github.com/barretot/go-order-book/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestNumericFromFloat(t *testing.T) {
	numeric := utils.NumericFromFloat(500000.25)

	value, err := numeric.Float64Value()
	require.NoError(t, err)
	require.Equal(t, 500000.25, value.Float64)
	require.True(t, value.Valid)
}
