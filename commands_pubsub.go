// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"encoding/json"
	"strings"

	"github.com/alash3al/go-pubsub"
	"github.com/tidwall/redcon"
	"gopkg.in/resty.v1"
)

// publishCommand - PUBLISH <channel> <paylaod>
func publishCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("PUBLISH command must have at least one argument: PUBLISH <channel> <message>")
		return
	}

	if changelog.Subscribers(c.args[0]) > 0 {
		changelog.Broadcast(c.args[1], c.args[0])
	}

	c.WriteInt(1)
}

// subscribeCommand - SUBSCRIBE <channel>
func subscribeCommand(c Context) {
	topics := c.args

	if len(topics) < 1 {
		topics = []string{defaultPubSubAllTopic}
	}

	subscriber, err := changelog.Attach()
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	conn := c.Detach()
	defer conn.Close()
	defer changelog.Detach(subscriber)

	go (func() {
		for {
			_, err := conn.ReadCommand()
			if err != nil {
				break
			}
			conn.NetConn().Write(redcon.AppendOK(nil))
		}
	})()

	d := redcon.AppendArray(nil, 3)
	d = redcon.AppendBulkString(d, "subscribe")
	d = redcon.AppendBulkString(d, strings.Join(topics, ", "))
	d = redcon.AppendInt(d, 1)

	conn.NetConn().Write(d)

	changelog.Subscribe(subscriber, topics...)

	msgsChan := subscriber.GetMessages()
	for msg := range msgsChan {
		data, ok := msg.GetPayload().(string)
		if !ok {
			d, _ := json.Marshal(msg.GetPayload())
			data = string(d)
		}

		d := redcon.AppendArray(nil, 3)

		d = redcon.AppendBulkString(d, "message")
		d = redcon.AppendBulkString(d, msg.GetTopic())
		d = redcon.AppendBulkString(d, data)

		if _, err := conn.NetConn().Write(d); nil != err {
			break
		}
	}
}

// webhooksetCommand - WEBHOOKSET <channel> <http://url.here/>
func webhooksetCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("WEBHOOKSET command requires at least two arguments: WEBHOOKSET <channel> <http://url.here/>")
		return
	}

	channel, url := c.args[0], c.args[1]

	user, err := changelog.Attach()
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	changelog.Subscribe(user, channel)
	webhooks.Store(user.GetID(), make(chan bool))

	go (func() {
		done := (func() chan bool {
			ch, _ := webhooks.Load(user.GetID())
			return ch.(chan bool)
		})()

	eventloop:
		for {
			select {
			case msg := <-user.GetMessages():
				resty.R().SetHeader("Content-Type", "application/json").SetBody(map[string]interface{}{
					"channel": msg.GetTopic(),
					"payload": msg.GetPayload(),
					"time":    msg.GetCreatedAt(),
				}).Post(url)
			case <-done:
				changelog.Detach(user)
				close(done)
				break eventloop
			}
		}
	})()

	c.WriteString(user.GetID())
}

// webhookdel - WEBHOOKDEL <channel> <http://url.here/>
func webhookdelCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("WEBHOOKDEL command requires at least one argument: WEBHOOKDEL <webhook-id>")
		return
	}

	webhook, found := webhooks.Load(c.args[0])
	if !found {
		c.WriteInt(1)
		return
	}

	webhooks.Delete(c.args[0])
	webhookChan := webhook.(chan bool)
	webhookChan <- true

	c.WriteInt(1)
}

// websocketopenCommand - WEBSOCKETOPEN <channel>
func websocketopenCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("WEBSOCKETOPEN command requires at least one argument: WEBSOCKETOPEN <channel>")
		return
	}

	channel := c.args[0]

	user, err := changelog.Attach()
	if err != nil {
		c.WriteError(":: " + err.Error())
		return
	}

	changelog.Subscribe(user, channel)
	websockets.Store(user.GetID(), user)

	c.WriteString(user.GetID())
}

// websocketcloseCommand - WEBSOCKETCLOSE <ID>
func websocketcloseCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("WEBSOCKETCLOSE command requires at least one argument: WEBSOCKETCLOSE <id>")
		return
	}

	user, found := websockets.Load(c.args[0])
	if !found {
		c.WriteInt(1)
		return
	}

	websockets.Delete(c.args[0])
	changelog.Detach(user.(*pubsub.Subscriber))

	c.WriteInt(1)
}
