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
    dev.SendISCP(onkyoctl.ISCPCommand("MVLQSTN"))

    // wait for the reply or time out
}

func runForever(dev onkyoctl.Device) {
    err := dev.Start()
    if err != nil {
        log.Fatal(err)
    }
    defer dev.Stop()

    dev.SendISCP(onkyoctl.ISCPCommand("PWRQSTN"))
    dev.SendISCP(onkyoctl.ISCPCommand("MVLQSTN"))
    err = dev.SendCommand("volume", "up")
    if err != nil {
        log.Printf("Error sending command: %v", err)
    }

    err = dev.SendCommand("mute", "off")
    if err != nil {
        log.Printf("Error sending command: %v", err)
    }

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop  // wait for SIGINT
}
