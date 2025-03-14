#!/bin/bash
NETWORK="tp0_testing_net"
SERVER="server"
PORT=12345
MSG="Hello server im a test"

RESPONSE=$(
    docker run --rm --network "$NETWORK" \
        alpine sh -c " echo $MSG | nc $SERVER $PORT" \
)

#Alpine ya tiene netcat

if [ "$RESPONSE" = "$MSG" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi