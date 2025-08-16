// CpuinfoRoute streams live CPU and memory usage using SSE

// TimeOnlyRoute returns only the current time as plain text

// SSE handler for time and client count only

package main

import (
	"fmt"
	"gosse/Live"
	"gosse/chat"
	"gosse/futurepaper"
	"gosse/gift"
	"gosse/lottosociety"
	"gosse/threedata"
	"gosse/twoddata"
	"gosse/user"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

func main() {
	// Set Yangon timezone
	yangonLoc, err := time.LoadLocation("Asia/Yangon")
	if err != nil {
		log.Fatalf("Failed to load Yangon timezone: %v", err)
	}
	time.Local = yangonLoc
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

	//Recovery From Panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			debug.PrintStack()
		}
	}()

	// Initialize the report table in the database
	// Initialize SQLite DB and table
	// Initialize twoddata database

	db := twoddata.InitDB("twoddata.db")
	chat.InitBanTable(db) // Initialize the ban table in the database
	chat.InitReportTable(db)
	// Initialize other databases
	giftDB := gift.InitGiftDB("twoddata.db")
	threedDB := threedata.InitThreedDB("twoddata.db")
	defer db.Close()
	defer giftDB.Close()

	lottosociety.InitLottoSocietyTable(db)

	es := user.CreateUserAccountTable(db)
	if es != nil {
		log.Fatal("Failed to create useraccount table:", es)
	}
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

	http.HandleFunc("/live", brokerr.SSEHandler)
	http.HandleFunc("/history", Live.TwoddataHandler(db))
	http.HandleFunc("/addlive", Live.AddLiveDataHandler)
	http.HandleFunc("/livess", Live.LiveDataPageHandler)
	http.HandleFunc("/livedata/sse", Live.LiveDataSSEHandler)
	http.HandleFunc("/threed", threedata.ThreedDataHandler(threedDB))
	http.HandleFunc("/gift", gift.GiftDataHandler(giftDB))
	http.HandleFunc("/addgift/", gift.AddGiftHandler(giftDB))

	http.HandleFunc("/futurepaper/getallpaper/low", futurepaper.GetLowPaperHandler)
	http.HandleFunc("/futurepaper/getallpaper/high", futurepaper.GetHighPaperHandler)

	http.HandleFunc("/chat/sendmessage", chat.SendMessageHandler(db))
	http.HandleFunc("/chat/sse", chat.ChatSSEHandler)
	http.HandleFunc("/register", user.RegisterUserHandler(db))
	http.HandleFunc("/chat/ban", chat.BanHandler(db)) // Alias for ban handler
	http.HandleFunc("/chat/report", chat.ReportHandler(db))
	http.HandleFunc("/futurepaper/addpaper", futurepaper.UploadPaperImageHandler)       // Alias for add paper handler
	http.HandleFunc("/lottosociety/addlotto", lottosociety.AddOrUpdateLottoHandler(db)) // Alias for add lotto handler
	http.HandleFunc("/lottosociety/getlotto", lottosociety.GetLottoHandler(db))         // Alias for get lotto handler
	// Alias for delete all lotto handler
	// Alias for login handler
	// Alias for report handler
	// --- WebSocket Broker Setup (NEW) ---
	wsBroker := Live.NewWebSocketBroker()             // Initialize the new WebSocket broker
	wsBroker.Start()                                  // Start the WebSocket broker's main loop
	go wsBroker.StartBroadcastingTimeAndClients()     // Start broadcasting WS data
	http.HandleFunc("/ws", wsBroker.WebSocketHandler) // Handle WebSocket connections

	// Serve static images from /images/
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.Handle("/gift/images/", http.StripPrefix("/gift/images/", http.FileServer(http.Dir("gift/images"))))
	http.Handle("/futurepaper/images/daily/", http.StripPrefix("/futurepaper/images/daily/", http.FileServer(http.Dir("futurepaper/images/daily"))))
	http.Handle("/futurepaper/images/weekly/", http.StripPrefix("/futurepaper/images/weekly/", http.FileServer(http.Dir("futurepaper/images/weekly"))))
	http.Handle("/futurepaper/images/calendar/", http.StripPrefix("/futurepaper/images/calendar/", http.FileServer(http.Dir("futurepaper/images/calendar"))))

	log.Println("SSE server started on :1411")
	fmt.Println(http.ListenAndServe(":1411", nil))
}
