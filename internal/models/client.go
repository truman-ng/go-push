package models

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	Conn         net.Conn
	IP           string
	UserId       string
	DeviceId     string
	IsLogin      bool
	LastPingTime time.Time
	Lock         sync.RWMutex
	Token        string
	Send         chan []byte `json:"-"`
	RoomIds      []string
}

func (client *Client) UpdatePingTime() {
	client.Lock.Lock()
	defer client.Lock.Unlock()
	client.LastPingTime = time.Now()
}
