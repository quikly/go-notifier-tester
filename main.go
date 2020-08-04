package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	numConnections = flag.Int("connections", 1, "number of concurrent connections")
	maxConcurrent  = flag.Int("concurrency", 50, "max number of connections to dial concurrently")
	host           = flag.String("host", "www.quikly.localhost:5000", "websocket server address")
	dealHashid     = flag.String("dealHashid", "PVZhve", "The deal hashid")
	scheme         = flag.String("scheme", "ws", "ws or wss (like http or https)")
	requestOrigin  = flag.String("origin", "https://quikly.github.io", "request origin to set in header")
	apiKey         string
	connections    = &connManager{conns: make(map[*connection]bool)}
	done           = make(chan bool)
	msgLog         = make(chan string)
)

func main() {
	flag.Parse()

	log.SetFlags(0) // simplify log formatting

	if *numConnections < 1 {
		return
	}

	u := url.URL{Scheme: *scheme, Host: *host, Path: "/websocket"}

	apiKey = os.Getenv("API_KEY")

	var waitGroup sync.WaitGroup

	for i := 0; i < *numConnections; i++ {
		waitGroup.Add(1)
		c := &connection{id: i, send: make(chan []byte, 256)}

		go func() {
			c.dial(u.String(), &waitGroup)
		}()

		if (i+1)%*maxConcurrent == 0 {
			log.Println("Waiting", i)
			waitGroup.Wait()
		}
	}

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-done:
				waitGroup.Done()
				return
			case msg := <-msgLog:
				log.Println(msg)
			case <-ticker.C:
				log.Println("connections:", len(connections.conns))
			case <-interrupt:
				log.Println("Stopping...")
				connections.stop()
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				close(done)
				return
			}
		}
	}()

	<-done
}
