// CpuinfoRoute streams live CPU and memory usage using SSE

// TimeOnlyRoute returns only the current time as plain text

// SSE handler for time and client count only

package main

import (
	"gosse/Live"
	"gosse/gift"
	"gosse/threedata"
	"gosse/twoddata"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	// 1. Log File ဖွင့်ခြင်း (သို့မဟုတ် အသစ်ဖန်တီးခြင်း)
	// Output log to a file named server.log
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close() // Application ပိတ်တဲ့အခါ log file ကို ပိတ်ဖို့ သေချာပါစေ။

	// 2. Go ရဲ့ Default Logger ရဲ့ Output ကို Log File ဆီ ပြောင်းခြင်း
	log.SetOutput(logFile)

	// Optional: Log message format ကို ချိန်ညှိနိုင်ပါတယ်။ (ဥပမာ - ရက်စွဲ၊ အချိန် နဲ့ log entry file name ထည့်ဖို့)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
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
			log.Println("Goroutines:", runtime.NumGoroutine())
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
	http.HandleFunc("/gemini", Live.GeminiPageHandler)

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
