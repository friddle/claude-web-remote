#!/bin/bash

# 读取环境变量
PIKO_UPSTREAM_PORT=${PIKO_UPSTREAM_PORT:-8022}
LISTEN_PORT=${LISTEN_PORT:-8088}

echo "启动服务..."
echo "PIKO_UPSTREAM_PORT: $PIKO_UPSTREAM_PORT"
echo "LISTEN_PORT: $LISTEN_PORT"

sed -i "s/{{LISTEN_PORT}}/$LISTEN_PORT/g" /etc/nginx/conf.d/piko.conf

# 启动piko server
echo "启动piko server..."
if [ -z "$PIKO_TOKEN" ]; then   
    piko server --upstream.bind-addr ":$PIKO_UPSTREAM_PORT" --proxy.bind-addr ":8023" &
else
    piko server --upstream.bind-addr ":$PIKO_UPSTREAM_PORT" --proxy.bind-addr ":8023" --token $PIKO_TOKEN &
fi

# 等待piko启动
sleep 2

# 启动nginx
echo "启动nginx..."
nginx -g "daemon off;"

