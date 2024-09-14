FROM registry.cn-hangzhou.aliyuncs.com/ripper/nginx:latest

# 复制Nginx配置文件
COPY nginx.conf /etc/nginx/nginx.conf
COPY conf.d/default.conf /etc/nginx/conf.d/default.conf

# 复制SSL证书
COPY ssl/mycopilot.crt /etc/nginx/ssl/mycopilot.crt
COPY ssl/mycopilot.key /etc/nginx/ssl/mycopilot.key

# 设置正确的权限
RUN chmod 644 /etc/nginx/ssl/mycopilot.crt \
    && chmod 600 /etc/nginx/ssl/mycopilot.key

# 暴露443端口
EXPOSE 443

CMD ["nginx", "-g", "daemon off;"]