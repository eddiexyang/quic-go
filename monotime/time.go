package monotime

import (
	"time"

	"github.com/eddiexyang/quic-go/internal/monotime"
)

type Time = monotime.Time

func Now() Time                    { return monotime.Now() }
func Since(t Time) time.Duration   { return monotime.Since(t) }
func Until(t Time) time.Duration   { return monotime.Until(t) }
func FromTime(t time.Time) Time    { return monotime.FromTime(t) }
