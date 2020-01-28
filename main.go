// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var hostFlag = flag.String("host", "www.quikly.localhost:5000", "http service address")
var numCon = flag.Int("numCon", 5, "max number of concurrent clients")
var schemeFlag = flag.String("scheme", "ws", "ws or wss (like http or https)")

const subscribeJSON = `{"command":"subscribe","identifier":"{\"channel\":\"GraphQLChannel\",\"channelId\":\"16fe8fc2f8e\"}"}`
const messageJSON = `{"command":"message","identifier":"{\"channel\":\"GraphQLChannel\",\"channelId\":\"16fe8fc2f8e\"}","data":"{\"query\":\"subscription getMessage {\\n  newMessage {\\n    message\\n    __typename\\n  }\\n}\\n\",\"variables\":{},\"operationName\":\"getMessage\",\"action\":\"execute\"}"}`

func main() {
	flag.Parse()
	log.SetFlags(0)

	var waitGroup sync.WaitGroup
	var max = *numCon

	if len(os.Getenv("NUM_CON")) > 0 {
		i, _ := strconv.ParseInt(os.Getenv("NUM_CON"), 10, 64)
		max = int(i)
	}

	for i := 0; i < max; i++ {
		go createClient(&waitGroup, i)
	}

	waitGroup.Wait()
}

func createClient(waitGroup *sync.WaitGroup, clientID int) {
	waitGroup.Add(1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := *hostFlag

	if len(os.Getenv("TARGET_HOST")) > 0 {
		host = os.Getenv("TARGET_HOST")
	}

	scheme := *schemeFlag
	if len(os.Getenv("SCHEME")) > 0 {
		scheme = os.Getenv("SCHEME")
	}

	u := url.URL{Scheme: scheme, Host: host, Path: "/cable"}
	log.Printf("client: %d connecting to %s", clientID, u.String())

	c, res, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err, "response: ", res)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("client: %d read: %v", clientID, err)
				return
			}
			log.Printf("client: %d recv: %v", clientID, message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	for {
		select {
		case <-done:
			waitGroup.Done()
			return
		case <-ticker.C:

			err := c.WriteMessage(websocket.TextMessage, []byte(subscribeJSON))
			if err != nil {
				log.Println("write:", err)
				return
			}

			err = c.WriteMessage(websocket.TextMessage, []byte(messageJSON))
			if err != nil {
				log.Println("write:", err)
				return
			}

			ticker.Stop()

		case <-interrupt:
			log.Printf("client: %d interrupt", clientID)

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("client: %d write close: %v", clientID, err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			waitGroup.Done()
			return
		}
	}
}
