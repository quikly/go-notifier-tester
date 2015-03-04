package main

import (
  "flag"
  "time"
)

var (
  numConn = flag.Int("connections", 0, "number of concurrent connections")
  numMsg = flag.Int("messages", 5, "Number of messages to be broadcasted")
  host = flag.String("host", "ws://localhost:8080/ws?poolId=deal:228", "websocket server address")
  done = make(chan bool)
  connections = make(map[int]bool)
)

func main() {
  flag.Parse()

  if *numConn > 0 {
    connections := make([]*Connection, *numConn)

    for i, c := range connections {
      c = &Connection{id: i, send: make(chan []byte, 256)}
      connections[i] = c
      go connections[i].dial(*host)
      time.Sleep(10 * time.Millisecond)
    }

    <-done
  }
}