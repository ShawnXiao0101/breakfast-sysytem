package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
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
}
