# Elasticsearch 服务设置指南

## 方法一：使用 Docker Compose

### 1. 配置 Docker Compose

在项目根目录创建 `docker-compose.yml` 文件：

```yaml
version: '3'

services:
  elasticsearch:
    image: elasticsearch:8.11.0
    container_name: elasticsearch
    environment:
      - discovery.type=single-node
      - ES_JAVA_OPTS=-Xms1g -Xmx1g
      - ELASTIC_PASSWORD=changeme
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
      - "9300:9300"
    volumes:
      - es_data:/usr/share/elasticsearch/data
    networks:
      - es_network

volumes:
  es_data:

networks:
  es_network:
    driver: bridge
```

### 2. 启动服务

```bash
docker-compose up -d
```

### 3. 检查服务状态

```bash
docker ps
```

### 4. 测试连接

```bash
curl http://localhost:9200
```

## 方法二：本地安装 Elasticsearch

### 1. 下载 Elasticsearch

从 [Elasticsearch 官网](https://www.elastic.co/downloads/elasticsearch) 下载适合你操作系统的版本。

### 2. 安装 Elasticsearch

#### Windows

1. 解压下载的 zip 文件到一个目录
2. 打开命令提示符，进入 Elasticsearch 目录
3. 运行 `bin\elasticsearch.bat` 启动服务

#### Linux/macOS

1. 解压下载的 tar 文件到一个目录
2. 打开终端，进入 Elasticsearch 目录
3. 运行 `bin/elasticsearch` 启动服务

### 3. 配置 Elasticsearch

编辑 `config/elasticsearch.yml` 文件，添加以下配置：

```yaml
cluster.name: my-cluster
node.name: node-1
discovery.type: single-node
```

### 4. 测试连接

```bash
curl http://localhost:9200
```

## 方法三：使用 Elastic Cloud

如果你无法在本地运行 Elasticsearch，可以考虑使用 [Elastic Cloud](https://www.elastic.co/cloud/) 提供的托管服务。

## 导入数据

当 Elasticsearch 服务启动后，运行以下命令导入数据：

```bash
cd backend
go run cmd/import.go
```

## 常见问题

### 1. 网络连接问题

如果遇到网络连接问题，尝试：
- 检查网络连接
- 检查防火墙设置
- 尝试使用不同的 Docker 镜像源

### 2. 内存不足

Elasticsearch 需要至少 1GB 的内存。如果遇到内存不足的问题，尝试：
- 增加系统内存
- 修改 `ES_JAVA_OPTS` 环境变量，减少内存使用

### 3. 端口冲突

如果端口 9200 或 9300 已被占用，尝试：
- 停止占用这些端口的服务
- 修改 `docker-compose.yml` 文件中的端口映射

### 4. 数据导入失败

如果数据导入失败，尝试：
- 检查 Elasticsearch 服务是否正常运行
- 检查 JSON 文件格式是否正确
- 检查网络连接
