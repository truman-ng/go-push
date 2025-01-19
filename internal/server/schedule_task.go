package server

import "time"

func ClientTimeoutChecker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		RemoveStaleClient()
	}
}
