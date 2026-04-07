package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"breakfast-system/pkg/protocol"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "server base url")
	flag.Parse()

	if err := printDisplayOrders(*serverURL); err != nil {
		fmt.Fprintf(os.Stderr, "读取取餐屏列表失败: %v\n", err)
		os.Exit(1)
	}

	go listenDisplayUpdates(*serverURL)

	fmt.Println("正在监听取餐大屏更新，按 Ctrl+C 退出")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func printDisplayOrders(serverURL string) error {
	resp, err := http.Get(serverURL + "/display/orders")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	var orders []protocol.Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return err
	}

	fmt.Println("取餐大屏当前订单:")
	for _, order := range orders {
		fmt.Printf("- 订单ID: %d, 状态: %s, 菜品: %v\n", order.ID, order.Status, order.Items)
	}
	return nil
}

func listenDisplayUpdates(serverURL string) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(serverURL, "/display/ws"), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接取餐屏 WebSocket 失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("取餐屏 WebSocket 已连接")

	for {
		var event protocol.Event
		if err := conn.ReadJSON(&event); err != nil {
			fmt.Fprintf(os.Stderr, "读取取餐屏实时消息失败: %v\n", err)
			return
		}

		fmt.Printf("取餐屏更新 => 订单ID: %d, 状态: %s\n", event.Data.ID, event.Data.Status)
	}
}

func wsURL(serverURL, path string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return ""
	}

	switch parsed.Scheme {
	case "https":
		parsed.Scheme = "wss"
	default:
		parsed.Scheme = "ws"
	}

	parsed.Path = path
	parsed.RawQuery = ""
	return parsed.String()
}
