#!/bin/sh
set -eu

set -- -addr 0.0.0.0 \
  -port "${PORT:-8080}"

PUBLIC_URL="${PUBLIC_URL:-${DOMAIN:-}}"

if [ -n "$PUBLIC_URL" ]; then
  set -- "$@" -public-url "$PUBLIC_URL"
fi

if [ -n "${REMOTE:-}" ]; then
  set -- "$@" -remote "$REMOTE"
fi

if [ -n "${REMOTE_URL:-}" ]; then
  set -- "$@" -remote-url "$REMOTE_URL"
fi

case "${DELETE_ENABLED:-false}" in
  true|1) set -- "$@" -del ;;
esac

case "${UPLOAD_DISABLED:-false}" in
  true|1) set -- "$@" -upload-disabled ;;
esac

if [ -n "${META_PATH:-}" ]; then
  set -- "$@" -meta-path "$META_PATH"
fi

if [ -n "${LOGIN_USER:-}" ]; then
  set -- "$@" -user "$LOGIN_USER"
fi

if [ -n "${LOGIN_PASS:-}" ]; then
  set -- "$@" -pass "$LOGIN_PASS"
fi

exec /app/ipasd "$@"
