#!/bin/sh

case "$1" in
    start)
        echo host > /sys/devices/platform/soc/1c19000.usb/musb-hdrc.4.auto/mode
        DEVID=$(aplay -l | grep USB | cut -f2 -d' ' | tr -d :)
        tinymix -D $DEVID set 6 0
        tinymix -D $DEVID set 15 2
        {
            while true
            do
                #alsaloop -C hw:1,0 -P default -t 10000 -A 3 -S 1 -b -n
                arecord -f S16_LE -t wav -r 48000 -c 2 -D hw:$DEVID | aplay -r 48000 -c 2 -D default -t wav
                sleep 1
            done
        } >/dev/null 2>&1 &
        exit 0
        ;;

    stop)
        killall aplay
        killall arecord
        killall `basename $0`
        ;;
esac
