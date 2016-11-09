package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"
)

const MAX_CONCURRENT_DIAL = 50

type ConnManager struct {
	sync.Mutex
	conns map[*Connection]bool
}

var (
	numConn      = flag.Int("connections", 0, "number of concurrent connections")
	numMsg       = flag.Int("messages", 5, "Number of messages to be broadcasted")
	host         = flag.String("host", "ws://localhost:8081/ws", "websocket server address")
	deal         = flag.String("deal", "10107", "deal id")
	done         = make(chan bool)
	connections  = &ConnManager{conns: make(map[*Connection]bool)}
	msgStats     = make(map[string]int64)
	receivedMsgs = make(chan string)
)

func main() {
	flag.Parse()
	dialAddr := strings.Join([]string{*host, "?channelId=deal:", *deal}, "")

	if !(*numConn > 0) {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < *numConn; i++ {
		wg.Add(1)
		c := &Connection{id: i, send: make(chan []byte, 256)}

		go func() {
			c.dial(dialAddr, &wg)
		}()

		if (i+1)%MAX_CONCURRENT_DIAL == 0 {
			wg.Wait()
		}
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			log.Println("\n\nReceived messages: ")
			for m, n := range msgStats {
				log.Println(m, n)
			}
			<-ticker.C
		}
	}()

	for {
		m := <-receivedMsgs
		if _, ok := msgStats[m]; ok {
			msgStats[m] += 1
		} else {
			msgStats[m] = 1
		}
	}
}
