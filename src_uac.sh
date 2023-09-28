#!/bin/sh

RIVA_HOME="/root/riva"
UAC=$(grep -w "uacMode" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed 's/ //g')

SFMT="S16_LE"
SRATE=48000
SCHAN=2

if [ "$UAC" == "2" ]; then
SFMT="S32_LE"
fi

case "$1" in
    start)
        echo peripheral > /sys/devices/platform/soc/1c19000.usb/musb-hdrc.4.auto/mode
        sleep 1
        DEVID=`aplay -l | grep UAC | cut -f2 -d' ' | tr -d :`
        {
            while true
            do
                #alsaloop -C hw:1,0 -P default -t 10000 -A 3 -S 1 -b -n
                arecord -f $SFMT -t wav -r $SRATE -c $SCHAN -D hw:$DEVID | aplay -r $SRATE -c $SCHAN -D default -t wav
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
