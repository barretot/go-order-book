package utils_test

import (
	"testing"

	"github.com/barretot/go-order-book/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestMinFloat(t *testing.T) {
	assert.Equal(t, 1.5, utils.MinFloat(1.5, 2.5))
	assert.Equal(t, 1.5, utils.MinFloat(2.5, 1.5))
	assert.Equal(t, 2.0, utils.MinFloat(2.0, 2.0))
}
