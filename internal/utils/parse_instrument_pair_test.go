package utils_test

import (
	"testing"

	"github.com/barretot/go-order-book/internal/domain/models"
	"github.com/barretot/go-order-book/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestParseInstrumentPair(t *testing.T) {
	base, quote := utils.ParseInstrumentPair(models.Instrument("BTC/BRL"))

	assert.Equal(t, models.Instrument("BTC"), base)
	assert.Equal(t, models.Instrument("BRL"), quote)
}

func TestParseInstrumentPairWithoutQuote(t *testing.T) {
	base, quote := utils.ParseInstrumentPair(models.Instrument("BTC"))

	assert.Equal(t, models.Instrument("BTC"), base)
	assert.Empty(t, quote)
}
