#!/bin/sh

/app/ipasd \
    -port "$PORT" \
    -public-url "${PUBLIC_URL:-$DOMAIN}" \
    -qiniu "$QINIU" \
    -qiniu-url "$QINIU_URL"
