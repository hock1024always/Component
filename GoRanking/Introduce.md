# 实时排行榜数据广播系统

## 项目概述

本项目实现了一个简洁的实时排行榜系统，支持分数更新、自动排名计算和WebSocket实时广播。

## 核心架构

### 数据模型 (models/)
- **Leaderboard**: 核心排行榜数据结构
  - 使用读写锁保证并发安全
  - 自动排序和排名计算
  - 支持TopN查询和用户排名查询
- **PlayerScore**: 玩家分数信息
  - 用户ID、用户名、分数、排名
  - 时间戳记录

### 服务层 (services/)
- **ConnectionManager**: WebSocket连接管理
  - 连接注册和注销
  - 消息广播功能

### 处理器层 (handlers/)
- **WebSocketHandler**: WebSocket连接处理
  - 升级HTTP连接为WebSocket
  - 发送初始数据和维持连接
- **APIHandler**: REST API处理
  - `POST /api/update-score`: 更新分数
  - `GET /api/top`: 获取排行榜

### 前端界面 (static/)
- 简单的HTML界面
- 实时WebSocket连接
- 分数更新表单

## 技术特点

- **并发安全**: 使用sync.RWMutex保护共享数据
- **实时广播**: WebSocket推送排行榜更新
- **自动排序**: 分数变化后自动重新排名
- **RESTful API**: 简单易用的HTTP接口

## 使用方法

### 启动服务
```bash
go run main.go
```

访问 `http://localhost:8080`

### API接口

#### 更新分数
```bash
POST /api/update-score
{
  "user_id": "user1",
  "username": "Alice",
  "score": 1000
}
```

#### 获取排行榜
```bash
GET /api/top
```

## 测试

```bash
go test ./...
```