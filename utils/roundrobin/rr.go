// Package roundrobin implements a very simple roundrobin algorithm
package roundrobin

import "sync"

type RR struct {
	next int

	targets []interface{}

	sync.RWMutex
}

func New(targets []interface{}) *RR {
	return &RR{
		next:    0,
		targets: targets,
	}
}

func (rr *RR) Add(target interface{}) *RR {
	rr.Lock()
	defer rr.Unlock()

	rr.targets = append(rr.targets, target)

	return rr
}

func (rr *RR) Remove(target interface{}) *RR {
	rr.Lock()
	defer rr.Unlock()

	for i, t := range rr.targets {
		if t != target {
			continue
		}

		rr.targets = append(rr.targets[0:i], rr.targets[i+1:]...)
	}

	return rr
}

func (rr *RR) Len() int {
	return len(rr.targets)
}

func (rr *RR) Next() interface{} {
	rr.Lock()
	defer rr.Unlock()

	next := rr.next
	rr.next = (rr.next + 1) % len(rr.targets)

	return rr.targets[next]
}
