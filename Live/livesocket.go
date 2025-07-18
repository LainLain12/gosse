package Live

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

var hub = &Hub{
	clients:    make(map[*Client]struct{}),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				client.conn.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					client.conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// LiveWebSocketHandler handles websocket connections and broadcasts liveDataStore updates only when changed
func LiveWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
	hub.register <- client

	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// StartLiveWebSocketBroadcast broadcasts liveDataStore only when changed, with client count
func StartLiveWebSocketBroadcast() {
	go hub.run()
	go func() {
		var prevLive string
		for {
			time.Sleep(1 * time.Second)
			liveDataMu.Lock()
			var currLive string
			if len(liveDataStore) > 0 {
				currLive = liveDataStore[0].Live // Use your actual field
			}
			if currLive != prevLive {
				hub.mu.RLock()
				clientCount := len(hub.clients)
				broadcast := map[string]interface{}{
					"clients": clientCount,
					"data":    liveDataStore,
				}
				data, _ := json.Marshal(broadcast)
				hub.broadcast <- data
				hub.mu.RUnlock()
				prevLive = currLive
			}
			liveDataMu.Unlock()
		}
	}()
}

// LiveDataStore is a placeholder for the actual live data structure
