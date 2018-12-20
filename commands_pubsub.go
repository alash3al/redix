package main

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/redcon"
)

// publishCommand - PUBLISH <channel> <paylaod>
func publishCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("PUBLISH command must have at least 1 argument, PUBLISH <channel> <message>")
		return
	}

	changelog.Broadcast(c.args[1], c.args[0])

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
	defer changelog.Detach(subscriber)

	conn := c.Detach()
	defer conn.Close()

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
