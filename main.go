package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery" // 引入 BigQuery 套件
	"github.com/gorilla/websocket"
)

// 定義 WebSocket 收到的資料格式
type TradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

// 定義要寫入 BigQuery 的資料格式 (對應我們剛剛建的 Table)
type BigQueryRow struct {
	EventTime time.Time `bigquery:"event_time"`
	Symbol    string    `bigquery:"symbol"`
	Price     float64   `bigquery:"price"`
}

var bqClient *bigquery.Client
var projectID string

func main() {
	// --- 0. 初始化環境變數與 BigQuery ---
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		// 本機開發時如果沒設環境變數，這裡要填你的 Project ID (或是讓它報錯)
		log.Println("⚠️ Warning: GOOGLE_CLOUD_PROJECT not set. BigQuery writes might fail locally.")
	}

	ctx := context.Background()
	var err error
	// 初始化 BigQuery Client
	// 注意：在 Cloud Run 上它會自動讀取權限；在本機你可能需要設定 key.json 才能測通
	if projectID != "" {
		bqClient, err = bigquery.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("❌ Failed to create BigQuery client: %v", err)
		} else {
			log.Println("✅ BigQuery client initialized")
		}
	}

	// --- 1. 啟動 Web Server ---
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Crypto Monitor with BigQuery is Running! 🚀")
		})
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	// --- 2. WebSocket 連線邏輯 (含斷線重連) ---
	url := "wss://stream.binance.com:9443/ws/btcusdt@trade"
	retryDelay := 1 * time.Second
	maxDelay := 60 * time.Second

	for {
		log.Printf("Connecting to Binance (%s)...", url)
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in %v...", err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
			if retryDelay > maxDelay {
				retryDelay = maxDelay
			}
			continue
		}

		log.Println("✅ Connected to Binance!")
		retryDelay = 1 * time.Second // 重置退避時間

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("❌ Read error: %v", err)
				c.Close()
				break
			}

			// 解析 JSON
			var event TradeEvent
			if err := json.Unmarshal(message, &event); err == nil {
				// 轉換價格字串為浮點數
				priceFloat, _ := strconv.ParseFloat(event.Price, 64)

				// 印出 Log
				log.Printf("[%s] %s: %s", time.Now().Format("15:04:05"), event.Symbol, event.Price)

				// --- 寫入 BigQuery (核心新增) ---
				if bqClient != nil {
					writeToBigQuery(event.Symbol, priceFloat)
				}
			}
		}
	}
}

// 獨立函式：寫入資料到 BigQuery
func writeToBigQuery(symbol string, price float64) {
	ctx := context.Background()
	// 定義要寫入的資料
	row := BigQueryRow{
		EventTime: time.Now(),
		Symbol:    symbol,
		Price:     price,
	}

	// 執行寫入 (Inserter)
	inserter := bqClient.Dataset("crypto_data").Table("trades").Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		// 這裡只印錯誤但不中斷程式，避免資料庫問題影響連線
		log.Printf("⚠️ BigQuery insert error: %v", err)
	}
}
