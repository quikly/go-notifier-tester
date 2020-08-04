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

type connManager struct {
	sync.Mutex
	conns map[*connection]bool
}

func (cMgr *connManager) stop() {
	for conn := range cMgr.conns {
		conn.close()
	}
}

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
		msgString := fmt.Sprintf("%v:     <- %s", c.id, string(message[:]))
		msgLog <- msgString
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		msgString := fmt.Sprintf("%v: -> %s", c.id, message)
		msgLog <- msgString
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Cannot write message: ", err)
			break
		}
	}
	c.ws.Close()
}

func (c *connection) subscribeToGraphQL() {
	channelID := strconv.FormatInt(time.Now().Unix()+rand.Int63n(100000), 16)
	c.send <- []byte(fmt.Sprintf("{\"command\":\"subscribe\",\"identifier\":\"{\\\"channel\\\":\\\"GraphQLChannel\\\",\\\"channelId\\\":\\\"%s\\\"}\"}", channelID))

	command := `{"command":"message","identifier":"{\"channel\":\"GraphQLChannel\",\"channelId\":\"%s\"}","data":"{\"query\":\"subscription claimedForDeal($dealHashid: String!) {\\n  claimedForDeal(dealHashid: $dealHashid) {\\n    quikly {\\n      id\\n      incentiveTiers {\\n        id\\n        rank\\n        description\\n        upperResponseTime\\n        quantity\\n        claimCount\\n        __typename\\n      }\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\",\"variables\":{\"dealHashid\":\"%v\"},\"operationName\":\"claimedForDeal\",\"action\":\"execute\"}"}`
	c.send <- []byte(fmt.Sprintf(command, channelID, *dealHashid))

	// log.Println(fmt.Sprintf("%v: channelID=%s dealHashid=%v", c.id, channelID, *dealHashid))
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
	headers.Set("Origin", *requestOrigin)

	if apiKey != "" {
		headers.Set("APIKey", apiKey)
	}

	for {
		ws, _, err := dialer.Dial(host, headers)

		if err == nil {
			return ws
		}

		if retries >= maxDialRetries {
			log.Println(c.id, ": Cannot open a websocket connection: ", err)
			return nil
		}

		time.Sleep(dialRetryDelay)
		retries++

	}
}

func (c *connection) close() {
	err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("%d: write close: %v", c.id, err)
		return
	}
}

func (c *connection) dial(host string, waitGroup *sync.WaitGroup) {
	log.Printf("%v: connecting to %s", c.id, host)
	if c.ws = c.getWebSocket(host); c.ws != nil {
		log.Printf("%v: connected", c.id)
		waitGroup.Done()

		connections.Lock()
		connections.conns[c] = true
		connections.Unlock()

		// c.writer() reads from a 'send' channel and
		// transmits it through the websocket connection
		go c.writer()

		c.subscribeToGraphQL()

		// go c.ping()

		// c.reader() start a loop that listens for messages and
		// sends them to the receivedMsgs channel
		c.reader()
	} else {
		waitGroup.Done()
		log.Printf("%v: Could not establish connection", c.id)
	}
}
