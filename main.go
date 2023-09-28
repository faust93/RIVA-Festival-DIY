package main

import (
    "fmt"
    "log"
    "os/exec"
    "errors"
    "strings"
    "strconv"
    "time"
    "os"
    "os/signal"
    "syscall"
    "net/http"

    "github.com/warthog618/gpiod"
    )

// Mixer controls
const (
        volMixerID = 2
        volMin     = 0
        volMax     = 31
        DACreverse = 10
      )

// source modes
const (
        OFF     = 0
        UAC     = 1
        LINE_IN = 2
        SPDIF   = 3
      )

var audioSource int = OFF

// Buttons
var (
     btnOK     gpioButton
     btnUP     gpioButton
     btnDOWN   gpioButton
     btnLEFT   gpioButton
     btnRIGHT  gpioButton
     btnSWITCH gpioButton
    )

// LEDS
var (
     Red   *led
     Blue  *led
     Green *led
     White *led
     )


const on bool = true
const off bool = false

var volume int
var reversDAC bool = false
var btActive bool = false

var isPlaying bool
var hushTime time.Time

var ampGPIO *gpiod.Line
var ampState int

const NOTIFY_F = "/proc/asound/Codec/pcm0p/sub0/status"


func setVolume(vol int) {
    command := exec.Command("tinymix", "set", fmt.Sprintf("%d", volMixerID), fmt.Sprintf("%d", vol))
    err := command.Run()
    if err != nil {
        log.Println("Error setting volume:", err)
        return
    }
    volume = vol
}

func getVolume() {
    volume = 0

    command := exec.Command("tinymix", "get", fmt.Sprintf("%d", volMixerID))
    out, err := command.CombinedOutput()
    if err != nil {
        log.Println("Error getting volume:", err)
        return
    }

    svol := strings.Split(string(out), " ")[0]
    vol, err := strconv.Atoi(svol)
    if err != nil {
        log.Println("Error getting volume:", err)
        return
    }

    if vol > volMax {
        vol = volMax
    } else if vol < volMin {
        vol = volMin
    }

    volume = vol
}

func dacRevers(on bool) {
    vol := 0
    if on {
        vol = 1
    }
    command := exec.Command("tinymix", "set", fmt.Sprintf("%d", DACreverse), fmt.Sprintf("%d", vol), fmt.Sprintf("%d", vol))
    err := command.Run()
    if err != nil {
        log.Println("Error setting DAC reverse:", err)
    }
}


func okShortPress() {
    log.Println("DAC reversed")
    if reversDAC {
        reversDAC = false
        Green.setValue(0)
    } else {
        reversDAC = true
        Green.setValue(1)
    }
    dacRevers(reversDAC)
}

func okLongPress() {
    fmt.Println("OK long")
}

func upShortPress() {
    if volume < volMax {
        setVolume(volume+1)
    }
}

func upLongPress() {
    fmt.Println("UP long")
}

func downShortPress() {
    if volume > volMin {
        setVolume(volume-1)
    }
}

func downLongPress() {
    fmt.Println("DOWN long")
}

func leftShortPress() {
    fmt.Println("LEFT short")
}

func leftLongPress() {
    log.Println("BT", btActive)
    if btActive {
        if err := runCmd("./src_bt.sh stop"); err != 0 {
            log.Println("Unable to stop BT source")
        } else {
            Blue.setValue(0)
            btActive = false
        }
    } else {
        if err := runCmd("./src_bt.sh start"); err != 0 {
            log.Println("Unable to launch BT source")
        } else {
            Blue.setValue(1)
            btActive = true
        }
    }
}

// UAC
func rightShortPress() {
    switch audioSource {
        case SPDIF:
            srcSPDIF(off)
        case LINE_IN:
            srcLINEIN(off)
        case UAC:
            audioSource = OFF
            srcUAC(off)
            return
    }
    audioSource = UAC
    srcUAC(on)
}

// LINE_IN
func rightLongPress() {
    switch audioSource {
        case UAC:
            srcUAC(off)
        case SPDIF:
            srcSPDIF(off)
        case LINE_IN:
            audioSource = OFF
            srcLINEIN(off)
            return
    }
    audioSource = LINE_IN
    srcLINEIN(on)
}

// SPDIF
func rightLong2Press() {
    switch audioSource {
        case UAC:
            srcUAC(off)
        case LINE_IN:
            srcLINEIN(off)
        case SPDIF:
            audioSource = OFF
            srcSPDIF(off)
            return
    }
    audioSource = SPDIF
    srcSPDIF(on)
}

func srcUAC(on bool) {
    log.Println("UAC", on)
    if on {
        if err := runCmd("./src_uac.sh start"); err != 0 {
            log.Println("Unable to launch UAC1 source")
        } else {
            Blue.setValue(1)
            Red.setValue(1)
        }
    } else {
        runCmd("./src_uac.sh stop")
        Blue.setValue(0)
        Red.setValue(0)
    }
}

func srcLINEIN(on bool) {
    log.Println("LINEIN", on)
    if on {
        if err := runCmd("./src_linein.sh start"); err != 0 {
            log.Println("Unable to launch LINE_IN source")
        } else {
            Blue.setValue(1)
            Green.setValue(1)
        }
    } else {
        runCmd("./src_linein.sh stop")
        Blue.setValue(0)
        Green.setValue(0)
    }
}

