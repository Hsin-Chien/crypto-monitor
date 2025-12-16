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

type TradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

func main() {
	// --- 1. 啟動 Web Server (讓 Cloud Run 知道我們活著) ---
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Crypto Monitor is Running! 🚀")
		})
		log.Printf("Web server listening on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// --- 2. 核心業務：WebSocket 斷線重連機制 ---
	url := "wss://stream.binance.com:9443/ws/btcusdt@trade"

	// 指數退避設定
	retryDelay := 1 * time.Second
	maxDelay := 60 * time.Second

	// 外層：負責「重連」的無窮迴圈
	for {
		log.Printf("Connecting to Binance (%s)...", url)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)

		if err != nil {
			log.Printf("Connection failed: %v", err)
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)

			// 失敗時，等待時間加倍 (1s -> 2s -> 4s -> ... -> 60s)
			retryDelay *= 2
			if retryDelay > maxDelay {
				retryDelay = maxDelay
			}
			continue // 跳回迴圈開頭重試
		}

		// 連線成功！重置等待時間
		log.Println("✅ Connected to Binance!")
		retryDelay = 1 * time.Second

		// 內層：負責「讀取資料」的迴圈
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("❌ Disconnected: %v", err)
				c.Close() // 確保關閉舊連線
				break     // 跳出內層迴圈，觸發外層的重連邏輯
			}

			var event TradeEvent
			if err := json.Unmarshal(message, &event); err == nil {
				log.Printf("[%s] %s: %s", time.Now().Format("15:04:05"), event.Symbol, event.Price)
			}
		}
		// 當程式執行到這裡，代表內層迴圈 break 了，會自動回到外層迴圈進行重連
	}
}
