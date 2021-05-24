# ipa-server

ipa-server 已经更新到v2, 使用golang重构, [老版本v1](https://github.com/iineva/ipa-server/tree/v1)

使用浏览器上传和部署 `.ipa` 文件

* 自动识别ipa包内信息
* 自动生成图标
* 开箱即用
* 只需要一台低配云主机, 一个域名
* 支持生成文件完全存储在外部存储，目前支持 `七牛对象存储`
* 单二进制文件即可运行，编译后体积仅`10M`左右

Home | Detail |
 --- | ---
![](snapshot/en/1.jpeg) | ![](snapshot/en/2.jpeg)


# 安装本地试用

```shell
# clone
git clone https://github.com/iineva/ipa-server
# build and start
cd ipa-server
docker-compose up -d
# 启动后在浏览器打开 http://localhost:9008
```

# Heroku 部署

### 配置

* PUBLIC_URL: 本服务的公网URL, 如果为空试用Heroku默认的 `$DOMAIN`
* QINIU: 七牛配置 `AK:SK:[ZONE]:BUCKET`, `ZONE` 区域参数可选, 绝大多数情况可以自动检测
* QINIU_URL: 七牛CDN对应的URL，注意需要开启HTTPS支持才能正常安装！例子：https://cdn.example.com

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)


# 正式部署

* 本仓库代码不包含SSL证书部分，由于苹果在线安装必须具备HTTPS，所以本程序必须运行在HTTPS反向代理后端。

* 部署后，你可以使用浏览器访问 *https://\<YOUR_DOMAIN\>*

* 最简单的办法开启完整服务，使用下面的配置替换 `docker-compose.yml` 文件:

```

# ***** 更换所有 <YOUR_DOMAIN> 成你的真实域名 *****

version: "2"

services:
  web:
    image: ineva/ipa-server:v2.0
    container_name: ipa-server
    restart: unless-stopped
    environment:
      # 本服务公网IP
      - PUBLIC_URL=https://<YOUR_DOMAIN>
      # option, 七牛配置 AK:SK:[ZONE]:BUCKET
      - QINIU=
      # option, 七牛DNS域名，注意要加 https://
      - QINIU_URL=
      # option, 元数据存储路径, 使用一个随机路径来保护元数据，因为在使用远程存储的时候，没有更好的方法防止外部直接访问元数据文件
      - META_PATH=appList.json
    volumes:
      - "/docker/data/ipa-server:/app/upload"
  caddy:
    image: abiosoft/caddy:0.11.5
    restart: always
    ports:
      - "80:80"
      - "443:443"
    entrypoint: |
      sh -c 'echo "$$CADDY_CONFIG" > /etc/Caddyfile && /usr/bin/caddy --conf /etc/Caddyfile --log stdout'
    environment:
      CADDY_CONFIG: |
        <YOUR_DOMAIN> {
          gzip
          proxy / web:8080
        }
```

# 源码编译

```shell
# install golang v1.16 first
git clone https://github.com/iineva/ipa-server
# build and start
cd ipa-server
# build binary
make build
# run local server
make
```

# TODO

- [ ] 设计全新的鉴权方式，初步考虑试用GitHub登录鉴权
- [x] 支持七牛存储
- [x] 兼容v1产生数据，无缝升级
- [ ] 支持命令行生成静态文件部署
