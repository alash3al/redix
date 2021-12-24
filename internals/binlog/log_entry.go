package binlog

import "fmt"

// LogEntry represents a log entry
type LogEntry struct {
	TimeNS  int64
	ID      int64
	Action  string
	Payload interface{}
}

// Key returns the primary key of the entry
func (l *LogEntry) Key() string {
	return fmt.Sprintf("%d-%d", l.TimeNS, l.ID)
}
