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
	Key  string
	Type EntryValueType

	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type RawValueEntry struct {
	Entry

	Value string
}

func (v RawValueEntry) Float() float64 {
	fval, _ := strconv.ParseFloat(v.Value, 64)

	return fval
}

type CounterEntry struct {
	Entry

	IncrementBy float64
}
