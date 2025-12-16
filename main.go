package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// 定義資料結構
type TradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

func main() {
	// --- 關鍵修改：Cloud Run 必要設定 ---
	// 必須要有一個 HTTP Server 監聽 PORT，否則 Cloud Run 會判定失敗
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 啟動一個背景 Goroutine 來處理 HTTP 請求
	go func() {
		log.Printf("Starting web server on port %s", port)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Crypto Monitor is Running! 🚀")
		})
		// 如果 Web Server 啟動失敗，直接讓程式崩潰重啟
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()
	// ----------------------------------

	// --- 你的業務邏輯 (WebSocket) ---
	url := "wss://stream.binance.com:9443/ws/btcusdt@trade"
	log.Printf("Connecting to %s", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("WebSocket connection failed: %v", err)
		// 為了防止程式直接退出導致 Cloud Run 以為我們死了，
		// 這裡即使連線失敗，我們也讓程式保持活著 (用 select{})
		// 下一步我們再來寫「斷線重連」
		select {}
	}
	defer c.Close()

	log.Println("Connected to Binance!")

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		var event TradeEvent
		json.Unmarshal(message, &event)
		log.Printf("[%s] %s: %s", time.Now().Format("15:04:05"), event.Symbol, event.Price)
	}
}
