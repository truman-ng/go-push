package server

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go-push/common/utils"
	"go-push/internal/models"
	"go-push/pkg/redis"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	clientKey     = "ws:client:token:"
	clientManager = DefaultClientManager
)

// 启动 WebSocket 服务
func StartWebSocketServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("✅ WebSocket Server started on port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ WebSocket Server failed to start: %v", err)
	}
}

// WebSocket 握手处理逻辑
func wsHandler(w http.ResponseWriter, r *http.Request) {
	//for k, v := range r.Header {
	//	log.Printf("🔍 Header [%s] = %v", k, v)
	//}

	// 1. Token 校验
	token := r.URL.Query().Get("token")
	if token == "" {
		httpError(w, "Missing token", http.StatusUnauthorized)
		return
	}

	client := &models.Client{}
	err := redis.GetStructValue(clientKey+token, client)
	client.Token = token
	client.LastPingTime = time.Now() // ✅ 第一次心跳时间
	if err != nil {
		log.Printf("❌ Redis token error: %v, token: %v", err, token)
		httpError(w, "Token error", http.StatusInternalServerError)
		return
	}
	// 2. WebSocket 请求头校验（更健壮）
	if !isWebSocketRequest(r) {
		log.Printf("❌ Invalid WebSocket headers. Upgrade: %s, Connection: %s",
			r.Header.Get("Upgrade"), r.Header.Get("Connection"))
		httpError(w, "Invalid WebSocket handshake", http.StatusBadRequest)
		return
	}
	// 强制覆盖 header，兼容 gobwas 的严格逻辑
	r.Header.Set("Connection", "Upgrade")

	// 3. 执行协议升级
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		httpError(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	client.Conn = conn // 注册前保存连接
	// 4. 初始化连接
	client.IP = utils.GetClientIP(r)
	// 初始化 client.Send
	client.Send = make(chan []byte, 100)
	log.Printf("✅ Client connected: %+v", client)
	clientManager.Register <- client

	// ✅ 客户端只读，服务端不需要接收
	go handleWrite(conn, client) // 写消息协程
	select {}                    // 阻塞连接，不退出
}

// isWebSocketRequest checks if the request is a valid WebSocket handshake request.
func isWebSocketRequest(r *http.Request) bool {
	// Check if Upgrade header equals "websocket" (case-insensitive, trimmed)
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return false
	}

	// Parse Connection header values (e.g., "keep-alive, Upgrade")
	connHeader := r.Header.Values("Connection")
	for _, val := range connHeader {
		// Split multiple values by comma
		for _, part := range strings.Split(val, ",") {
			if strings.EqualFold(strings.TrimSpace(part), "upgrade") {
				return true
			}
		}
	}
	return false
}

// 将错误返回客户端
func httpError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}

// 写消息循环
func handleWrite(conn net.Conn, client *models.Client) {
	for message := range client.Send {
		err := wsutil.WriteServerMessage(conn, ws.OpText, message)
		if err != nil {
			log.Printf("❌ 写入失败：%v", err)
		}
	}

}

// 断开连接监听
func handleClose(conn net.Conn, client *models.Client) {
	defer func() {
		if client.Conn != nil {
			if err := client.Conn.Close(); err != nil {
				log.Printf("⚠️ Client %v close error: %v", client, err)
			}
		}

	}()
	defer func() {
		clientManager.Unregister <- client
	}()

	for {
		_, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			if err == io.EOF {
				log.Printf("🔌 Client disconnected: %v", client)
			} else {
				log.Printf("❌ Error reading from client %v: %v", client, err)
			}
			break
		}

		log.Printf("⚠️ Client %v tried to send message while disconnected", client)
		break
	}
}
