package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type NextDate struct {
	date   string
	repeat string
	want   string
}

func TestNextDate(t *testing.T) {
	tbl := []NextDate{
		{"20240126", "", ""},
		{"20240126", "k 34", ""},
		{"20240126", "ooops", ""},
		{"15000156", "y", ""},
		{"ooops", "y", ""},
		{"16890220", "y", `20240220`},
		{"20250701", "y", `20260701`},
		{"20240101", "y", `20250101`},
		{"20231231", "y", `20241231`},
		{"20240229", "y", `20250301`},
		{"20240301", "y", `20250301`},
		{"20240113", "d", ""},
		{"20240113", "d 7", `20240127`},
		{"20240120", "d 20", `20240209`},
		{"20240202", "d 30", `20240303`},
		{"20240320", "d 401", ""},
		{"20231225", "d 12", `20240130`},
		{"20240228", "d 1", "20240229"},
	}

	for _, testLine := range tbl {
		got, _ := nextDate(time.Now(), testLine.date, testLine.repeat)
		assert.Equal(t, testLine.want, got)
	}
}
