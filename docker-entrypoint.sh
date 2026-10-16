#!/bin/sh
set -eu

PORT="${PORT:-10000}"

set -- -addr 0.0.0.0 -port "$PORT"

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

if [ "${DELETE_ENABLED:-false}" = "true" ] || [ "${DELETE_ENABLED:-false}" = "1" ]; then
  set -- "$@" -del
fi

if [ "${UPLOAD_DISABLED:-false}" = "true" ] || [ "${UPLOAD_DISABLED:-false}" = "1" ]; then
  set -- "$@" -upload-disabled
fi

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
