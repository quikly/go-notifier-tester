package main

import (
	"flag"
	"log"
	"net/url"
	"sync"
	"time"
)

const maxConcurrentDial = 50

type connManager struct {
	sync.Mutex
	conns map[*connection]bool
}

var (
	numConn      = flag.Int("connections", 0, "number of concurrent connections")
	host         = flag.String("host", "www.quikly.localhost:5000", "websocket server address")
	dealHashid   = flag.String("dealHashid", "PVZhve", "The deal hashid")
	done         = make(chan bool)
	connections  = &connManager{conns: make(map[*connection]bool)}
	msgStats     = make(map[string]int64)
	receivedMsgs = make(chan string)
)

func main() {
	flag.Parse()

	if !(*numConn > 0) {
		return
	}

	u := url.URL{Scheme: "wss", Host: *host, Path: "/websocket"}

	var wg sync.WaitGroup

	for i := 0; i < *numConn; i++ {
		wg.Add(1)
		c := &connection{id: i, send: make(chan []byte, 256)}

		go func() {
			c.dial(u.String(), &wg)
		}()

		if (i+1)%maxConcurrentDial == 0 {
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
			msgStats[m]++
		} else {
			msgStats[m] = 1
		}
	}
}
