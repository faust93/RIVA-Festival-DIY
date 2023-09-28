package main

import (
    "log"
    "time"
    "github.com/warthog618/gpiod"
    )

// GPIO buttons struct
type gpio struct {
    pin  int
    chip string
}

// GPIO buttons configuration
var (
    GPIO_OK       = gpio{203, "gpiochip1"}
    GPIO_UP       = gpio{ 12, "gpiochip1"}
    GPIO_DOWN     = gpio{ 00, "gpiochip1"}
    GPIO_LEFT     = gpio{ 02, "gpiochip1"}
    GPIO_RIGHT    = gpio{ 11, "gpiochip1"}
    GPIO_SWITCH   = gpio{ 11, "gpiochip0"}
)


type gpioButton struct {
    name string
    line *gpiod.Line
    state int
    swtch bool
    skipHandlers bool
    evtTime time.Time
    longPthr int64
    long2Pthr int64
    longPress func()
    long2Press func()
    shortPress func()
}

const (
      UP   = 0
      DOWN = 1
)

func NewGPIOButton(name string, bGPIO gpio, swtch bool, short func(), long func(), long2 func()) *gpioButton {
    button := &gpioButton{
        name:   name,
        state:  UP,
        swtch: swtch,
        skipHandlers: false,
        longPthr: 500,  // long press threshold im ms
        long2Pthr: 1000, // very long press threshold in ms
        longPress: long,
        long2Press: long2,
        shortPress: short,
    }
    var err error
    button.line, err = gpiod.RequestLine(bGPIO.chip, bGPIO.pin, gpiod.WithEventHandler(func(evt gpiod.LineEvent) {

        if evt.Type == gpiod.LineEventRisingEdge {
            if button.swtch && button.state != DOWN {
               button.state = DOWN
               button.shortPress()
               return
            }

            if button.state == DOWN {
                button.state = UP

                if button.skipHandlers {
                    return
                }

                diff := time.Now().Sub(button.evtTime)
                if diff.Milliseconds() > button.long2Pthr {
                    button.long2Press()
                } else if diff.Milliseconds() > button.longPthr {
                    button.longPress()
                } else {
                    button.shortPress()
                }
            }
        } else if evt.Type == gpiod.LineEventFallingEdge {
            if button.swtch {
               button.state = UP
               button.longPress()
               return
            }

            button.state = DOWN
            button.evtTime = time.Now()
        }
    }), gpiod.WithBothEdges) //, gpiod.AsActiveLow)

    if err != nil {
            log.Fatal("RequestLine returned error: %w", err)
    }

    return button
}

func (b *gpioButton) Value() int {
    var err error
    var val int
    val, err = b.line.Value()
    if err != nil {
        log.Println("Error getting GPIO value")
        return -1
    }
    return val
}

func (b *gpioButton) Close() error {
    if b.line != nil {
        log.Println("Closing line")
        _ = b.line.Close()
    }
    return nil
}

