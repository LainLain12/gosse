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
				fmt.Println("newclients", b.totalClients)

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
						// We can log this or implement more sophisticated backpressure.
						log.Printf("Client channel blocked, potentially slow consumer.")
						// Optionally, you could mark this client for disconnection if consistently slow.
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
	// r.Context().Done() is the most robust way to detect client disconnects.
	go func() {
		<-r.Context().Done() // This blocks until the client disconnects or request context is cancelled
		log.Printf("HTTP context done for client.")
		b.closedClients <- client // Signal the broker to remove this client
	}()

	// Keep connection alive and send messages
	for {
		select {
		case msg := <-client.MessageChannel:
			// Format and send the SSE message
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush() // Flush the data to the client immediately
		case <-client.Done:
			// Client was explicitly marked as done by the broker (e.g., due to disconnect)
			log.Printf("Client goroutine exiting due to Done signal.")
			return // Exit the handler goroutine
		}
	}
}

// StartBroadcastingTime continuously broadcasts the current time.
func (b *Broker) StartBroadcastingTime() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// currentTime := time.Now().Format(time.RFC3339)
			// message := fmt.Sprintf("Current time: %s", currentTime)

			data, _ := json.Marshal(liveDataStore)

			b.broadcaster <- string(data) // Send message to the broker for broadcasting
		case <-time.After(1 * time.Minute): // Example: Periodically check if clients exist
			if b.totalClients == 0 {
				log.Println("No active clients. Consider pausing broadcasts to save CPU.")
				// In a real application, you might stop this goroutine or reduce frequency
				// if there are no active listeners to save resources.
			}
		}
	}
}
