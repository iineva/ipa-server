#!/bin/sh

ipasd_args=""

if [ -n "$PORT" ];then
    ipasd_args=$ipasd_args"-port $PORT "
fi

PUBLIC_URL=${PUBLIC_URL:-$DOMAIN}
if [ -n "$PUBLIC_URL" ];then
    ipasd_args=$ipasd_args"-public-url $PUBLIC_URL "
fi

if [ -n "$REMOTE" ];then
    ipasd_args=$ipasd_args"-remote $REMOTE "
fi

if [ -n "$REMOTE_URL" ];then
    ipasd_args=$ipasd_args"-remote-url $REMOTE_URL "
fi

if [ "$DELETE_ENABLED" = "true" -o "$DELETE_ENABLED" = "1" ];then
    ipasd_args=$ipasd_args"-del "
fi

if [ -n "$META_PATH" ];then
    ipasd_args=$ipasd_args"-meta-path $META_PATH "
fi

/app/ipasd $ipasd_args
