#!/bin/bash

KILL=

while getopts ":k" opt; do
    case $opt in
        k)
            KILL=1
            ;;
        *)
            ;;
    esac
done

# kill existing cluster
for f in $(find . -type f -name "*.pid"); do
    kill $(cat ${f})
    rm ${f}
done

[ "${KILL}" ] && exit 0

# run discovery server
pushd discover
go run *.go -pidfile ../discover.pid &
popd

# wait for startup
sleep 1

# register cluster
CLUSTER=$(curl localhost:8000/discover/new)
# run nodes
pushd node
for i in $(seq 1 5); do
    go run *.go -hostname node${i}.localhost -port 800${i} -pidfile ../node${i}.pid -token ${CLUSTER} &
done
popd

