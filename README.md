# 📡 go-push WebSocket 推送服务

一个基于 Go 实现的高性能 WebSocket 推送服务，支持客户端注册、心跳机制、消息广播、连接超时清理等功能。适用于比分推送、实时通知等高频率场景。

---

## 🛠️ 技术栈

- **Gin**：轻量级 Web 框架（用于 RESTful API）
- **Redis**：客户端信息持久化存储
- **Gobwas/ws**：高性能 WebSocket 实现
- **Go Modules**：依赖管理

---

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/yourname/go-push.git
cd go-push
```

### 2. 安装依赖

```bash
go mod tidy
```
### 3. 启动服务

```bash
go run main.go
```

服务将启动两个端口：

- RESTful API：http://localhost:8080/

- WebSocket 服务：ws://localhost:8081/ws

---

## 🧩 功能介绍

- **获取Token**：用户/设备注册生成连接Token。

- **WebSocket连接**：携带Token进行连接认证。

- **心跳机制**：维持连接活性。

- **消息推送**：服务端广播消息到所有在线客户端。

- **连接超时清理**：定时移除30秒内未心跳的客户端。

---

## 📡 API 接口

### 获取连接Token

```bash
GET http://localhost:8080/get/token?deviceId=xxx&userId=yyy&isLogin=true
```

返回示例：

```json
{
  "data": "base64TokenString"
}
```

### 心跳

```bash
GET http://localhost:8080/heartbeat?token=base64TokenString
```

### 推送消息

```bash
POST http://localhost:8080/push/new/score/message
Content-Type: application/json

{
  "message": "这是一条广播消息"
}
```

---

## 健康检查接口（Health Check）

本服务提供了用于 `Kubernetes` 就绪检查（`readinessProbe`）与存活检查（`livenessProbe`）的接口：

### 接口说明

- 路径：`GET /health`

- 端口：`8080`

- 响应示例：

```json
{
  "status": "healthy"
}
```

### 用途

在 `Kubernetes` 的 `Deployment.yaml` 中，已将该接口配置为健康检查探针：

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 3

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

- **Readiness Probe**：确保服务已准备好处理请求。

- **Liveness Probe**：确保服务健康运行，若失败会触发自动重启。

---

##  WebSocket 连接方式

连接示例：

```bash
wss://localhost:8081/ws?token=base64TokenString
```

客户端建立连接后，将持续接收服务端广播的消息。

---

## 🧹 客户端清理机制

定时任务 `ClientTimeoutChecker` 每10秒检查一次所有客户端：

- 若超过30秒未发送心跳，将关闭连接并移除缓存。


---

## 📁 项目结构概览

| 文件名                  | 说明                      |
| -------------------- | ----------------------- |
| `websocket.go`       | WebSocket 启动与握手逻辑（8081） |
| `client-manager.go`  | 客户端管理与注册处理              |
| `message-manager.go` | 消息推送处理                  |
| `route.go`           | HTTP 路由注册（8080）         |
| `schedule_task.go`   | 客户端定时检查任务               |



