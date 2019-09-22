package main

import (
    "log"
    "fmt"
    "os"
	"os/signal"
    "path"
    "time"

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

    app := kingpin.New("onkyo", "Control Onkyo receiver.")
    app.HelpFlag.Short('h')

    var host = app.Flag("host", "Hostname or IP address").String()
    var port = app.Flag("port", "Port number").Default("60128").Short('p').Int()
    var cfgPath = app.Flag("config", "Path to configuration file").Short('c').String()
    var verbose = app.Flag("verbose", "Verbose output").Short('v').Bool()

    do := app.Command("do", "Execute a command").Default()
    var name = do.Arg("name", "The property to change").Required().String()
    var value = do.Arg("value", "The value to set").String()

    status := app.Command("status", "Show device status")
    watch := app.Command("watch", "Watch device status")

    onkyoctl.SetLogLevel(onkyoctl.Error)

    var err error
    switch kingpin.MustParse(app.Parse(os.Args[1:])) {
    case do.FullCommand():
        if *verbose {
            onkyoctl.SetLogLevel(onkyoctl.Debug)
        }
        err = doCommand(*cfgPath, *host, *port, *name, *value)

    case status.FullCommand():
        if *verbose {
            onkyoctl.SetLogLevel(onkyoctl.Debug)
        }
        err = doStatus(*cfgPath, *host, *port)

    case watch.FullCommand():
        if *verbose {
            onkyoctl.SetLogLevel(onkyoctl.Debug)
        }
        err = doWatch(*cfgPath, *host, *port)
    }

    if err != nil {
        log.Fatal(err)
    }
}

func doStatus(cfgPath, host string, port int) error {
    device := setup(cfgPath, host, port)

    err := device.Start()
    if err != nil {
        return err
    }
    defer device.Stop()

    fmt.Printf("Device Status:\n")

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

func doWatch(cfgPath, host string, port int) error {
    device := setup(cfgPath, host, port)

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

func doCommand(cfgPath, host string, port int, name, value string) error {
    device := setup(cfgPath, host, port)

    err := device.Start()
    if err != nil {
        return err
    }
    defer device.Stop()

    err = device.SendCommand(name, value)
    if err != nil {
        return err
    }

    device.WaitSend(1 * time.Second)
    return nil
}

func setup(cfgPath, host string, port int) onkyoctl.Device {
    var err error
    cfg := onkyoctl.DefaultConfig()

    // explicit param or default
    if cfgPath == "" {
        cfgBase, err := os.UserConfigDir()
        if err == nil {
            cfgPath = path.Join(cfgBase, "onkyoctl.ini")
        }
    }
    if cfgPath != "" {
        cfg, err = onkyoctl.ReadConfig(cfgPath)
        if err != nil {
            log.Printf("Error reading config from %q: %v", cfgPath, err)
            cfg = onkyoctl.DefaultConfig()
        }
    }

    // override some config settings from command line
    if host != "" {
        cfg.Host = host
    }
    if port != 0 {
        cfg.Port = port
    }

    device := onkyoctl.NewDevice(cfg)
    device.OnMessage(func(name, value string) {
        fmt.Printf("%v = %v\n", name, value)
    })
    return device
}
