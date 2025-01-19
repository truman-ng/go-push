package test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestWebSocketConnection(t *testing.T) {
	// 创建上下文，设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// WebSocket 服务器地址
	serverAddr := "ws://localhost:8089/ws"

	// 使用 ws.Dialer 连接
	conn, _, _, err := ws.Dialer{}.Dial(ctx, serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	// 升级为 WebSocket 连接
	_, err = ws.Upgrade(conn)
	if err != nil {
		t.Fatalf("Failed to upgrade connection to WebSocket: %v", err)
	}
	log.Println("Connected to WebSocket server")

	// 定时发送消息
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	// 启动一个 Goroutine 监听服务端消息
	go func() {
		defer close(done)
		for {
			// 从服务端读取消息
			msg, op, err := wsutil.ReadServerData(conn)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}
			log.Printf("Received message from server: %s (op: %d)", string(msg), op)
		}
	}()

	// 每 5 秒发送一条消息
	for i := 0; i < 5; i++ { // 发送 5 次后退出
		message := []byte(`{"type":"ping","message":"Hello, WebSocket!"}`)
		err = wsutil.WriteClientMessage(conn, ws.OpText, message)
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}
		log.Printf("Sent message: %s", message)
		time.Sleep(5 * time.Second)
	}

	// 关闭 WebSocket 连接
	log.Println("Closing WebSocket connection")
	err = wsutil.WriteClientMessage(conn, ws.OpClose, []byte(`closed`))
	if err != nil {
		log.Printf("Failed to send close message: %v", err)
	}
	<-done
}
