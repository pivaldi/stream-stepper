#!/usr/bin/env bash

echo "Starting deployment process..."
sleep 1

dothink() {
    local i

    for i in {1..5}; do
        echo "I'm doing something $1-$i..."
        sleep "$2"
    done

}

# Emit our flag "==>" followed by the log message
for i in {1..6}; do
    echo "==> Processing things $i"
    dothink "$i" "$(awk 'BEGIN{srand(); print rand()}')"
    if [ "$i" = '5' ]; then
        >&2 echo "An error occurs"
        # exit 1
    fi
done

echo "==> Deployment finished!"
sleep 0.5
