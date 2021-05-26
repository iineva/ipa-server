# ipa-server

ipa-server is updated to v2, to [older version v1](https://github.com/iineva/ipa-server/tree/v1)

Upload and install IPA in web.

* [中文文档](README_zh.md)

Home | Detail |
 --- | ---
![](snapshot/en/1.jpeg) | ![](snapshot/en/2.jpeg)


# Install for local trial

```shell
# clone
git clone https://github.com/iineva/ipa-server
# build and start
cd ipa-server
docker-compose up -d
# than open http://localhost:9008 in your browser.
```

# Heroku Deploy

### config

* PUBLIC_URL: public URL for this server, empty to use `$DOMAIN`
* QINIU: qiniu config `AK:SK:[ZONE]:BUCKET`
* QINIU_URL: qiniu bucket public url, https://cdn.example.com
* ALIOSS: alioss config `ENDPOINT:ID:SECRET:BUCKET`
* ALIOSS_URL: alioss bucket public url, https://xxxx.oss-cn-shenzhen.aliyuncs.com
* DELETE_ENABLED: delete app enabled, `true` `false`

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/iineva/ipa-server)


# Docker Deploy

* This server is not included SSL certificate. It must run behide the reverse proxy with HTTPS.

* After deployed, you can access *https://\<YOUR_DOMAIN\>* in your browser.

* There is a simple way to setup a HTTPS with replace `docker-compose.yml` file:

```

# ***** Replace ALL <YOUR_DOMAIN> to you really domain *****

version: "2"

services:
  web:
    image: ineva/ipa-server:latest
    container_name: ipa-server
    restart: unless-stopped
    environment:
      # server public url
      - PUBLIC_URL=https://<YOUR_DOMAIN>
      # option, qiniu config AK:SK:[ZONE]:BUCKET
      - QINIU=
      # option, qiniu public url
      - QINIU_URL=
      # option, alicloud OSS config ENDPOINT:ID:SECRET:BUCKET
      - ALIOSS=
      # option, alioss public url
      - ALIOSS_URL=
      # option, metadata storage path, use random secret path to keep your metadata safer in case of remote storage
      - META_PATH=appList.json
      # delete app enabled, true/false
      - DELETE_ENABLED="false"
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

# Build or run from source code

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
