#!/bin/sh

set -eu

should_auto_migrate() {
  case "${AUTO_MIGRATE:-true}" in
    1|true|TRUE|yes|YES|on|ON)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

if should_auto_migrate; then
  /app/idsai-migrate up
fi

exec /app/idsai
