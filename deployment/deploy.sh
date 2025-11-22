#!/bin/bash

# ============================================================
# NoFX 生产环境部署脚本
# ============================================================

set -e

echo "🚀 NoFX 生产环境部署"
echo "======================================"

# 检查是否在项目根目录
if [ ! -f "main.go" ]; then
    echo "❌ 错误: 请在项目根目录运行此脚本"
    exit 1
fi

# 部署方式选择
echo ""
echo "选择部署方式:"
echo "  1) 直接部署（编译并运行）"
echo "  2) systemd 服务"
echo "  3) Docker"
echo "  4) PM2"
echo ""
read -p "请选择 (1-4): " deploy_method

case $deploy_method in
    1)
        echo ""
        echo "📦 直接部署模式"
        echo "======================================"
        
        # 编译
        echo "🔨 编译应用..."
        go build -o nofx main.go
        
        echo "✅ 编译完成"
        echo ""
        echo "📝 下一步:"
        echo "  1. 设置环境变量:"
        echo "     export APP_ENV=production"
        echo "     export DATABASE_URL='postgresql://...'"
        echo ""
        echo "  2. 运行应用:"
        echo "     ./nofx"
        echo ""
        echo "  或使用 nohup 后台运行:"
        echo "     nohup ./nofx > nofx.log 2>&1 &"
        ;;
        
    2)
        echo ""
        echo "🔧 systemd 服务部署"
        echo "======================================"
        
        # 编译
        echo "🔨 编译应用..."
        go build -o nofx main.go
        
        # 创建部署目录
        DEPLOY_DIR="/opt/nofx"
        echo "📁 创建部署目录: $DEPLOY_DIR"
        sudo mkdir -p $DEPLOY_DIR
        
        # 复制文件
        echo "📋 复制文件..."
        sudo cp nofx $DEPLOY_DIR/
        sudo cp -r config.json $DEPLOY_DIR/ 2>/dev/null || true
        
        # 复制 systemd 服务文件
        echo "📝 安装 systemd 服务..."
        sudo cp deployment/nofx.service /etc/systemd/system/
        
        echo ""
        echo "⚠️  请编辑服务文件设置环境变量:"
        echo "     sudo nano /etc/systemd/system/nofx.service"
        echo ""
        read -p "按回车继续..."
        
        # 重载 systemd
        echo "🔄 重载 systemd..."
        sudo systemctl daemon-reload
        
        # 启用服务
        echo "✅ 启用服务..."
        sudo systemctl enable nofx
        
        # 启动服务
        echo "🚀 启动服务..."
        sudo systemctl start nofx
        
        # 查看状态
        echo ""
        echo "📊 服务状态:"
        sudo systemctl status nofx
        
        echo ""
        echo "📝 常用命令:"
        echo "  启动: sudo systemctl start nofx"
        echo "  停止: sudo systemctl stop nofx"
        echo "  重启: sudo systemctl restart nofx"
        echo "  状态: sudo systemctl status nofx"
        echo "  日志: sudo journalctl -u nofx -f"
        ;;
        
    3)
        echo ""
        echo "🐳 Docker 部署"
        echo "======================================"
        
        # 检查 Docker
        if ! command -v docker &> /dev/null; then
            echo "❌ 错误: 未安装 Docker"
            exit 1
        fi
        
        # 构建镜像
        echo "🔨 构建 Docker 镜像..."
        docker build -f docker/Dockerfile.backend -t nofx:latest .
        
        # 停止旧容器
        echo "🛑 停止旧容器..."
        docker stop nofx-production 2>/dev/null || true
        docker rm nofx-production 2>/dev/null || true
        
        echo ""
        echo "⚠️  请设置环境变量后启动容器:"
        echo ""
        echo "docker run -d \\"
        echo "  --name nofx-production \\"
        echo "  --restart always \\"
        echo "  -p 8080:8080 \\"
        echo "  -e APP_ENV=production \\"
        echo "  -e DATABASE_URL='postgresql://...' \\"
        echo "  nofx:latest"
        echo ""
        echo "或使用 docker-compose:"
        echo "  docker-compose -f deployment/docker-compose.production.yml up -d"
        ;;
        
    4)
        echo ""
        echo "⚡ PM2 部署"
        echo "======================================"
        
        # 检查 PM2
        if ! command -v pm2 &> /dev/null; then
            echo "❌ 错误: 未安装 PM2"
            echo "安装: npm install -g pm2"
            exit 1
        fi
        
        # 编译
        echo "🔨 编译应用..."
        go build -o nofx main.go
        
        # 使用 PM2 启动
        echo "🚀 使用 PM2 启动..."
        pm2 start deployment/ecosystem.config.js --env production
        
        # 保存 PM2 配置
        echo "💾 保存 PM2 配置..."
        pm2 save
        
        # 设置开机启动
        echo "🔧 设置开机启动..."
        pm2 startup
        
        echo ""
        echo "📊 PM2 状态:"
        pm2 status
        
        echo ""
        echo "📝 常用命令:"
        echo "  状态: pm2 status"
        echo "  日志: pm2 logs nofx-production"
        echo "  重启: pm2 restart nofx-production"
        echo "  停止: pm2 stop nofx-production"
        echo "  监控: pm2 monit"
        ;;
        
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac

echo ""
echo "======================================"
echo "✅ 部署完成！"
echo "======================================"