func srcSPDIF(on bool) {
    log.Println("SPDIF", on)
    if on {
        if err := runCmd("./src_spdif.sh start"); err != 0 {
            log.Println("Unable to launch SPDIF source")
        } else {
            Red.setValue(1)
            Green.setValue(1)
        }
    } else {
        runCmd("./src_spdif.sh stop")
        Red.setValue(0)
        Green.setValue(0)
    }
}

func longX2Press() {
    fmt.Println("Long x2 press")
}

func switchOn() {
    fmt.Println("SWITCH on")
}

func switchOff() {
    fmt.Println("SWITCH off")
}

func runCmd(cmdStr string) int {
        cmd := exec.Command("sh", "-c", cmdStr)
        err := cmd.Run()
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            log.Println("Error launching command:", err)
            return exitErr.ExitCode()
        }
        return 0
}

func runCmdOut(cmdStr string) string {
        out, err := exec.Command("sh", "-c", cmdStr).CombinedOutput()
        if err != nil {
            log.Println("Error launching command:", err)
            return ""
        }
        return string(out)
}

func signalHandler() {
    signChan := make(chan os.Signal, 1)
    signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)

    <-signChan

    runCmd("./sysconfig.sh svc_stop");

    log.Println("Terminating")
    os.Exit(1)
}

func ampControl() {
    f, err := os.Open(NOTIFY_F)
    if err != nil {
        log.Fatal("Sound system error, no asound node found")
    }
    defer f.Close()

    buf := make([]byte, 16)
    nb_p, err := f.Read(buf)
    if err != nil {
        log.Fatal("Sound system error reading asound node")
    }

    for {
        _, err = f.Seek(0, 0)
        nb, err := f.Read(buf)
        if err != nil {
            log.Println("Sound system error reading asound node")
        }

        if nb != nb_p {
            if nb <= 7 {
                isPlaying = off
                hushTime = time.Now()
            } else {
                isPlaying = on
                if ampState == 0 {
                    ampState = 1
                    ampGPIO.SetValue(ampState)
                    White.setValue(1)
                    Green.setValue(0)
                    Red.setValue(0)
                    log.Println("Switching AMP on")
                }
            }
            nb_p = nb
        }
        time.Sleep(1 * time.Second)
    }
}

func ampPowerSave(timeout int) {
    for {
        time.Sleep(time.Duration(timeout) * time.Second)
        if isPlaying == off && ampState == 1 {
            diff := time.Now().Sub(hushTime)
            if diff.Seconds() > float64(timeout) && ampState == 1 {
                ampState = 0
                ampGPIO.SetValue(ampState)
                White.setValue(0)
                Green.setValue(1)
                Red.setValue(1)
                log.Println("Switching AMP off due to powersave threshold reached")
            }
        }
    }
}

func main() {

    loadConfig("config.toml")
    err := os.Chdir(conf.BasePath)
    if err != nil {
        log.Fatal("Unable to change home directory:", err)
    }

    Red   = ledInit("RED", "/sys/class/leds/riva:red", 0)
    Blue  = ledInit("BLUE", "/sys/class/leds/riva:blue", 0)
    Green = ledInit("GREEN", "/sys/class/leds/riva:green", 0)
    White = ledInit("WHITE", "/sys/class/leds/riva:white", 0)

    btnOK     := NewGPIOButton("OK", GPIO_OK, false, okShortPress, okLongPress, longX2Press)
    btnUP     := NewGPIOButton("UP", GPIO_UP, false, upShortPress, upLongPress, longX2Press)
    btnDOWN   := NewGPIOButton("DOWN", GPIO_DOWN, false, downShortPress, downLongPress, longX2Press)
    btnLEFT   := NewGPIOButton("LEFT", GPIO_LEFT, false, leftShortPress, leftLongPress, longX2Press)
    btnRIGHT  := NewGPIOButton("RIGHT", GPIO_RIGHT, false, rightShortPress, rightLongPress, rightLong2Press)
    btnSWITCH := NewGPIOButton("SWITCH", GPIO_SWITCH, true, switchOn, switchOff, nil)

    defer btnOK.Close()
    defer btnUP.Close()
    defer btnDOWN.Close()
    defer btnLEFT.Close()
    defer btnRIGHT.Close()
    defer btnSWITCH.Close()

    ampGPIO, err = gpiod.RequestLine("gpiochip1", 67, gpiod.AsOutput(1))
    if err != nil {
            log.Fatal("AMP GPIO RequestLine returned error: %w", err)
    }
    defer ampGPIO.Close()

    log.Println("Setting services")
    if err := runCmd("./sysconfig.sh svc_start"); err != 0 {
        log.Println("Services failed to start")
    }

    Green.blink(2)
    White.setValue(1)

    getVolume()

    ampState = 1
    isPlaying = off

    fs := http.FileServer(http.Dir("web/"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

    go signalHandler()
    go ampControl()
    go ampPowerSave(conf.AmpTimeout)

    http.HandleFunc("/", httpRoot)
    http.ListenAndServe(":80", nil)
}
