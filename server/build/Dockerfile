FROM docker.linkos.org/openresty/openresty:1.25.3.2-5-alpine-fat

# 安装必要的工具 
RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.aliyun.com#g' /etc/apk/repositories
RUN apk add --no-cache curl bash

ENV PIKO_UPSTREAM_PORT=8022
ENV LISTEN_PORT=8080
# 从piko镜像复制二进制文件
COPY --from=ghcr.linkos.org/andydunstall/piko:latest /bin/piko /usr/local/bin/piko

# 设置工作目录
WORKDIR /app

# 复制启动脚本和配置文件
COPY build/start.sh /app/start.sh
COPY build/nginx.conf /etc/nginx/nginx.conf
COPY build/piko.conf /etc/nginx/conf.d/piko.conf

# 创建日志目录
RUN mkdir -p /var/log/nginx

# 设置脚本权限
RUN chmod +x /app/start.sh

# 暴露端口
EXPOSE 8088 8022 8023

# 启动命令
CMD ["/app/start.sh"]