package container

import (
	"fmt"
	"github.com/alash3al/redix/pkg/db/driver"
	"strconv"
)

func (c Container) Set(k, v []byte, ttl int) error {
	return c.db.Put(driver.Entry{
		Key:   k,
		Value: v,
		TTL:   ttl,
	})
}

func (c Container) Get(k []byte) ([]byte, error) {
	return c.db.Get(k)
}

func (c Container) Del(k []byte) error {
	return c.db.Put(driver.Entry{
		Key: k,
	})
}

func (c Container) Incr(k []byte, delta float64, ttl int) error {
	return c.db.Put(driver.Entry{
		Key:   k,
		Value: []byte(fmt.Sprintf("%f", delta)),
		TTL:   ttl,
		WriteMerger: func(oldValue []byte, newValue []byte) []byte {
			c, _ := strconv.ParseFloat(string(oldValue), 64)
			delta, _ := strconv.ParseFloat(string(newValue), 64)

			c += delta

			return []byte(fmt.Sprintf("%f", c))
		},
	})
}
