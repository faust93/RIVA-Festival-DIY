package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "time"
    "strconv"
    )

type led struct {
    state int
    sysfsnode string
    name string
}

func ledInit(name string, node string, state int) *led {
    log.Println("LED: Initialising", name, "at", node, "with state", state)

    err := ioutil.WriteFile(node + "/brightness", []byte(fmt.Sprintf("%d", state)), 0)
    if err != nil {
        log.Fatal(err)
    }

    return &led{
        state: state,
        sysfsnode: node,
        name: name,
    }
}

func (led *led) setValue(value int) {
    err := ioutil.WriteFile(led.sysfsnode + "/brightness", []byte(fmt.Sprintf("%d", value)), 0)
    if err != nil {
        log.Println(err)
    }
}

func (led *led) getValue() int {
    data, err := ioutil.ReadFile(led.sysfsnode + "/brightness")
    if err != nil {
        log.Println("Error getting led state:", err)
        return 0
    }

    val, err := strconv.Atoi(string(data[0]))
    if err != nil {
        log.Println("Error getting led state:", err)
        return 0
    }
    return val
}

func (led *led) blink(times int) {
    for ; times > 0; times-- {
        led.setValue(1)
        time.Sleep(time.Millisecond * 100)
        led.setValue(0)
        time.Sleep(time.Millisecond * 100)
    }
}
