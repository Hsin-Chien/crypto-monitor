# 1. 選擇基底：使用官方的 Go 語言環境
FROM golang:1.25.5

# 2. 設定工作目錄
WORKDIR /app

# 3. 複製身分證
COPY go.mod go.sum ./

# 4. 安裝零件
RUN go mod download

# 5. 複製程式碼
COPY . .

# 6. 編譯
RUN go build -o main .

# 7. 啟動命令
CMD ["./main"]
