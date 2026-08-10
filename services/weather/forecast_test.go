package weather

import (
	"testing"
	"time"
)

func TestFrenchDayName(t *testing.T) {
	t.Parallel()
	if got := frenchDayName("2026-08-10", mustLoc(t)); got != "lundi" {
		t.Fatalf("got %q", got)
	}
	if got := frenchDayName("2026-08-11", mustLoc(t)); got != "mardi" {
		t.Fatalf("got %q", got)
	}
}

func TestParseOpenMeteoLocalTime(t *testing.T) {
	t.Parallel()
	loc := mustLoc(t)
	got, err := parseOpenMeteoLocalTime("2026-08-11T06:42", loc)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 6 || got.Minute() != 42 {
		t.Fatalf("got %v", got)
	}
}

func mustLoc(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}
