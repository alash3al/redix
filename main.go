package main

import (
	"fmt"

	"github.com/alash3al/go-color"
)

func main() {
	fmt.Printf("⇨ redix resp server available at: %s \n", color.GreenString(*flagRESPListenAddr))
	fmt.Printf("⇨ redix http server available at: %s \n", color.GreenString(*flagHTTPListenAddr))

	err := make(chan error)

	go (func() {
		err <- initRespServer()
	})()

	go (func() {
		err <- initHTTPServer()
	})()

	if err := <-err; err != nil {
		color.Red(err.Error())
	}
}
