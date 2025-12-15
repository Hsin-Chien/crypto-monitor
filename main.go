package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// 定義資料結構：這就像是告訴 Go，等一下收到的 JSON 資料長什麼樣子
// 我們只抓取兩個欄位：'s' (交易對名稱) 和 'p' (成交價格)
type TradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

func main() {
	// 1. 設定幣安 (Binance) 的公開 WebSocket 網址
	// btcusdt@trade 代表訂閱比特幣/USDT 的即時成交資訊
	url := "wss://stream.binance.com:9443/ws/btcusdt@trade"

	fmt.Printf("準備連線到幣安: %s ...\n", url)

	// 2. 建立連線
	// DefaultDialer 就像是幫我們撥電話的總機
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("連線失敗，請檢查網路:", err)
	}
	// defer 確保程式結束前會把電話掛斷 (關閉連線)
	defer c.Close()

	fmt.Println("連線成功！正在等待比特幣價格進來...")

	// 3. 開始無窮迴圈，持續收聽
	for {
		// 讀取從幣安傳過來的訊息
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("讀取錯誤:", err)
			break
		}

		// 4. 解析資料 (把 JSON 文字轉成我們看得懂的 Go 結構)
		var event TradeEvent
		// Unmarshal 就是「解碼」的意思
		if err := json.Unmarshal(message, &event); err != nil {
			log.Println("解析錯誤:", err)
			continue
		}

		// 5. 漂亮地印出來
		// time.Now() 抓取現在時間
		fmt.Printf("[%s] %s 目前價格: %s\n",
			time.Now().Format("15:04:05"),
			event.Symbol,
			event.Price)
	}
}
