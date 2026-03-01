# Elasticsearch 搜索服务

## 项目结构

```
els-search/
├── backend/        # 后端代码
│   ├── api/        # API 层，处理 HTTP 请求和响应
│   ├── config/     # 配置管理
│   ├── repository/ # 数据访问层，处理 Elasticsearch 操作
│   ├── service/    # 业务逻辑层
│   ├── go.mod      # 依赖管理
│   ├── main.go     # 应用入口
│   └── README.md   # 项目说明
└── front/          # 前端代码
```

## 技术栈

- Go 1.20
- Gin 框架
- Elasticsearch 8.x

## 模块设计

### 1. 配置层 (config)
- 负责加载和管理应用配置
- 支持从环境变量读取配置

### 2. 数据访问层 (repository)
- 定义 Elasticsearch 操作接口
- 实现 Elasticsearch 搜索功能
- 与 Elasticsearch 客户端直接交互

### 3. 业务逻辑层 (service)
- 处理搜索业务逻辑
- 构建 Elasticsearch 查询
- 调用 repository 层执行搜索

### 4. API 层 (api)
- 处理 HTTP 请求和响应
- 解析查询参数
- 调用 service 层执行搜索

## 低耦合设计

- **依赖倒置**：使用接口定义各层之间的依赖关系
- **分层架构**：各层职责明确，边界清晰
- **依赖注入**：通过构造函数注入依赖，便于测试和替换

## 流量控制

本项目实现了基于令牌桶算法的流量控制功能，可以有效防止系统过载。

### 流量控制配置

在 `config/config.yaml` 文件中配置流量控制参数：

- `rate_limit.global.rps`：全局每秒最大请求数
- `rate_limit.search.rps`：搜索接口每秒最大请求数
- `rate_limit.search.burst`：搜索接口突发请求数

### 流量控制中间件

- **RateLimiter**：基本的令牌桶限流中间件
- **BurstRateLimiter**：支持突发流量的令牌桶限流中间件
- **TimeWindowRateLimiter**：时间窗口限流中间件

### 响应状态码

当请求超过限流阈值时，系统会返回 `429 Too Many Requests` 状态码，并在响应体中包含重试建议。

## 安装和运行

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置 YAML 文件

编辑 `backend/config/config.yaml` 文件：

```yaml
server:
  addr: "localhost:8080"

elasticsearch:
  hosts:
    - "http://localhost:9200"
    - "http://localhost:9201"
    - "http://localhost:9202"

rate_limit:
  global:
    rps: 100  # 全局每秒最大请求数
  search:
    rps: 50   # 搜索接口每秒最大请求数
    burst: 100 # 搜索接口突发请求数
```

### 3. 运行应用

#### 3.1 单实例运行

```bash
cd backend
go run main.go
```

#### 3.2 多实例运行（负载均衡）

1. 复制配置文件并修改端口：

```bash
cd backend
cp config/config.yaml config/config_8081.yaml
cp config/config.yaml config/config_8082.yaml
```

编辑 `config/config_8081.yaml` 文件，将端口改为 8081：

```yaml
server:
  addr: "localhost:8081"
```

编辑 `config/config_8082.yaml` 文件，将端口改为 8082：

```yaml
server:
  addr: "localhost:8082"
```

2. 启动多个应用实例：

```bash
# 终端 1
cd backend
CONFIG_FILE=config/config.yaml go run main.go

# 终端 2
cd backend
CONFIG_FILE=config/config_8081.yaml go run main.go

# 终端 3
cd backend
CONFIG_FILE=config/config_8082.yaml go run main.go
```

3. 配置 Nginx 负载均衡：

编辑项目根目录下的 `nginx.conf` 文件，然后启动 Nginx：

```bash
nginx -c E:/code/T5/els-search/nginx.conf
```

4. 访问应用：

通过 Nginx 访问应用：`http://localhost/front/index.html`

## API 接口

### 搜索接口

- **URL**: `/api/search`
- **方法**: GET
- **参数**:
  - `index`: Elasticsearch 索引名（必填）
  - `q`: 搜索关键词（必填）
  - `fields`: 搜索字段，多个字段用逗号分隔（可选）
  - `page`: 页码，默认 1（可选）
  - `size`: 每页大小，默认 10（可选）

- **示例请求**:
  ```
  GET /api/search?index=products&q=phone&fields=name,description&page=1&size=10
  ```

- **响应**:
  ```json
  {
    "hits": {
      "total": {
        "value": 10
      },
      "hits": [
        {
          "_source": {}
        }
      ]
    }
  }
  ```

### 健康检查接口

- **URL**: `/health`
- **方法**: GET
- **响应**:
  ```json
  {
    "status": "ok"
  }
  ```
