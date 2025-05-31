## 🧩 接口名称：获取推送 Token

- **接口地址**：`GET /get/token`
- **请求方式**：`GET`
- **接口用途**：根据登录状态生成用于消息推送的 Base64 编码 Token。

---

### 📥 请求参数

| 参数名     | 类型    | 是否必传 | 说明                                     |
|------------|---------|----------|------------------------------------------|
| `deviceId` | string  | 条件必传 | 当 `isLogin=false` 时，必须传入设备 ID。 |
| `userId`   | string  | 条件必传 | 当 `isLogin=true` 时，必须传入用户 ID。   |
| `isLogin`  | boolean | 必传     | 是否为已登录用户，取值 `"true"` 或 `"false"`（字符串类型） |

> ⚠️ 注意：参数 `isLogin` 会在后端解析为布尔值处理。

---

### 🔍 参数校验逻辑

- 当 `isLogin=false`，**必须传入 `deviceId`**；
- 当 `isLogin=true`，**必须传入 `userId`**；
- 不满足上述条件将返回 `400 Bad Request`。

---

### 📤 响应参数

#### ✅ 成功响应

```json
{
  "data": "base64编码后的Token字符串"
}
```

#### ❌ 失败响应（参数错误）

```json
{
  "data": "Invalid request parameters"
}
```

## 🫀 接口名称：心跳维持接口（HeartBeat）

- **接口地址**：`GET /heartbeat`
- **请求方式**：`GET`
- **接口用途**：维持 WebSocket 连接活跃状态，防止连接超时断开。

---

### 📥 请求参数

| 参数名   | 类型   | 是否必传 | 说明                         |
|----------|--------|----------|------------------------------|
| `token`  | string | 必传     | 客户端连接时分配的唯一标识（通常为登录后生成的 token） |

---

### 🕐 调用建议

- WebSocket 服务端会在 **30 秒内未收到心跳包** 时主动断开连接；
- 建议客户端每 **20~29 秒之间** 向此接口发送一次心跳请求。

---

### ✅ 成功响应

```json
{}
```

- HTTP 状态码：`200 OK`

- 说明：心跳成功，无需内容返回

---

### ❌ 错误响应

```json
{
  "data": "token error"
}
```

- HTTP 状态码：`500 Internal Server Error`

- 原因：传入的 `token` 无效，可能已过期或未注册到 Redis

## 🔌 WebSocket 接口：实时消息推送连接

- **接口地址**：`ws://<host>:<port>/ws`
- **连接协议**：`WebSocket`（使用标准 WebSocket 握手）
- **接口用途**：建立实时推送连接，支持订阅最多 2 个房间，用于消息广播、心跳维持等。

---

### 📥 握手参数（Query 参数）

| 参数名     | 类型   | 是否必传 | 说明                                              |
|------------|--------|----------|-------------------------------------------------|
| `token`    | string | ✅ 必传   | 通过 `/get/token` 接口获取，用于客户端身份识别（从 Redis 获取 Client） |
| `roomIds`  | string | ✅ 必传   | 房间号列表，英文逗号分隔，**最多支持 2 个**，如：`room1,room2`       |

---

### 🧠 握手逻辑说明

1. 校验是否为 WebSocket 请求（检查 `Upgrade` 和 `Connection` Header）
2. 验证 token 是否在 Redis 中存在（通过 `clientKey + token` 查询）
3. 限制最多只允许订阅 2 个房间（`roomIds`）
4. 初始化 Client 结构体并注册连接（包括 IP、发送通道、心跳时间）
5. 启动写协程 `handleWrite`，等待服务端推送消息
6. 不处理客户端主动消息，仅由服务端向客户端发送

---

### 📤 服务端主动推送消息（示例）

```json
{
  "type": "notice",
  "data": {}
}
```

> 所有消息通过文本帧（Text Frame）推送，格式为 JSON

---

### 🫀 心跳要求

- 客户端应每 20~29 秒 调用一次 /heartbeat?token=xxx 接口维持连接；

- 如果服务器 30 秒未收到心跳，将主动断开 WebSocket 连接；

---

### ⚠️ 错误响应（握手阶段）
| 错误码 | 信息                          | 说明                       |
| --- | --------------------------- | ------------------------ |
| 401 | Missing token               | 未提供 token 参数             |
| 401 | Missing room                | 未提供 roomIds 参数           |
| 500 | Token error                 | token 无效，Redis 中无法找到对应数据 |
| 400 | Invalid WebSocket handshake | 非法 WebSocket 请求头         |
| 500 | WebSocket upgrade failed    | 协议升级失败，连接未建立成功           |

---

### 📌 注意事项

- 连接参数必须在 URL 中传递，不支持通过 Header 传 token。

- 每个客户端最多只能订阅 2 个房间，多余的将被忽略。

- 不支持客户端主动发送消息（仅服务端推送）。

- 若 Redis 中未找到 token 对应信息，将无法建立连接。


