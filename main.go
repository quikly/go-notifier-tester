package main

import (
  "flag"
  "time"
  "log"
	"sync"
)

const MAX_CONCURRENT_DIAL = 50

type ConnManager struct {
	sync.Mutex
	conns map[*Connection]bool
}

var (
  numConn = flag.Int("connections", 0, "number of concurrent connections")
  numMsg = flag.Int("messages", 5, "Number of messages to be broadcasted")
  host = flag.String("host", "ws://localhost:8080/ws?channelId=deal:228", "websocket server address")
  done = make(chan bool)
	connections = &ConnManager{conns: make(map[*Connection]bool)}
  msgStats = make(map[string]int)
  receivedMsgs = make(chan string)
)

func main() {
  flag.Parse()

  if *numConn > 0 {

		var wg sync.WaitGroup

    for i := 0; i < *numConn; i++ {
			wg.Add(1)
      c := &Connection{id: i, send: make(chan []byte, 256)}

      go func() {
				c.dial(*host, &wg)
			}()

			if (i+1)%MAX_CONCURRENT_DIAL == 0 {
				wg.Wait()
			}
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