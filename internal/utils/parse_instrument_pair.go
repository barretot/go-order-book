package utils

import (
	"strings"

	"github.com/barretot/go-order-book/internal/domain/models"
)

func ParseInstrumentPair(instrument models.Instrument) (models.Instrument, models.Instrument) {
	base, quote, found := strings.Cut(string(instrument), "/")
	if !found {
		return instrument, ""
	}

	return models.Instrument(base), models.Instrument(quote)
}
