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
    c.send <- []byte(fmt.Sprintf("{\"type\": \"ping\", \"id\": %d}", c.id))
  }
}

var dialer = &websocket.Dialer{
  ReadBufferSize: 4096,
  WriteBufferSize: 4096,
}

func (c *Connection) dial(host string) {
  ws, _, err := dialer.Dial(host, nil)

  if err != nil {
    log.Fatalln("Cannot open a connection", err)
  }

  connections[c.id] = true
  log.Println("Connection #", c.id, " opened", len(connections))
  c.ws = ws
  go c.writer()
  go c.ping()
  c.reader()
}