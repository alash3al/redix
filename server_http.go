// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"net/http"

	"github.com/alash3al/go-pubsub"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func initHTTPServer() error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Recover())

	if *flagVerbose {
		e.Use(middleware.Logger())
	}

	upgrader := websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, "PONG ;)")
	})

	e.GET("/stream/ws/:userID", func(c echo.Context) error {
		subscriber := (func() *pubsub.Subscriber {
			sub, found := websockets.Load(c.Param("userID"))
			if !found {
				return nil
			}

			return sub.(*pubsub.Subscriber)
		})()

		if nil == subscriber {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   "websocket subscriber not found",
			})
		}

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		defer ws.Close()

		running := true

		ws.SetCloseHandler(func(_ int, _ string) error {
			running = false
			return nil
		})

		for msg := range subscriber.GetMessages() {
			if !running {
				break
			}
			if ws.WriteJSON(map[string]interface{}{
				"channel": msg.GetTopic(),
				"payload": msg.GetPayload(),
				"time":    msg.GetCreatedAt(),
			}) != nil {
				break
			}
		}

		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": "closed",
		})
	})

	return e.Start(*flagHTTPListenAddr)
}
