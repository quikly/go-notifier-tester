package main

import (
  "log"
  "time"
  "fmt"

  //"quikly/helper"

  "github.com/gorilla/websocket"
)

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
      break
    }
  }
  c.ws.Close()
}

func (c *Connection) ping() {
  for {
    time.Sleep(25 * time.Second)
    c.send <- []byte(fmt.Sprintf("{\"type\": \"ping\"}"))
  }
}

var dialer = &websocket.Dialer{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

func (c *Connection) getWebSocket(host string) *websocket.Conn {
  retries := 0

  for {
    ws, _, err := dialer.Dial(host, nil)

    if err == nil {
      return ws
    } else {
      if retries >= 3 {
        log.Fatalln("Cannot open a websocket connection: ", err)
      } else {
        retries += 1
      }
    }
  }
}

func (c *Connection) dial(host string) {
  ws := c.getWebSocket(host)
  connections[c.id] = true
  //log.Println("Connection #", c.id, " opened", len(connections))
  c.ws = ws
  go c.writer()
  go c.ping()
  c.reader()
}