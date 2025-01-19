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
)

var (
	//clients     = make(map[string]*models.Client)
	//clientsLock sync.RWMutex
	clientKey     = "ws:client:token:"
	clientManager = NewClientManager()
)

func StartWebSocketServer(port string) {
	http.HandleFunc("/ws", wsHandler)
	log.Println("Starting WebSocket Server on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func wsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}
	client := &models.Client{}
	err := redis.GetStructValue(clientKey+token, client)
	if err != nil {
		log.Printf("token error: %v, token: %v", err, token)
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		http.Error(w, "websocket upgrade error", http.StatusInternalServerError)
		return
	}

	client.IP = utils.GetClientIP(r)

	log.Println("client info: ", client)
	clientManager.Register <- client

	go handleWrite(conn, client)

	handleClose(conn, client)

	for {
		_, _, err := wsutil.ReadClientData(conn)
		//msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			break
		}
		// 打印接收到的消息
		//log.Printf("Received message from client: %s (op: %d)", string(msg), op)

		// 可选：回显消息到客户端
		//err = wsutil.WriteServerMessage(conn, op, msg)
		//if err != nil {
		//	log.Printf("Error writing message to client: %v", err)
		//	break
		//}
	}
}

func handleWrite(conn net.Conn, client *models.Client) {
	for message := range client.Send {
		err := wsutil.WriteServerMessage(conn, ws.OpText, message)
		if err != nil {
			log.Printf("client: %v write error:\n %v", client, err)
		}
	}
}

func handleClose(conn net.Conn, client *models.Client) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("client: %v close error:\n %v", client, err)
		}
	}(conn)
	defer func() { clientManager.Unregister <- client }()
	for {
		_, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected", client)
			} else {
				log.Printf("Error reading client data: %v", err)
			}
			break
		}

		log.Printf("Client %s tried to send a message. Disconnecting.", client)
		break
	}
}
