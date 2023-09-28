#!/bin/sh
# 31Hz 63Hz 125Hz 250Hz 500Hz 1kHz 2kHz 4kHz 8kHz 16kHz
EQ0="70 70 66 68 72 74 77 79 79 77"
EQ1="59 69 73 71 75 81 74 69 63 52"
EQ2="39 39 58 58 60 66 66 66 66 66"
EQ3="93 74 74 68 80 80 74 76 79 77"
EQ4="72 75 75 68 70 74 68 72 72 77"
EQ5="66 66 74 68 64 74 66 71 72 64"

eq_update() {
    amixer -D equal set '00. 31 Hz' ${1}
    amixer -D equal set '01. 63 Hz' ${2}
    amixer -D equal set '02. 125 Hz' ${3}
    amixer -D equal set '03. 250 Hz' ${4}
    amixer -D equal set '04. 500 Hz' ${5}
    amixer -D equal set '05. 1 kHz' ${6}
    amixer -D equal set '06. 2 kHz' ${7}
    amixer -D equal set '07. 4 kHz' ${8}
    amixer -D equal set '08. 8 kHz' ${9}
    amixer -D equal set '09. 16 kHz' ${10}
}

case "$1" in
    1)
        eq_update $EQ0
        ;;
    2)
        eq_update $EQ1
        ;;
    3)
        eq_update $EQ2
        ;;
    4)
        eq_update $EQ3
        ;;
    5)
        eq_update $EQ4
        ;;
    6)
        eq_update $EQ5
        ;;
    *)
        echo "Usage: $0 {1..5}"
        exit 1
esac