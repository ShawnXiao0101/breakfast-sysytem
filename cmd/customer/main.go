package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"breakfast-system/pkg/protocol"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "server base url")
	itemsFlag := flag.String("items", "包子,豆浆,蛋挞", "comma-separated order items")
	flag.Parse()

	items := splitItems(*itemsFlag)
	order, err := createOrder(*serverURL, items)
	if err != nil {
		fmt.Fprintf(os.Stderr, "下单失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("下单成功，订单ID: %d，当前状态: %s，菜品: %v\n", order.ID, order.Status, order.Items)

	go listenCustomerUpdates(*serverURL, order.ID)

	fmt.Println("正在监听顾客订单更新，按 Ctrl+C 退出")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func createOrder(serverURL string, items []string) (protocol.Order, error) {
	reqBody, err := json.Marshal(protocol.CreateOrderRequest{Items: items})
	if err != nil {
		return protocol.Order{}, err
	}

	resp, err := http.Post(serverURL+"/customer/orders", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return protocol.Order{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return protocol.Order{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return protocol.Order{}, fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	var order protocol.Order
	if err := json.Unmarshal(body, &order); err != nil {
		return protocol.Order{}, err
	}

	return order, nil
}

func listenCustomerUpdates(serverURL string, orderID int) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(serverURL, fmt.Sprintf("/customer/ws/orders/%d", orderID)), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接顾客 WebSocket 失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("顾客 WebSocket 已连接，监听订单 %d\n", orderID)

	for {
		var event protocol.Event
		if err := conn.ReadJSON(&event); err != nil {
			fmt.Fprintf(os.Stderr, "读取顾客实时消息失败: %v\n", err)
			return
		}

		fmt.Printf("顾客视角更新 => 类型: %s, 订单ID: %d, 状态: %s\n", event.Type, event.Data.ID, event.Data.Status)
	}
}

func splitItems(items string) []string {
	parts := strings.Split(items, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
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
