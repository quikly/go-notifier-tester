package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const maxDialRetries = 5
const dialRetryDelay = 500 * time.Millisecond

type connection struct {
	id int
	ws *websocket.Conn
	// Buffered channel of outbound messages
	send chan []byte
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("Cannot read message: ", err)
			break
		}
		msgString := string(message[:])
		receivedMsgs <- msgString
		//log.Println("connection# ", c.id, " - Message: ", msgString)
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Cannot write message: ", err)
			break
		}
	}
	c.ws.Close()
}

func (c *connection) subscribe() {
	channelID := strconv.FormatInt(time.Now().Unix()+rand.Int63n(100000), 16)
	c.send <- []byte(fmt.Sprintf("{\"command\":\"subscribe\",\"identifier\":\"{\\\"channel\\\":\\\"GraphQLChannel\\\",\\\"channelId\\\":\\\"%s\\\"}\"}", channelID))
	command := `{"command":"message","identifier":"{\"channel\":\"GraphQLChannel\",\"channelId\":\"%s\"}","data":"{\"query\":\"subscription claimedForDeal($dealHashid: String!) {\\n  claimedForDeal(dealHashid: $dealHashid) {\\n    quikly {\\n      id\\n      incentiveTiers {\\n        id\\n        rank\\n        description\\n        upperResponseTime\\n        quantity\\n        claimCount\\n        __typename\\n      }\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\",\"variables\":{\"dealHashid\":\"%v\"},\"operationName\":\"claimedForDeal\",\"action\":\"execute\"}"}`
	log.Println(fmt.Sprintf("Here: %s and %v", channelID, *dealHashid))
	c.send <- []byte(fmt.Sprintf(command, channelID, *dealHashid))

}

func (c *connection) ping() {
	for {
		delay := time.Duration(rand.Intn(15) + 5)
		time.Sleep(delay * time.Second)
		c.send <- []byte(fmt.Sprintf("{\"type\": \"ping\"}"))
	}
}

func (c *connection) getWebSocket(host string) *websocket.Conn {
	retries := 0

	dialer := &websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	headers := http.Header{}
	headers.Set("Origin", "https://quikly.github.io")

	for {
		ws, _, err := dialer.Dial(host, headers)

		if err == nil {
			return ws
		}

		if retries >= maxDialRetries {
			log.Println(c.id, " - Cannot open a websocket connection: ", err)
			return nil
		}

		time.Sleep(dialRetryDelay)
		retries++

	}
}

func (c *connection) dial(host string, wg *sync.WaitGroup) {
	log.Printf("connecting to %s", host)
	if c.ws = c.getWebSocket(host); c.ws != nil {
		log.Println("Connection ID#", c.id, "opened.")
		wg.Done()

		connections.Lock()
		connections.conns[c] = true
		connections.Unlock()

		go c.writer()
		//go c.ping()
		c.subscribe()
		c.reader()
	} else {
		wg.Done()
		log.Println("Could not establish connection.")
	}
}
