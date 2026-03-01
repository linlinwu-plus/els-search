#!/bin/bash

# 启动 Elasticsearch 服务
echo "Starting Elasticsearch service..."
docker-compose up -d

# 等待 Elasticsearch 服务启动
echo "Waiting for Elasticsearch to start..."
sleep 10

# 导入数据
echo "Importing data..."
cd backend
go run cmd/import.go

# 启动应用
echo "Starting application..."
go run main.go
