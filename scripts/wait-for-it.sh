#!/bin/sh

set -e

host=$(echo "$1" | cut -d : -f 1)
port=$(echo "$1" | cut -d : -f 2)

until timeout 1 bash -c "cat < /dev/null > /dev/tcp/$host/$port"; do
  echo "Waiting for $host:$port connection..."
  sleep 1
done

shift
exec "$@"