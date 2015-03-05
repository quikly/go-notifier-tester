package main

import (
  "flag"
  "time"
  "log"
)

var (
  numConn = flag.Int("connections", 0, "number of concurrent connections")
  numMsg = flag.Int("messages", 5, "Number of messages to be broadcasted")
  host = flag.String("host", "ws://localhost:8080/ws?poolId=deal:228", "websocket server address")
  done = make(chan bool)
  connections = make(map[int]bool)
  msgStats = make(map[string]int)
  receivedMsgs = make(chan string)
)

func main() {
  flag.Parse()

  if *numConn > 0 {

    connections := make([]*Connection, *numConn)

    for i, c := range connections {
      c = &Connection{id: i, send: make(chan []byte, 256)}
      connections[i] = c
      go connections[i].dial(*host)
      time.Sleep(2 * time.Millisecond)
    }

    go func () {
      for {
        log.Println("\n\nReceived messages: ")
        for m, n := range msgStats {
          log.Println(m, n)
        }
        time.Sleep(1 * time.Second)
      }
    }();

    for {
      select {
      case m := <- receivedMsgs:
        if _, ok := msgStats[m]; ok {
          msgStats[m] += 1
        } else {
          msgStats[m] = 1
        }
      }
    }

  }
}