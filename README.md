# 🚀 Elasticsearch 数据检索系统

基于 Elasticsearch 的高性能数据检索系统，使用 Go Gin 框架构建，支持中文搜索、负载均衡、缓存优化和实时监控。

## ✨ 项目特点

- **⚡ 高性能搜索**：基于 Elasticsearch 8.11.0，支持高效的全文搜索和分词
- **🏗️ 低耦合架构**：采用分层架构设计，API 层、Service 层、Repository 层、Config 层
- **🔄 负载均衡**：使用 Nginx 实现后端服务负载均衡，提高系统可用性
- **💾 缓存优化**：集成 Redis 缓存，提高搜索响应速度
- **📊 实时监控**：包含用户行为分析、前端性能监控和错误监控
- **⚙️ 灵活配置**：使用 YAML 文件进行配置管理
- **🎨 响应式前端**：Google 风格的搜索界面，支持关键词高亮和结果分组

## 🛠️ 技术栈

- **后端**：Go 1.23, Gin 框架, Elasticsearch 8.11.0, Redis 6.0
- **前端**：HTML5, CSS3, JavaScript
- **部署**：Docker, Docker Compose, Nginx
- **监控**：Performance API, Navigation Timing API

## 🏛️ 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                     前端层                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │  搜索页面       │  │  性能监控       │  │  错误监控   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────┐
│                     负载均衡层                            │
│                    ┌─────────────────┐                    │
│                    │     Nginx       │                    │
│                    └─────────────────┘                    │
└─────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────┐
│                     后端服务层                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │  API 层         │  │  Service 层    │  │  中间件     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────┐
│                     数据层                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │  Elasticsearch  │  │    Redis       │  │  数据导入   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 🖥️ 搜索界面

![搜索界面](./front/image.png)

![搜索结果](./front/image1.png)

## 📚 数据集介绍

425万社区问答webtext2019zh知识类数据集，包含410万个预先过滤过的、高质量问题和回复。每个问题属于一个【话题】，总共有2.8万个各式话题，话题包罗万象。

### 📋 数据结构

```json
{
    "qid":<qid>,
    "title":<title>,
    "desc":<desc>,
    "topic":<topic>,
    "star":<star>,
    "content":<content>,
    "answer_id":<answer_id>,
    "answerer_tags":<answerer_tags>
}
```

其中:
- `qid`: 问题的唯一ID
- `title`: 问题的标题
- `desc`: 问题的描述（可为空）
- `topic`: 问题所属的话题
- `star`: 该回复的点赞个数
- `content`: 回复的内容
- `answer_id`: 回复的ID
- `answerer_tags`: 回复者所携带的标签

### 📝 数据样例

```json
{
  "qid": 20619381,
  "title": "张献忠血洗四川是否属实？",
  "desc": "", 
  "topic": "明朝", 
  "star": 3, 
  "content": "四川人历史上有三次大灭绝，现在的川人基本都是湖广填四川填过来的，所以我认为这个基本属实。",
  "answer_id": 17447047,
  "answerer_tags": "如是我闻"
},
{
  "qid": 36651654, 
  "title": "你发现了哪些基于个人经验的神秘规律？",
  "desc": "One 里看到的，看看问答社区的朋友们有什么更加有趣的规律。",
  "topic": "经验",
  "star": 22,
  "content": "去吃饭的路比吃饭回来的路长",
  "answer_id": 112831136,
  "answerer_tags": "大盈若冲"
}
```

## 🏃‍♂️ 快速开始

### 1. 📦 环境要求

- Docker 和 Docker Compose
- Go 1.23 或更高版本
- 至少 2GB 内存

### 2. 🚀 启动服务

在项目根目录执行：

```bash
docker-compose up -d
```

### 3. 📤 导入数据

当 Elasticsearch 服务启动后，运行以下命令导入数据：

```bash
cd backend
go run cmd/import.go
```

### 4. 🔧 启动后端服务

```bash
cd backend
go run main.go
```

### 5. 🌐 访问搜索界面

打开浏览器，访问：http://localhost:8080

## ⚙️ 配置说明

### 1. 📄 主配置文件

`backend/config/config.yaml`：

