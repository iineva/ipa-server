FROM node:8.4.0
MAINTAINER Steven <s@beeeye.cn>

# 设置工作目录
WORKDIR /app

# 安装依赖包,Copy代码
COPY package.json .
COPY package-lock.json .
RUN npm install --production --registry=https://registry.npm.taobao.org
COPY . .

VOLUME /app/upload

CMD ["node", "index.js"]
