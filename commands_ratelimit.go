// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"strconv"
	"strings"
	"time"
)

// ratelimitsetCommand - RATELIMITSET <bucket> <limit> <seconds>
func ratelimitsetCommand(c Context) {
	if len(c.args) < 3 {
		c.WriteError("RATELMITSET command requires at least 3 arguments: RATELIMITSET <bucket> <limit> <seconds>")
		return
	}

	bucket, limit, seconds := c.args[0], c.args[1], c.args[2]

	if err := c.db.MSet(map[string]string{
		"/{RATELIMITBUCKET}/" + bucket: limit + ";" + seconds,
		"/{RATELIMITSTAT}/" + bucket + "/" + strconv.FormatInt(time.Now().Unix(), 10): "0",
	}); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(1)
}

// ratelimittakeCommand - RATELIMITTAKE <bucket>
func ratelimittakeCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("RATELIMITTAKE command requires at least 1 argument: RATELIMITTAKE <bucket>")
		return
	}

	bucket := c.args[0]

	meta, err := c.db.Get("/{RATELIMITBUCKET}/" + bucket)
	if err != nil {
		c.WriteInt(-1)
		return
	}

	parts := strings.SplitN(meta, ";", 2)
	if len(parts) < 2 {
		c.db.Del([]string{"/{RATELIMITBUCKET}/" + bucket})
		c.WriteInt(-1)
		return
	}

	limit, _ := strconv.Atoi(parts[0])
	seconds, _ := strconv.Atoi(parts[1])
	now := int(time.Now().Unix())

	if limit < 1 {
		c.WriteInt(-1)
		return
	}

	key := "/{RATELIMITSTAT}/" + bucket + "/" + strconv.Itoa(now/seconds)
	reachedVal, _ := c.db.Get("/{RATELIMITSTAT}/" + bucket + "/" + strconv.Itoa(now/seconds))
	reachedInt, _ := strconv.Atoi(reachedVal)

	if reachedInt >= limit {
		c.WriteInt(0)
		return
	}

	val, err := c.db.Incr(key, 1)
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt64(int64(limit) - val)
}

// ratelimitgetCommand - RATELIMITGET <bucket>
func ratelimitgetCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("RATELIMITGET command requires at least 1 argument: RATELIMITGET <bucket>")
		return
	}

	bucket := c.args[0]

	val, err := c.db.Get("/{RATELIMITBUCKET}/" + bucket)
	if err != nil {
		c.WriteInt(-1)
		return
	}

	parts := strings.SplitN(val, ";", 2)
	limit, _ := strconv.Atoi(parts[0])
	seconds, _ := strconv.Atoi(parts[1])
	now := time.Now().Unix()
	val, _ = c.db.Get("/{RATELIMITSTAT}/" + bucket + "/" + strconv.Itoa(int(now/int64(seconds))))
	reached, _ := strconv.Atoi(val)

	c.WriteArray(4)
	c.WriteInt(limit)
	c.WriteInt(seconds)
	c.WriteInt64(now / int64(seconds))
	c.WriteInt(reached)
}
