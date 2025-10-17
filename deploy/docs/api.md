# 模型服务管理 API 文档

## 1. 获取模型路径列表

**请求方法**：`GET`  
**请求 URL**：`/api/v1/modserv/list`

**响应体**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "modelName": "Llama-3-8B",
      "modelPath": "/models/llama-3-8b"
    },
    {
      "modelName": "bge-Large",
      "modelPath": "/models/bge-large"
    },
    {
      "modelName": "xdeepseekv3",
      "modelPath": "/models/xdeepseekv3"
    }
  ]
}
```

> **说明**：
> - `code`: `0` 表示成功，`1` 表示失败。
> - `modelPath` 字段仅供调试使用，前端不应展示。

---

## 2. 部署新服务（创建服务）

**请求方法**：`POST`  
**请求 URL**：`/api/v1/modserv/deploy`  
**请求头**：
```http
Content-Type: application/json
```

**请求体**：
```json
{
  "modelName": "xdeepseekv3",
  "resourceRequirements": {
    "acceleratorCount": 8
  },
  "replicaCount": 1,
  "contextLength": 1234
}
```

**响应体**：
```json
{
  "code": 0,
  "message": "deploy submit success",
  "data": {
    "serviceId": "xdeepseekv3-<UUID>"
  }
}
```

> **说明**：
> - `replicaCount` 当前固定为 `1`，前端可预留字段。
> - 成功后返回唯一 `serviceId`，用于后续操作。

---

## 3. 更新服务配置

**请求方法**：`PUT`  
**请求 URL**：`/api/v1/modserv/{serviceId}`（例如：`/api/v1/modserv/xdeepseekv3-a1b2c3d4`）  
**请求头**：
```http
Content-Type: application/json
```

**请求体**：
```json
{
  "modelName": "xdeepseekv3",
  "resourceRequirements": {
    "acceleratorCount": 2
  },
  "replicaCount": 1,
  "contextLength": 1234
}
```

**响应体**：
```json
{
  "code": 0,
  "message": "update submit success",
  "data": {
    "serviceId": "xdeepseekv3-<UUID>"
  }
}
```

> **说明**：
> - 通过 URL 中的 `serviceId` 定位要更新的服务。
> - 更新操作为异步提交，实际生效可能有延迟。

---

## 4. 查看服务状态

**请求方法**：`GET`  
**请求 URL**：`/api/v1/modserv/{serviceId}`

**响应体**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "serviceId": "xdeepseekv3-<UUID>",
    "modelName": "xdeepseekv3",
    "status": "running",
    "endpoint": "http://xxxx:xxxx/xx",
    "updateTime": "2025-09-01 14:30"
  }
}
```

> **状态说明**：
> - `running`：运行中
> - `pending`：阻塞中（资源不足等）
> - `failed`：部署失败
> - `initializing`：初始化中
> - `notExist`：服务不存在
> - `terminating`：正在删除中

---

## 5. 删除服务

**请求方法**：`DELETE`  
**请求 URL**：`/api/v1/modserv/{serviceId}`  
**请求体**：无

**响应体**：
```json
{
  "code": 0,
  "message": "delete submit success",
  "data": {
    "serviceId": "xdeepseekv3-<UUID>"
  }
}
```

> **说明**：
> - 删除为异步操作，服务状态将变为 `terminating`。
> - 资源释放可能需要一定时间。
