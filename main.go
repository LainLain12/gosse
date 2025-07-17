// CpuinfoRoute streams live CPU and memory usage using SSE

// TimeOnlyRoute returns only the current time as plain text

// SSE handler for time and client count only

package main

import (
	"fmt"
	"gosse/Live"
	"gosse/gift"
	"gosse/threedata"
	"gosse/twoddata"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func CpuinfoRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	notify := w.(http.CloseNotifier).CloseNotify()
	for {
		select {
		case <-time.After(1 * time.Second):
			cpuPercent, _ := cpu.Percent(0, false)
			memStat, _ := mem.VirtualMemory()
			msg := fmt.Sprintf("CPU: %.1f%% | MEM: %.1f%%", func() float64 {
				if len(cpuPercent) > 0 {
					return cpuPercent[0]
				} else {
					return 0
				}
			}(), memStat.UsedPercent)
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-notify:
			return
		}
	}
}

type clientChan chan string

type Broker struct {
	clients map[clientChan]struct{}
	lock    sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[clientChan]struct{}),
	}
}

func (b *Broker) AddClient(c clientChan) {
	b.lock.Lock()
	b.clients[c] = struct{}{}
	b.lock.Unlock()
}

func (b *Broker) RemoveClient(c clientChan) {
	b.lock.Lock()
	delete(b.clients, c)
	b.lock.Unlock()
}

func sseTimeHandler(b *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		client := make(clientChan, 10)
		notify := w.(http.CloseNotifier).CloseNotify()
		b.AddClient(client)
		defer b.RemoveClient(client)

		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()

		for {
			select {
			case msg := <-client:
				// Show time and actual client count from broker
				var timeStr string
				n, _ := fmt.Sscanf(msg, "%s | CPU: %*f%% | MEM: %*f%% | clients: %*s", &timeStr)
				b.lock.RLock()
				clientCount := len(b.clients)
				b.lock.RUnlock()
				if n == 1 {
					fmt.Fprintf(w, "data: %s | clients: %d\n\n", timeStr, clientCount)
				} else {
					fmt.Fprintf(w, "data: %s | clients: %d\n\n", msg, clientCount)
				}
				flusher.Flush()
			case <-notify:
				return
			}
		}
	}
}

func (b *Broker) Broadcast(msg string) {
	b.lock.RLock()
	clientCount := len(b.clients)
	for c := range b.clients {
		select {
		case c <- fmt.Sprintf("%s | clients: %d", msg, clientCount):
		default:
			// If client is slow or gone, skip
		}
	}
	b.lock.RUnlock()
}

func sseHandler(b *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		client := make(clientChan, 10)
		notify := w.(http.CloseNotifier).CloseNotify()
		// Add client to broker
		b.AddClient(client)
		defer b.RemoveClient(client)

		// Notify client of connection
		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()

		for {
			select {
			case msg := <-client:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-notify:
				return
			}
		}
	}
}

// LiveRoute streams live time, CPU, memory, and client count as SSE or shows it in browser
func LiveRoute(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "text/event-stream" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
				return
			}

			// Track this /live client in broker
			client := make(clientChan, 1)
			broker.AddClient(client)
			defer broker.RemoveClient(client)

			notify := w.(http.CloseNotifier).CloseNotify()
			go func() {
				for {
					select {
					case <-time.After(1 * time.Second):
						cpuPercent, _ := cpu.Percent(0, false)
						memStat, _ := mem.VirtualMemory()
						broker.lock.RLock()
						clientCount := len(broker.clients)
						broker.lock.RUnlock()
						msg := fmt.Sprintf("%s | CPU: %.1f%% | MEM: %.1f%% | clients: %d", time.Now().Format(time.RFC3339),
							func() float64 {
								if len(cpuPercent) > 0 {
									return cpuPercent[0]
								} else {
									return 0
								}
							}(),
							memStat.UsedPercent,
							clientCount,
						)
						fmt.Fprintf(w, "data: %s\n\n", msg)
						flusher.Flush()
					case <-notify:
						return
					}
				}
			}()
			// Block main handler until client disconnects
			<-notify
		} else {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><h1>Live Time, CPU, Memory, Clients</h1>
			<pre id='live'></pre>
			<script>
			var es = new EventSource('/live');
			es.onmessage = function(e) { document.getElementById('live').textContent = e.data; };
			</script></body></html>`)
		}
	}
}

func main() {
	// Route to show all twoddata rows as JSON
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			debug.PrintStack()
		}
	}()
	// Initialize SQLite DB and table

	db := twoddata.InitDB("twoddata.db")
	giftDB := gift.InitGiftDB("twoddata.db")
	threedDB := threedata.InitThreedDB("twoddata.db")
	defer db.Close()
	defer giftDB.Close()

	broker := NewBroker()

	// Start goroutine to broadcast time/cpu/mem/client count every second
	go func() {
		for {
			cpuPercent, _ := cpu.Percent(0, false)
			memStat, _ := mem.VirtualMemory()
			msg := fmt.Sprintf("%s | CPU: %.1f%% | MEM: %.1f%%", time.Now().Format(time.RFC3339),
				func() float64 {
					if len(cpuPercent) > 0 {
						return cpuPercent[0]
					} else {
						return 0
					}
				}(),
				memStat.UsedPercent,
			)
			broker.Broadcast(msg)
			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/events", sseHandler(broker))
	http.HandleFunc("/time", sseTimeHandler(broker))
	http.HandleFunc("/", CpuinfoRoute)
	http.HandleFunc("/timecpu", LiveRoute(broker))
	http.HandleFunc("/history", Live.TwoddataHandler(db))
	http.HandleFunc("/addlive", Live.AddLiveDataHandler)
	http.HandleFunc("/live", Live.LiveDataPageHandler)
	http.HandleFunc("/livedata/sse", Live.LiveDataSSEHandler)
	http.HandleFunc("/threed", threedata.ThreedDataHandler(threedDB))
	http.HandleFunc("/gift", gift.GiftDataHandler(giftDB))
	http.HandleFunc("/addimage/", gift.AddImageHandler(giftDB))
	// Serve static images from /images/
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	log.Println("SSE server started on :4597")
	log.Fatal(http.ListenAndServe(":4597", nil))
}
