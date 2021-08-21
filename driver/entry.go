package driver

import (
	"strconv"
	"time"
)

type EntryValueType string

const (
	EntryValueTypeRaw    = "raw"
	EntryValueTypeNumber = "number"
)

type Entry struct {
	Key   string
	Value []byte

	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (v Entry) Float() float64 {
	fval, _ := strconv.ParseFloat(string(v.Value), 64)

	return fval
}
