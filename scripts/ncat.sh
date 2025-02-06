#!/bin/sh
PORT=${1:-8000}
if ! [ -x "$(command -v nc)" ]; then
	echo "netcat is missing. Please install it" >&2
	exit 1
fi
trap 'exit 0' INT

echo "Netcating calls on port $PORT"

while true; do nc -l -p $PORT <<EOF
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 2

OK
EOF
done
