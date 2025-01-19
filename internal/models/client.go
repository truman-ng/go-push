package models

import (
	"sync"
	"time"
)

type Client struct {
	//Conn         *wsutil.Writer
	IP           string
	UserId       string
	DeviceId     string
	IsLogin      bool
	LastPingTime time.Time
	Lock         sync.RWMutex
	Token        string
	Send         chan []byte
}

func (client *Client) UpdatePingTime() {
	client.Lock.Lock()
	defer client.Lock.Unlock()
	client.LastPingTime = time.Now()
}
