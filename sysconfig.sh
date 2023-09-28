#!/bin/sh

RIVA_HOME="/root/riva"

IFAC=/etc/network/interfaces
WPA_CONF=/etc/wpa_supplicant/wpa_supplicant.conf
RESOLV=/etc/resolv.conf

SWITCH=$(gpioget 0 11)

AP_MODE=$(grep -w "apMode" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed 's/ //g')
CFG_REBUILD=$(grep -w "rebuildCfg" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed 's/ //g')

WLAN_SSID=$(grep -w "WlanSSID" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')

DLNA=$(grep -w "dlna" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ //g')
AIR2=$(grep -w "airplay2" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ //g')

sysconf_update() {
    WLAN_IP=$(grep -w "WlanIPaddr" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    WLAN_PASS=$(grep -w "WlanPass" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    WLAN_MASK=$(grep -w "WlanNMask" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    WLAN_GW=$(grep -w "WlanGW" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    DDMS_SSID=$(grep -w "DdmsSSID" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    DDMS_PASS=$(grep -w "DdmsPASS" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    DNS1=$(grep -w "PriDNS" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')
    DNS2=$(grep -w "SecDNS" $RIVA_HOME/config.toml | cut -f2 -d'=' | sed -E 's/ |"//g')

    echo -e "auto lo\n""iface lo inet loopback\n" > $IFAC
    if [ -n "$WLAN_SSID" ]
    then
    if [ "$AP_MODE" = "true" ]; then
        echo -e "auto wlan0\n" >> $IFAC
        echo -e "iface wlan0 inet static" >>$IFAC
        echo -e "\taddress 192.168.3.1" >>$IFAC
        echo -e "\tnetmask 255.255.255.0\n" >>$IFAC
        cp -f $RIVA_HOME/hostapd.conf /etc/hostapd.conf
        sed -i -e "s/#SSID#/$DDMS_SSID/g" /etc/hostapd.conf
        sed -i -e "s/#PASSWD#/$DDMS_PASS/g" /etc/hostapd.conf
    else
        if [ -n "$WLAN_PASS" ]
        then
             wpa_passphrase $WLAN_SSID $WLAN_PASS > $WPA_CONF
        else
            echo "network={" > $WPA_CONF
            echo "    ssid=\"$WLAN_SSID\"" >> $WPA_CONF
            echo "    key_mgmt=NONE" >> $WPA_CONF
            echo "}" >> $WPA_CONF
        fi

        echo -e "auto wlan0" >> $IFAC
        if [ -n "$WLAN_IP" ]
        then
            eval $(ipcalc --broadcast ${WLAN_IP} ${WLAN_MASK})
            echo -e "iface wlan0 inet static" >> $IFAC
            echo -e "\taddress $WLAN_IP\n""\tnetmask $WLAN_MASK\n""\tbroadcast $BROADCAST" >> $IFAC
            echo -e "\thwaddress ether B8:27:4C:55:61:E4" >> $IFAC
            echo -e "\thostname riva" >> $IFAC
            if [ -n "$WLAN_GW" ]; then
                echo -e "\tgateway $WLAN_GW" >> $IFAC
            fi
        else
            echo -e "iface wlan0 inet dhcp" >> $IFAC
            echo -e "\thwaddress ether B8:27:4C:55:61:E4" >> $IFAC
            echo -e "\thostname riva" >> $IFAC
        fi
    fi
    fi
    if [ -n "$DNS1" ]; then
        echo -e "search lan" > $RESOLV
        echo -e "nameserver $DNS1" >> $RESOLV
    fi
    if [ -n "$DNS2" ]; then
        echo -e "nameserver $DNS2" >> $RESOLV
    fi

}


case "$1" in
  start)
        if [ "$SWITCH" = "1" -a "$AP_MODE" = "false" ]; then
            CFG_REBUILD="true"
            AP_MODE="true"
            sed -i -e "s/apMode = false/apMode = true/g" $RIVA_HOME/config.toml
        fi

        if [ "$SWITCH" = "0" -a "$AP_MODE" = "true" ]; then
            CFG_REBUILD="true"
            AP_MODE="false"
            sed -i -e "s/apMode = true/apMode = false/g" $RIVA_HOME/config.toml
        fi

        if [ "$CFG_REBUILD" = "true" ]; then
            sysconf_update
            sed -i -e "s/rebuildCfg = true/rebuildCfg = false/g" $RIVA_HOME/config.toml
        fi

        echo "Starting network.."
        [[ "$CFG_REBUILD" = "true" ]] && /sbin/ifdown -a
        /sbin/ifup -a
        if [ -n "$WLAN_SSID" ]; then
            if [ "$AP_MODE" = "true" ]; then
                killall udhcpc
                start-stop-daemon -S -q -m -p /var/run/hostapd.pid -x /usr/sbin/hostapd -- -B /etc/hostapd.conf
                start-stop-daemon -S -q -m -p /var/run/udhcpdwlan.pid -x /usr/sbin/udhcpd -- $RIVA_HOME/dhcpdwlan.conf
            else
                start-stop-daemon -S -q -m -p /var/run/wpa_supplicant.pid -x /sbin/wpa_supplicant -- -B -i wlan0 -c $WPA_CONF
                /etc/init.d/chronyd start &
            fi
        fi
        ;;

  stop)
        echo "Stopping network.."
        pidof chronyd
        [ $? = 0 ] && /etc/init.d/chronyd stop
        start-stop-daemon -K -q -p /var/run/wpa_supplicant.pid
        pidof udhcpc
        [ $? = 0 ] && killall udhcpc
        pidof hostapd
        [ $? = 0 ] && killall hostapd
        start-stop-daemon -K -q -p /var/run/udhcpdwlan.pid
        start-stop-daemon -K -q -p /var/run/hostapd.pid
        /sbin/ifdown -a
        ;;

  restart|reload)
    "$0" stop
    "$0" start
    ;;

  svc_start)
        echo "Starting services.."
        $RIVA_HOME/udc.sh start
        if [ "$AIR2" = "true" ]; then
            /etc/init.d/shairport2 start
        else
            /etc/init.d/shairport start
        fi
        if [ "$DLNA" = "true" ]; then
            (sleep 60 ; /etc/init.d/gmrenderer start) &
        fi
        ;;

  svc_stop)
        echo "Stopping services.."
        $RIVA_HOME/udc.sh stop
        /etc/init.d/gmrenderer stop
        /etc/init.d/shairport2 stop
        /etc/init.d/shairport stop
        ;;

  svc_restart)
    "$0" svc_stop
    "$0" svc_start
    ;;

  *)
    echo "Usage: $0 {start|stop|restart|svc_start|svc_stop|svc_restart}"
    exit 1
esac

exit $?