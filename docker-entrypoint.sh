#!/bin/sh

ipasd_args=""

if [ -n "$PORT" ];then
    ipasd_args=$ipasd_args"-port $PORT "
fi

PUBLIC_URL=${PUBLIC_URL:-$DOMAIN}
if [ -n "$PUBLIC_URL" ];then
    ipasd_args=$ipasd_args"-public-url $PUBLIC_URL "
fi

if [ -n "$QINIU" ];then
    ipasd_args=$ipasd_args"-qiniu $QINIU "
fi

if [ -n "$QINIU_URL" ];then
    ipasd_args=$ipasd_args"-qiniu-url $QINIU_URL "
fi

if [ -n "$ALIOSS" ];then
    ipasd_args=$ipasd_args"-alioss $ALIOSS "
fi

if [ -n "$ALIOSS_URL" ];then
    ipasd_args=$ipasd_args"-alioss-url $ALIOSS_URL "
fi

if [ -n "$META_PATH" ];then
    ipasd_args=$ipasd_args"-meta-path $META_PATH "
fi

echo $ipasd_args

./ipasd $ipasd_args