```yaml
server:
  addr: "127.0.0.1:8080"

elasticsearch:
  hosts:
    - "http://localhost:9200"

rate_limit:
  global:
    rps: 100
  search:
    rps: 50
    burst: 100

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  ttl: 3600
```

### 2. 🔀 Nginx 配置

`nginx.conf`：

```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server 127.0.0.1:8080;
        server 127.0.0.1:8081;
        server 127.0.0.1:8082;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        location /api/ {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## 📡 API 接口

### 1. 🔍 搜索接口

```
GET /api/search
```

**参数**：
- `index`: 索引名称（必填）
- `q`: 搜索关键词（必填）
- `fields`: 搜索字段，逗号分隔（默认：title,content）
- `page`: 页码（默认：1）
- `size`: 每页大小（默认：10）
- `sort`: 排序字段，支持多字段排序（例如：star,-timestamp）
- `filter`: 过滤条件（例如：category:tech）
- `highlight`: 是否高亮（默认：true）

**响应**：

```json
{
  "hits": {
    "total": {
      "value": 100
    },
    "hits": [
      {
        "_source": "{\"title\": \"...\", \"content\": \"...\"}",
        "highlight": {
          "title": ["...<em>关键词</em>..."],
          "content": ["...<em>关键词</em>..."]
        }
      }
    ]
  }
}
```

### 2. 📊 分析接口

#### 🔥 热门查询

```
GET /api/analytics/top-queries
```

**参数**：
- `limit`: 返回数量（默认：10）

#### 📈 搜索趋势

```
GET /api/analytics/trends
```

**参数**：
- `days`: 天数（默认：7）

#### ⚡ 性能监控

```
POST /api/analytics/performance
```

**请求体**：

```json
{
  "type": "page_load",
  "timestamp": "2026-03-01T00:00:00Z",
  "metrics": {
    "dnsLookup": 10,
    "tcpConnection": 20,
    "tlsHandshake": 30,
    "serverResponse": 100,
    "domContentLoaded": 200,
    "pageLoad": 300
  }
}
```

#### 🚨 错误监控

```
POST /api/analytics/error
```

**请求体**：

```json
{
  "type": "global_error",
  "message": "Error message",
  "stack": "Error stack trace",
  "timestamp": "2026-03-01T00:00:00Z",
  "context": {
    "filename": "script.js",
    "lineno": 10,
    "colno": 5
  }
}
```

## 📢 监控与分析

### 1. 👥 用户行为分析

- 记录用户搜索行为
- 分析热门查询和搜索趋势
- 提供数据可视化接口

### 2. ⚡ 前端性能监控

- 监控页面加载性能
- 监控搜索请求响应时间
- 提供性能数据上报和分析

### 3. 🚨 错误监控

- 监控前端 JavaScript 错误
- 监控未处理的 Promise 拒绝
- 提供错误数据上报和分析

## ❓ 常见问题

### 1. 🌐 网络连接问题

如果遇到网络连接问题，尝试：
- 检查网络连接
- 检查防火墙设置
- 尝试使用不同的 Docker 镜像源

### 2. 🡹 内存不足

Elasticsearch 需要至少 1GB 的内存。如果遇到内存不足的问题，尝试：
- 增加系统内存
- 修改 `ES_JAVA_OPTS` 环境变量，减少内存使用

### 3. 🔄 端口冲突

如果端口 9200 或 9300 已被占用，尝试：
- 停止占用这些端口的服务
- 修改 `docker-compose.yml` 文件中的端口映射

### 4. ❌ 数据导入失败

如果数据导入失败，尝试：
- 检查 Elasticsearch 服务是否正常运行
- 检查 JSON 文件格式是否正确
- 检查网络连接

## 🚀 部署方式

### 1. 🔧 开发环境

使用 Docker Compose 启动所有服务：

```bash
docker-compose up -d
```

### 2. 📦 生产环境

1. 构建后端服务镜像
2. 配置 Nginx 反向代理
3. 部署到生产服务器
4. 配置监控和告警

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支
3. 提交代码
4. 推送到远程分支
5. 创建 Pull Request

## 📄 许可证

MIT License

## 📞 联系方式

如有问题或建议，请联系项目维护者。
