@echo off

REM 启动 Elasticsearch 服务
echo Starting Elasticsearch service...
docker-compose up -d

REM 等待 Elasticsearch 服务启动
echo Waiting for Elasticsearch to start...
timeout /t 10 /nobreak

REM 导入数据
echo Importing data...
cd backend
go run cmd/import.go

REM 启动应用
echo Starting application...
go run main.go
