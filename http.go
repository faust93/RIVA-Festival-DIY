package main

import (
    "net/http"
    "html/template"
    "log"
    "fmt"
    "strings"
    "io/ioutil"
    "strconv"
)

type PageData struct {
    NeworkName  string
    NeworkPass  string
    DdmsSSID    string
    DdmsPASS    string
    ManualSSID  bool
    PriDNS      string
    SecDNS      string
    NetIP       string
    NetMask     string
    NetGW       string
    Uptime      string
    AirPlay2    bool
    Dlna        bool
    UACMode     int
    AmpTimeout  int
    EqPreset    int
    USBEth      bool
    WNetworks   []string
    Temp        string
}

func genPageData() PageData {
    upst := runCmdOut("uptime")
    wnets := scanNetworks()
    tmp := readTemp()

    return PageData{
        NeworkName: conf.Net.WlanSSID,
        NeworkPass: conf.Net.WlanPass,
        DdmsSSID:   conf.Net.DdmsSSID,
        DdmsPASS:   conf.Net.DdmsPASS,
        ManualSSID: conf.Net.WManualSSID,
        PriDNS:     conf.Net.PriDNS,
        SecDNS:     conf.Net.SecDNS,
        NetIP:      conf.Net.WlanIPaddr,
        NetMask:    conf.Net.WlanNMask,
        NetGW:      conf.Net.WlanGW,
        Uptime:     upst,
        AirPlay2:   conf.AirPlay2,
        Dlna:       conf.Dlna,
        UACMode:    conf.UACMode,
        AmpTimeout: conf.AmpTimeout,
        EqPreset:   conf.EqPreset,
        USBEth:     conf.Net.UsbEth,
        WNetworks:  wnets,
        Temp:       tmp,
    }
}

func readTemp() string {
    t, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
    if err != nil {
        log.Println("Unable to obtain SOC temp")
        return "0.0"
    }
    base := t[:2]
    hund := t[2:]
    return (string(base) + "." + string(hund))
}

func scanNetworks() []string {
    nws := []string{}
    cmdOut := runCmdOut("iw wlan0 scan | grep SSID: | cut -f2 -d' '")
    for _, line := range strings.Split(cmdOut, "\n") {
        if len(line) > 0 {
            //log.Println(line)
            nws = append(nws, line)
        }
    }
    return nws
}

func httpRoot(w http.ResponseWriter, r *http.Request) {

   switch r.Method {
    case "GET":
        data := genPageData()
        tmpl, err := template.ParseFiles("web/index.html")
        if err != nil {
            log.Println("Error parsing index.html:", err)
            return
        }
        err = tmpl.Execute(w, data)
        if err != nil {
            log.Println("Error executing template:", err)
            return
        }
    case "POST":
        if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
        switch r.URL.Path {
            case "/Reboot":
                runCmd("/sbin/reboot");

            case "/HandleCfg":
                val := r.FormValue("cfg_dlna")
                if val == "on" {
                    conf.Dlna = true
                } else {
                    conf.Dlna = false
                }
                val = r.FormValue("cfg_air2")
                if val == "on" {
                    conf.AirPlay2 = true
                } else {
                    conf.AirPlay2 = false
                }
                val = r.FormValue("cfg_uac2")
                if val == "on" {
                    conf.UACMode = 2
                } else {
                    conf.UACMode = 1
                }
                val = r.FormValue("cfg_ampt")
                tmout, err := strconv.Atoi(val)
                if err != nil {
                    log.Println("Error wrong amp timeout value:", err)
                    return
                }
                conf.AmpTimeout = tmout

                saveConfig("config.toml")

                log.Println("Restarting services")
                if err := runCmd("./sysconfig.sh svc_restart"); err != 0 {
                    log.Println("Unable to restart services")
                }

                http.Redirect(w, r, "/", 302)

            case "/HandleEq":
                val := r.FormValue("eq")
                log.Println("EQ:", val)
                eqN, err := strconv.Atoi(val)
                if err != nil {
                    log.Println("Error wrong eq preset value:", err)
                    return
                }
                conf.EqPreset = eqN
                saveConfig("config.toml")
                runCmd("./eq.sh " + val)
                http.Redirect(w, r, "/", 302)

            case "/HandleDDMS":
                ddmsName := r.FormValue("DDMSIDName")
                ddmsPass := r.FormValue("DDMSPassword")
                log.Println("DDMSNAME:", ddmsName)
                log.Println("DDMSPASS:", ddmsPass)
                conf.Net.DdmsSSID = ddmsName
                conf.Net.DdmsPASS = ddmsPass
                conf.CfgRebuild = true

                saveConfig("config.toml")
                http.Redirect(w, r, "/", 302)

            case "/HandleNW":
                val := r.FormValue("SSID")
                log.Println("SSID:", val)
                conf.Net.WlanSSID = val

                val = r.FormValue("configure_SSIDcheckbox")
                if val == "on" {
                    val = r.FormValue("SSIDManual")
                    log.Println("Manual SSID:", val)
                    conf.Net.WlanSSID = val
                    conf.Net.WManualSSID = true
                } else {
                    conf.Net.WManualSSID = false
                }

                nManual := r.FormValue("checkboxConfigure")
                if nManual == "on" {
                    val = r.FormValue("StaticIP")
                    log.Println("IP Manual:", val)
                    conf.Net.WlanIPaddr = val

                    val = r.FormValue("NetMask")
                    log.Println("NetMask Manual:", val)
                    conf.Net.WlanNMask = val

                    val = r.FormValue("Gateway")
                    log.Println("GW Manual:", val)
                    conf.Net.WlanGW = val

                } else {
                    conf.Net.WlanIPaddr = ""
                    conf.Net.WlanNMask = ""
                    conf.Net.WlanGW = ""
                }

                val = r.FormValue("PrimaryDNS")
                log.Println("DNS1 Manual:", val)
                conf.Net.PriDNS = val

                val = r.FormValue("SecondaryDNS")
                log.Println("DNS2 Manual:", val)
                conf.Net.SecDNS = val
                conf.CfgRebuild = true

                saveConfig("config.toml")
                http.Redirect(w, r, "/", 302)
        }

    default:
        fmt.Fprintf(w, "Only GET and POST methods are supported.")
    }
}
