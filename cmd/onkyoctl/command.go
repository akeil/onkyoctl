package main

import (
    "log"
    "os"
	"os/signal"

    "akeil.net/akeil/onkyoctl"
)

func main() {

    dev := onkyoctl.NewDevice("192.168.3.142")
    //runOnce(dev)
    runForever(dev)
}

func runOnce(dev onkyoctl.Device) {

    err := dev.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer dev.Stop()

    // send command
    dev.SendCommand(onkyoctl.ISCPCommand("MVLQSTN"))

    // wait for the reply or time out
}

func runForever(dev onkyoctl.Device) {
    err := dev.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer dev.Stop()

    dev.SendCommand(onkyoctl.ISCPCommand("MVLUP"))
    dev.SendCommand(onkyoctl.ISCPCommand("PWRQSTN"))
    dev.SendCommand(onkyoctl.ISCPCommand("MVLQSTN"))
    //dev.SendCommand(onkyoctl.ISCPCommand("PWR00"))

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop  // wait for SIGINT
}
