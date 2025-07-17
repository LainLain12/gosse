package Live

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// --- WebSocket client management ---
var (
	wsClients      = make(map[*websocket.Conn]struct{})
	wsClientsMutex sync.RWMutex
)

// LiveWebSocketHandler handles websocket connections and broadcasts liveDataStore updates every second
func LiveWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wsClientsMutex.Lock()
	wsClients[conn] = struct{}{}
	wsClientsMutex.Unlock()
	defer func() {
		wsClientsMutex.Lock()
		delete(wsClients, conn)
		wsClientsMutex.Unlock()
		conn.Close()
	}()

	// Reader goroutine (optional, for receiving messages from client)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Block main handler until disconnect
	select {}
}

// BroadcastLiveDataToWebSockets sends the latest liveDataStore to all websocket clients every second
func StartLiveWebSocketBroadcast() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			wsClientsMutex.RLock()
			data, _ := json.Marshal(liveDataStore)
			for conn := range wsClients {
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					// Remove dead connection
					wsClientsMutex.RUnlock()
					wsClientsMutex.Lock()
					delete(wsClients, conn)
					wsClientsMutex.Unlock()
					wsClientsMutex.RLock()
				}
			}
			wsClientsMutex.RUnlock()
		}
	}()
}
