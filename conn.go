package main

import (
  "log"
  "fmt"
	"math/rand"
	"sync"
  "time"

  //"quikly/helper"

  "github.com/gorilla/websocket"
)

const MAX_DIAL_RETRIES = 5
const DIAL_RETRY_DELAY = 500 * time.Millisecond

type Connection struct {
  id int
  ws *websocket.Conn
  // Buffered channel of outbound messages
  send chan []byte
}

func (c *Connection) reader() {
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

func (c *Connection) writer() {
  for message := range c.send {
    err := c.ws.WriteMessage(websocket.TextMessage, message)
    if err != nil {
      log.Println("Cannot write message: ", err)
      break
    }
  }
  c.ws.Close()
}

func (c *Connection) ping() {
  for {
		delay := time.Duration(rand.Intn(15) + 5)
    time.Sleep(delay * time.Second)
    c.send <- []byte(fmt.Sprintf("{\"type\": \"ping\"}"))
  }
}

func (c *Connection) getWebSocket(host string) *websocket.Conn {
  retries := 0

	dialer := &websocket.Dialer{
	  ReadBufferSize: 1024,
	  WriteBufferSize: 1024,
	}

  for {
    ws, _, err := dialer.Dial(host, nil)

    if err == nil {
      return ws
    } else {
      if retries >= MAX_DIAL_RETRIES {
        log.Println(c.id, " - Cannot open a websocket connection: ", err)
        return nil
      } else {
				time.Sleep(DIAL_RETRY_DELAY)
        retries += 1
      }
    }
  }
}

func (c *Connection) dial(host string, wg *sync.WaitGroup) {
  if c.ws = c.getWebSocket(host); c.ws != nil {
		log.Println("Connection ID#", c.id, "opened.")
		wg.Done()

		connections.Lock()
    connections.conns[c] = true
		connections.Unlock()

    go c.writer()
    go c.ping()
    c.reader()
  } else {
		wg.Done()
  	log.Println("Could not establish connection.")
  }
}