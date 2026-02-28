#!/bin/bash

dothink() {
    local i

    for i in {1..5}; do
        echo "I'm doing something $1-$i..."
        sleep "$2"
    done

}

[ -z "$ERROR" ] && ERROR=false
[ -z "$EXIT" ] && EXIT=false

$EXIT && ERROR=true

echo "** Starting deployment process… **"
sleep 1

# Emit our flag "==>" followed by the log message
for i in {1..6}; do
    echo "==> Processing things $i"
    dothink "$i" "$(awk 'BEGIN{srand(); print rand()}')"
    if [ "$i" = '3' ]; then
        $ERROR && >&2 echo "An error occurs"
        $EXIT && exit 1
    fi
done

echo "==> Deployment finished!"
sleep 0.5
