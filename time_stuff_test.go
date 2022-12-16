package main

import (
	"fmt"
	"testing"
	"time"
)

/*
timestamp":1671133371657,
"event_ts":"2022-12-15T12:42:51.663327"
*/

func tsStringToTime(in string) (*time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000000", in)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &t, nil
}

func TestTsStringToX(t *testing.T) {
	input := "2022-12-15T12:42:51.663327"
	output, err := tsStringToX(input)
	if err != nil {
		t.Error("Did not parse")
	}
	if output.Year() != 2022 || output.Month() != 12 || output.Day() != 15 {
		t.Error("date did not match")
	}
	if output.Hour() != 12 || output.Minute() != 42 || output.Minute() != 51 {
		t.Error("Time did not match")
	}
}
