package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	go listenRealtimeUpdates()

	time.Sleep(200 * time.Millisecond)

	body := []byte(`{"items":["包子","豆浆","蛋挞"]}`)

	resp, err := http.Post("http://localhost:8080/order", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "下单失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "下单失败: %s\n响应内容: %s\n", resp.Status, string(respBody))
		os.Exit(1)
	}

	fmt.Printf("下单成功: %s\n响应内容: %s\n", resp.Status, string(respBody))

	fmt.Println("正在监听实时更新，按 Ctrl+C 退出")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

type orderEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type order struct {
	ID     int      `json:"id"`
	Items  []string `json:"items"`
	Status string   `json:"status"`
}

func listenRealtimeUpdates() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接 WebSocket 失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("WebSocket 已连接")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取实时消息失败: %v\n", err)
			return
		}

		var event orderEvent
		if err := json.Unmarshal(message, &event); err != nil {
			fmt.Printf("收到无法解析的消息: %s\n", string(message))
			continue
		}

		var o order
		if err := json.Unmarshal(event.Data, &o); err != nil {
			fmt.Printf("收到事件 %s，但订单数据解析失败: %s\n", event.Type, string(message))
			continue
		}

		fmt.Printf("实时更新 => 类型: %s, 订单ID: %d, 状态: %s, 菜品: %v\n", event.Type, o.ID, o.Status, o.Items)
	}
}
