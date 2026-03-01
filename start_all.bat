@echo off

REM 启动 Elasticsearch 服务
echo 启动 Elasticsearch 服务...
docker-compose up -d

REM 等待 Elasticsearch 服务启动
echo 等待 Elasticsearch 服务启动...
timeout /t 15 /nobreak

REM 导入数据
echo 导入数据...
cd backend
go run cmd/import.go

REM 启动后端服务
echo 启动后端服务...
start "Backend Server" go run main.go

REM 启动前端服务（使用 Python 内置服务器）
echo 启动前端服务...
cd ../front
start "Frontend Server" python -m http.server 8000

echo 服务启动完成！
echo 前端地址: http://localhost:8000
echo 后端地址: http://localhost:8081

pause
