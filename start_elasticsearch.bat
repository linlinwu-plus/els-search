@echo off

REM 启动 Elasticsearch 服务
docker-compose up -d

REM 等待服务启动
timeout /t 10 /nobreak

REM 测试连接
echo Testing connection to Elasticsearch...
curl http://localhost:9200

REM 导入数据
echo Importing data...
cd backend
go run cmd/import.go

REM 启动应用
echo Starting application...
go run main.go
