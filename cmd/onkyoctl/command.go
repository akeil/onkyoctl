package main

import (
    "log"
    "fmt"
    "os"
	"os/signal"

    "gopkg.in/alecthomas/kingpin.v2"

    "akeil.net/akeil/onkyoctl"
)

func main() {
    // command line interface is:
    // PROG watch
    // PROG status
    // PROG <name> <param>
    //
    // with optional args --host and --port

    app := kingpin.New("onkyo", "Control Onky receiver.")
    app.HelpFlag.Short('h')

    var host = app.Flag("host", "Hostname or IP address").String()
    var port = app.Flag("port", "Port number").Default("60128").Short('p').Int()

    do := app.Command("do", "Execute a command").Default()
    var name = do.Arg("name", "The proprety to change").Required().String()
    var value = do.Arg("value", "The value to set").String()

    status := app.Command("status", "Show device status")
    watch := app.Command("watch", "Watch device status")

    var err error
    switch kingpin.MustParse(app.Parse(os.Args[1:])) {
    case do.FullCommand():
        err = doCommand(*host, *port, *name, *value)
    case status.FullCommand():
        err = doStatus(*host, *port)
    case watch.FullCommand():
        err = doWatch(*host, *port)
    }

    if err != nil {
        log.Fatal(err)
    }
}

func doStatus(host string, port int) error {
    device := setup(host, port)

    err := device.Start()
    if err != nil {
        return err
    }
    defer device.Stop()

    // TODO: use command lines args and use this list as default/fallback
    names := []string{
        "power",
        "volume",
        "mute",
        "speaker-a",
        "speaker-b",
        "input",
        "listen-mode",
        "display",
        "dimmer",
    }
    for _, name := range(names) {
        device.Query(name)
    }
    // TODO: wait for responses - not *forever*
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop  // wait for SIGINT

    return nil
}

func doWatch(host string, port int) error {
    device := setup(host, port)

    err := device.Start()
    if err != nil {
        return err
    }
    defer device.Stop()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop  // wait for SIGINT

    return nil
}

func doCommand(host string, port int, name, value string) error {
    device := setup(host, port)

    err := device.Start()
    if err != nil {
        return err
    }
    defer device.Stop()

    // TODO: wait until command has been send
    return device.SendCommand(name, value)
}

func setup(host string, port int) onkyoctl.Device {
    device := onkyoctl.NewDevice(host)
    device.OnMessage(func(name, value string) {
        fmt.Printf("Status: %v = %v\n", name, value)
    })
    return device
}
