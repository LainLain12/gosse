// gosse/Live/websocket_broker.go
package Live

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic" // For atomic operations on totalClients
	"time"

	"github.com/gorilla/websocket" // Make sure to 'go get github.com/gorilla/websocket'
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development. In production, restrict this.
		return true
	},
}

// WebSocketClient represents a single WebSocket connection.
type WebSocketClient struct {
	broker *WebSocketBroker // Reference to the broker managing this client
	conn   *websocket.Conn  // The WebSocket connection
	send   chan []byte      // Buffered channel for outbound messages
	done   chan struct{}    // Signal channel for client goroutine shutdown
}

// readPump pumps messages from the WebSocket connection to the broker.
// It also handles incoming pings/pongs and sets read deadlines.
func (c *WebSocketClient) readPump() {
	defer func() {
		// Signal broker that this client is done
		c.broker.unregister <- c
		c.conn.Close()
		close(c.done) // Ensure done channel is closed when readPump exits
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		// Reset read deadline on pong
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error for client %p: %v", c, err)
			}
			break // Exit loop on error or close
		}
		// For this example, we don't process incoming client messages,
		// but you would handle them here (e.g., send to another channel for processing)
		log.Printf("Received message from client %p: %s", c, message)
	}
}

// writePump pumps messages from the broker to the WebSocket connection.
// It also handles sending pings.
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The broker closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting next writer for client %p: %v", c, err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current WebSocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for client %p: %v", c, err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error for client %p: %v", c, err)
				return
			}
		case <-c.done:
			// Signal to stop writing for this client
			log.Printf("writePump for client %p exiting due to done signal.", c)
			return
		}
	}
}

// WebSocketBroker manages WebSocket clients.
type WebSocketBroker struct {
	clients      *sync.Map             // Registered clients: map[*WebSocketClient]bool
	register     chan *WebSocketClient // Channel for new client connections
	unregister   chan *WebSocketClient // Channel for disconnected clients
	broadcast    chan []byte           // Channel to receive messages for broadcasting
	totalClients atomic.Int64          // Atomic counter for total active clients
}

// NewWebSocketBroker creates and initializes a new WebSocketBroker.
func NewWebSocketBroker() *WebSocketBroker {
	return &WebSocketBroker{
		clients:    &sync.Map{},
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan []byte, 256), // Buffered channel for broadcast messages
	}
}

// Start begins the WebSocketBroker's main loop for managing clients and broadcasting messages.
func (b *WebSocketBroker) Start() {
	a := 0
	go func() {
		for {
			select {
			case client := <-b.register:
				// A new client has connected
				b.clients.Store(client, true)
				b.totalClients.Add(1)
				a++
				os.WriteFile("clientscount", []byte(strconv.Itoa(a)), 0644)
				log.Printf("New WS client connected. Total WS clients: %d", b.totalClients.Load())

			case client := <-b.unregister:
				// A client has disconnected or is dead
				if _, loaded := b.clients.LoadAndDelete(client); loaded {
					close(client.send) // Close the client's send channel
					b.totalClients.Add(-1)
					log.Printf("WS client disconnected. Total WS clients: %d", b.totalClients.Load())
				}

			case message := <-b.broadcast:
				// Broadcast message to all active clients
				b.clients.Range(func(key, value interface{}) bool {
					client, ok := key.(*WebSocketClient)
					if !ok {
						log.Printf("Error: Non-WebSocketClient found in clients map.")
						return true // Continue iteration
					}
					select {
					case client.send <- message:
						// Message sent successfully
					case <-client.done:
						// Client's goroutine is already exiting, will be unregistered soon
						log.Printf("Skipping dead WS client during broadcast (done signal).")
					default:
						// Client's channel is blocked, indicating a slow consumer.
						// We can log this or implement more sophisticated backpressure.
						log.Printf("WS client channel blocked, potentially slow consumer. Unregistering.")
						b.unregister <- client // Signal to unregister this slow client
					}
					return true // Continue iteration
				})
			}
		}
	}()
}

// WebSocketHandler is the HTTP handler for WebSocket connections.
func (b *WebSocketBroker) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	client := &WebSocketClient{
		broker: b,
		conn:   conn,
		send:   make(chan []byte, 256), // Buffered channel for client's outbound messages
		done:   make(chan struct{}),
	}

	b.register <- client // Register the new client with the broker

	// Start goroutines for reading and writing WebSocket messages for this client
	go client.writePump()
	go client.readPump() // readPump will handle unregistering on disconnect
}

// StartBroadcastingTimeAndClients continuously broadcasts the current time and client count.
func (b *WebSocketBroker) StartBroadcastingTimeAndClients() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get current time and total clients
			currentTime := time.Now().Format(time.RFC3339)
			clientCount := b.totalClients.Load() // Atomically load the count

			// Create a JSON message
			messageData := struct {
				Time        string `json:"time"`
				ClientCount int64  `json:"client_count"`
			}{
				Time:        currentTime,
				ClientCount: clientCount,
			}
			jsonMessage, err := json.Marshal(messageData)
			if err != nil {
				log.Printf("Error marshalling broadcast message: %v", err)
				continue
			}

			// Send message to the broker for broadcasting
			select {
			case b.broadcast <- jsonMessage:
				// Message sent to broadcast channel
			default:
				log.Println("Broadcast channel is full, dropping message.")
			}

		case <-time.After(1 * time.Minute):
			// Periodically check if clients exist (optional)
			if b.totalClients.Load() == 0 {
				log.Println("No active WS clients. Consider pausing broadcasts to save CPU.")
			}
		}
	}
}

// Accessors for external use (e.g., from main.go if needed)
func (b *WebSocketBroker) GetTotalClients() int64 {
	return b.totalClients.Load()
}
