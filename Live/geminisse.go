package Live

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client represents a single SSE client connection.
type Client struct {
	MessageChannel chan string   // Channel to send messages to this specific client
	Done           chan struct{} // Signal channel for client disconnection
}

// Broker manages all connected SSE clients and broadcasts messages.
type Broker struct {
	clients       map[*Client]bool // Registered clients
	newClients    chan *Client     // Channel for new client connections
	closedClients chan *Client     // Channel for disconnected clients
	broadcaster   chan string      // Channel to receive messages for broadcasting
	totalClients  int64            // Atomic counter for total active clients
	mu            sync.RWMutex     // Mutex to protect client map
}

// NewBroker creates and initializes a new Broker.
func NewBroker() *Broker {
	return &Broker{
		clients:       make(map[*Client]bool),
		newClients:    make(chan *Client),
		closedClients: make(chan *Client),
		broadcaster:   make(chan string, 100), // Buffered channel for messages
		totalClients:  0,
	}
}

// Start begins the Broker's main loop for managing clients and broadcasting messages.
func (b *Broker) Start() {
	go func() {
		for {
			select {
			case s := <-b.newClients:
				// A new client has connected
				b.mu.Lock()
				b.clients[s] = true
				b.totalClients++
				b.mu.Unlock()
				log.Printf("New client connected. Total clients: %d", b.totalClients)
				log.Printf("newclients: %d", b.totalClients)

			case s := <-b.closedClients:
				// A client has disconnected
				b.mu.Lock()
				delete(b.clients, s)
				close(s.MessageChannel) // Close the client's channel
				close(s.Done)           // Signal client's goroutine to stop
				b.totalClients--
				b.mu.Unlock()
				log.Printf("Client disconnected. Total clients: %d", b.totalClients)

			case msg := <-b.broadcaster:
				// Broadcast message to all active clients
				b.mu.RLock() // Use RLock for reading map to allow concurrent reads
				for client := range b.clients {
					select {
					case client.MessageChannel <- msg:
						// Message sent successfully
					case <-client.Done:
						// Client already signaled disconnection, will be cleaned up by closedClients handler
						log.Printf("Skipping dead client during broadcast.")
					default:
						// Client's channel is blocked, indicating a slow consumer.
						log.Printf("Client channel blocked, potentially slow consumer.")
					}
				}
				b.mu.RUnlock()
			}
		}
	}()
}

// SSEHandler is the HTTP handler for Server-Sent Events.
func (b *Broker) SSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set necessary headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For CORS

	// Ensure the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new client
	client := &Client{
		MessageChannel: make(chan string),
		Done:           make(chan struct{}),
	}

	// Register the new client with the broker
	b.newClients <- client

	// Listen for client disconnection (HTTP connection close)
	go func() {
		<-r.Context().Done()
		log.Printf("HTTP context done for client.")
		b.closedClients <- client
	}()

	// Send the latest live data immediately on connect
	data, _ := json.Marshal(liveDataStore)
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	flusher.Flush()

	// Keep connection alive and send messages, with ping-pong
	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()
	for {
		select {
		case msg := <-client.MessageChannel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-pingTicker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-client.Done:
			log.Printf("Client goroutine exiting due to Done signal.")
			return
		}
	}
}

// StartBroadcastingTime continuously broadcasts the current time.
func (b *Broker) StartBroadcastingTime() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var previousLive string

	for {
		select {
		case <-ticker.C:
			data, _ := json.Marshal(liveDataStore)
			currentLive := string(data)
			if currentLive != previousLive {
				b.broadcaster <- currentLive
				previousLive = currentLive
			}
		case <-time.After(1 * time.Minute):
			if b.totalClients == 0 {
				log.Printf("No active clients. Consider pausing broadcasts to save CPU.")
			}
		}
	}
}
