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
	"runtime"
	"time"
)

func main() {

	// Route to show all twoddata rows as JSON
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Printf("Recovered from panic: %v", r)
	// 		debug.PrintStack()
	// 	}
	// }()
	// Initialize SQLite DB and table

	db := twoddata.InitDB("twoddata.db")
	giftDB := gift.InitGiftDB("twoddata.db")
	threedDB := threedata.InitThreedDB("twoddata.db")
	defer db.Close()
	defer giftDB.Close()

	/// check go routine count
	go func() {
		for {
			fmt.Println("Goroutines:", runtime.NumGoroutine())
			time.Sleep(10 * time.Second)
		}
	}()

	// Start goroutine to broadcast time/cpu/mem/client count every second

	brokerr := Live.NewBroker()
	brokerr.Start()
	go brokerr.StartBroadcastingTime()

	http.HandleFunc("/", brokerr.SSEHandler)
	http.HandleFunc("/ppp", Live.LiveDataSSEHandler)
	http.HandleFunc("/history", Live.TwoddataHandler(db))
	http.HandleFunc("/addlive", Live.AddLiveDataHandler)
	http.HandleFunc("/live", Live.LiveDataPageHandler)
	http.HandleFunc("/livedata/sse", Live.LiveDataSSEHandler)
	http.HandleFunc("/threed", threedata.ThreedDataHandler(threedDB))
	http.HandleFunc("/gift", gift.GiftDataHandler(giftDB))
	http.HandleFunc("/addimage/", gift.AddImageHandler(giftDB))

	// --- WebSocket Broker Setup (NEW) ---
	wsBroker := Live.NewWebSocketBroker()             // Initialize the new WebSocket broker
	wsBroker.Start()                                  // Start the WebSocket broker's main loop
	go wsBroker.StartBroadcastingTimeAndClients()     // Start broadcasting WS data
	http.HandleFunc("/ws", wsBroker.WebSocketHandler) // Handle WebSocket connections

	// Serve static images from /images/
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	log.Println("SSE server started on :4597")
	log.Fatal(http.ListenAndServe(":4597", nil))
}
