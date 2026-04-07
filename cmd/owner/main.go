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
	"syscall"

	"breakfast-system/pkg/protocol"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "server base url")
	orderID := flag.Int("order-id", 0, "update target order id")
	nextStatus := flag.String("status", "", "next status: 制作中|待取餐|已完成")
	flag.Parse()

	if *orderID > 0 && *nextStatus != "" {
		order, err := updateStatus(*serverURL, *orderID, protocol.Status(*nextStatus))
		if err != nil {
			fmt.Fprintf(os.Stderr, "更新订单状态失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("状态更新成功 => 订单ID: %d, 当前状态: %s\n", order.ID, order.Status)
	}

	if err := printOwnerOrders(*serverURL); err != nil {
		fmt.Fprintf(os.Stderr, "读取订单列表失败: %v\n", err)
		os.Exit(1)
	}

	go listenOwnerUpdates(*serverURL)

	fmt.Println("正在监听商家实时订单，按 Ctrl+C 退出")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func printOwnerOrders(serverURL string) error {
	resp, err := http.Get(serverURL + "/owner/orders")
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

	fmt.Println("商家视角订单列表:")
	for _, order := range orders {
		fmt.Printf("- 订单ID: %d, 状态: %s, 菜品: %v\n", order.ID, order.Status, order.Items)
	}
	return nil
}

func updateStatus(serverURL string, orderID int, status protocol.Status) (protocol.Order, error) {
	reqBody, err := json.Marshal(protocol.UpdateStatusRequest{Status: status})
	if err != nil {
		return protocol.Order{}, err
	}

	resp, err := http.Post(serverURL+fmt.Sprintf("/owner/orders/%d/status", orderID), "application/json", bytes.NewBuffer(reqBody))
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

func listenOwnerUpdates(serverURL string) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(serverURL, "/owner/ws"), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接商家 WebSocket 失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("商家 WebSocket 已连接")

	for {
		var event protocol.Event
		if err := conn.ReadJSON(&event); err != nil {
			fmt.Fprintf(os.Stderr, "读取商家实时消息失败: %v\n", err)
			return
		}

		fmt.Printf("商家视角更新 => 类型: %s, 订单ID: %d, 状态: %s, 菜品: %v\n", event.Type, event.Data.ID, event.Data.Status, event.Data.Items)
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
