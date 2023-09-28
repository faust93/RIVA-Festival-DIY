package main

import (
    "log"
    "os"
    "github.com/BurntSushi/toml"
)

type Config struct {
    BasePath    string     `toml:"root"`
    UACMode     int        `toml:"uacMode"`
    AirPlay2    bool       `toml:"airplay2"`
    Dlna        bool       `toml:"dlna"`
    CfgRebuild  bool       `toml:"rebuildCfg"`
    ApMode      bool       `toml:"apMode"`
    AmpTimeout  int        `toml:"ampTimeout"`
    EqPreset    int        `toml:"eqPreset"`
    Net         network    `toml:"network"`
}

type network struct {
    UsbEth      bool
    UsbIPaddr   string
    UsbNMask    string
    WManualSSID bool
    WlanSSID    string
    WlanPass    string
    WlanIPaddr  string
    WlanNMask   string
    WlanGW      string
    PriDNS      string
    SecDNS      string
    DdmsSSID    string
    DdmsPASS    string
}

var conf Config

func loadConfig(name string) {
    _, err := toml.DecodeFile(name, &conf)
    if err != nil {
        log.Fatal("Unable to load config: ", err)
    }
}

func saveConfig(name string) {
    fn, err := os.Create(name)
    if err != nil {
        log.Fatal("Unable to save config: ", err)
    }
    defer fn.Close()
    err = toml.NewEncoder(fn).Encode(conf)
    if err != nil {
        log.Fatal("Unable to encode config: ", err)
    }
}
