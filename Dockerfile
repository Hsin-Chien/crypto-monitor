# 1. 選擇基底：使用官方的 Go 語言環境 (版本對應你剛裝的)
FROM golang:1.25.5

# 2. 設定工作目錄：在箱子裡建立一個叫 /app 的資料夾
WORKDIR /app

# 3. 複製身分證：先把 go.mod 和 go.sum 複製進去
COPY go.mod go.sum ./

# 4. 安裝零件：讓 Docker 在箱子裡下載所有需要的套件
RUN go mod download

# 5. 複製程式碼：把剩下的程式碼 (.go 檔) 全部複製進去
COPY . .

# 6. 編譯：把程式碼變成可執行的二進位檔 (命名為 main)
RUN go build -o main .

# 7. 啟動命令：當箱子被打開時，執行這個指令
CMD ["./main"]
