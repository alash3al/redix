package leveldb

import (
	"github.com/alash3al/redix/pkg/db/driver"
	"github.com/vmihailenco/msgpack/v4"
	"time"
)

// Value represents a value with expiration date
type Value struct {
	Value   []byte
	Expires *time.Time
}

// Bytes encodes the value to a byte array
func (e Value) Bytes() []byte {
	b, _ := msgpack.Marshal(e)
	return b
}

// IsExpired whether the value is expired or not
func (e Value) IsExpired() bool {
	expires := *(e.Expires)
	return time.Now().After(expires) || time.Now().Equal(expires)
}

// EntryToValue build a value from entry representation
func EntryToValue(e driver.Entry) Value {
	val := Value{
		Value:   e.Value,
		Expires: nil,
	}

	if e.TTL > 0 {
		expires := time.Now().Add(time.Duration(e.TTL))
		val.Expires = &expires
	}

	return val
}

// BytesToValue Decodes the specified byte array to Value
func BytesToValue(b []byte) (v Value) {
	msgpack.Unmarshal(b, &v)
	return
}
