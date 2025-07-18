package Live

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin:     func(r *http.Request) bool { return true },
// }

// // --- WebSocket client management ---
// var (
// 	wsClients      = make(map[*websocket.Conn]struct{})
// 	wsClientsMutex sync.RWMutex
// )

// // LiveWebSocketHandler handles websocket connections and broadcasts liveDataStore updates only when changed
// func LiveWebSocketHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		return
// 	}
// 	wsClientsMutex.Lock()
// 	wsClients[conn] = struct{}{}
// 	fmt.Println("connect new client:", len(wsClients))
// 	wsClientsMutex.Unlock()
// 	defer func() {
// 		wsClientsMutex.Lock()
// 		delete(wsClients, conn)
// 		fmt.Println("WebSocket client disconnected")

// 		wsClientsMutex.Unlock()
// 		conn.Close()
// 	}()

// 	// Set up ping/pong to keep connection alive
// 	conn.SetReadLimit(512)
// 	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
// 	conn.SetPongHandler(func(string) error {
// 		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
// 		return nil
// 	})

// 	disconnect := make(chan struct{})

// 	go func() {
// 		for {
// 			if _, _, err := conn.ReadMessage(); err != nil {
// 				close(disconnect)
// 				return
// 			}
// 		}
// 	}()

// 	// Wait for disconnect signal
// 	<-disconnect
// }

// // StartLiveWebSocketBroadcast broadcasts liveDataStore only when changed, with client count
// func StartLiveWebSocketBroadcast() {
// 	go func() {
// 		var prevLive string
// 		for {
// 			time.Sleep(1 * time.Second)
// 			liveDataMu.Lock()
// 			var currLive string
// 			if len(liveDataStore) > 0 {
// 				currLive = liveDataStore[0].Live // Use your actual field
// 			}
// 			if currLive != prevLive {
// 				wsClientsMutex.RLock()
// 				clientCount := len(wsClients)
// 				broadcast := map[string]interface{}{
// 					"clients": clientCount,
// 					"data":    liveDataStore,
// 				}
// 				data, _ := json.Marshal(broadcast)
// 				for conn := range wsClients {
// 					conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
// 					_ = conn.WriteMessage(websocket.PingMessage, []byte{}) // Send ping
// 					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
// 						wsClientsMutex.RUnlock()
// 						wsClientsMutex.Lock()
// 						delete(wsClients, conn)
// 						wsClientsMutex.Unlock()
// 						wsClientsMutex.RLock()
// 					}
// 				}
// 				wsClientsMutex.RUnlock()
// 				prevLive = currLive
// 			}
// 			liveDataMu.Unlock()
// 		}
// 	}()
// }

// // LiveDataStore is a placeholder for the actual live data structure
