### question1 
#### time:2025/05/25 
#### 问题：websocket可以连接，但是注册不成功 
#### 解决：main.go - go server.DefaultClientManager.Start() 和 internal/server/client-manager.go 中的 var DefaultClientManager = NewClientManager() 还有 internal/server/websocket.go 中的 clientManager = DefaultClientManager 要都是同一个实例                
