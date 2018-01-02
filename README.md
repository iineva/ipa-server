# ipa-server

Upload and install IPA in web.

* [中文文档](README_zh.md)

# Online Demo

<https://ipa.ineva.cn>

⚠️ Note About This Demo:

* For test only
* Server is deploy in China
* Bandwidth only 1Mb/s
* DO NOT USE THIS ON PRODUCTION

# Install

```
$ git clone https://github.com/iineva/ipa-server
$ cd ipa-server
$ docker-compose up -d
```

# Test

Open <http://<HOST_NAME>:9008> in your browser.

# Deploy

* This server is not included SSL certificate. It must run behide the reverse proxy with HTTPS.

* There is a simple way to setup a HTTPS with replace `docker-compose.yml` file:

```

# ***** Replace ALL <YOUR_DOMAIN> to you really domain *****

version: "2"

services:
  web:
    build: .
    container_name: ipa-server
    restart: always
    environment:
      - NODE_ENV=production
      - PUBLIC_URL=https://<YOUR_DOMAIN>
    ports:
      - "9008:8080"
    volumes:
      - "/docker/data/ipa-server:/app/upload"
  caddy:
    image: ineva/caddy:0.10.3
    restart: always
    network_mode: host
    entrypoint: |
      sh -c 'echo "$$CADDY_CONFIG" > /etc/Caddyfile && /usr/bin/caddy --conf /etc/Caddyfile --log stdout'
    environment:
      CADDY_CONFIG: |
        <YOUR_DOMAIN> {
          gzip
          proxy / localhost:9008
        }
```

* now you can access *https://\<YOUR_DOMAIN\>* in your browser.

Home | Detail |
 --- | ---
![](snapshot/en/1.jpeg) | ![](snapshot/en/2.jpeg)
