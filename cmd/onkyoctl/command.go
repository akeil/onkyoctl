package main

import (
    "errors"
    "fmt"
    "log"
    "os"
	"os/signal"
    "path"
    "sync"
    "time"

    "gopkg.in/alecthomas/kingpin.v2"

    onkyo "akeil.net/akeil/onkyoctl"
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

    var (
        host = app.Flag("host", "Hostname or IP address").String()
        port = app.Flag("port", "Port number").Default("60128").Short('p').Int()
        cfgPath = app.Flag("config", "Path to configuration file").Short('c').String()
        verbose = app.Flag("verbose", "Verbose output").Short('v').Bool()
    )

    do := app.Command("do", "Execute a command").Default()
    var commands = do.Arg("commands", "Commands to send, pairs of <name> <value> - e.g. 'power on volume up'").Required().Strings()

    status := app.Command("status", "Show device status")
    var names = status.Arg("names", "Status items to query, e.g. 'power volume'. Leave empty to query defaults").Strings()

    watch := app.Command("watch", "Watch device status")
    version := app.Command("version", "Print version")

    subCommand := kingpin.MustParse(app.Parse(os.Args[1:]))

    if subCommand == version.FullCommand() {
        fmt.Println(onkyo.Version)
        return
    }

    logLevel := onkyo.Error
    if *verbose {
        logLevel = onkyo.Debug
    }

    device := setup(logLevel, *cfgPath, *host, *port)
    err := device.Start()
    defer device.Stop()
    if err != nil {
        log.Fatal(err)
    }

    switch subCommand {
    case do.FullCommand():
        err = doCommands(device, *commands)

    case status.FullCommand():
        err = doStatus(device, *names)

    case watch.FullCommand():
        err = doWatch(device)
    }

    if err != nil {
        log.Fatal(err)
    }
}

func doStatus(device *onkyo.Device, names []string) error {
    fmt.Printf("Status [%v]:\n", device.Host)

    if len(names) == 0 {
        names = []string{
            "power",
            "volume",
            "mute",
            "speaker-a",
            "speaker-b",
            "input",
        }
    }

    // expect a reply for every query we send
    var wait sync.WaitGroup

    device.OnMessage(func(name, value string) {
        fmt.Printf("%v: %v\n", name, value)
        // note: not *quite* correct - we accept duplicate responses
        if contains(names, name) {
            wait.Done()
        }
    })

    var err error
    for _, name := range(names) {
        wait.Add(1)
        err = device.Query(name)
        if err != nil {
            return err
        }
    }

    // wait until all responses are received or timeout
    done := make(chan int)
	go func() {
		defer close(done)
		wait.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("Timeout waiting for response")
	}
}

func doWatch(device *onkyo.Device) error {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop  // wait for SIGINT

    return nil
}

func doCommands(device *onkyo.Device, pairs []string) error {
    if len(pairs) % 2 != 0 {
        return errors.New("number of arguments must be even")
    }

    for i := 0; i < len(pairs); i+=2 {
        name := pairs[i]
        value := pairs[i+1]
        err := device.SendCommand(name, value)
        if err != nil {
            return err
        }
    }

    device.WaitSend(1 * time.Second)
    return nil
}

func setup(logLevel onkyo.LogLevel, cfgPath, host string, port int) *onkyo.Device {
    var err error
    cfg := onkyo.DefaultConfig()

    // explicit param or default
    if cfgPath == "" {
        cfgBase, err := os.UserConfigDir()
        if err == nil {
            cfgPath = path.Join(cfgBase, "onkyoctl.ini")
        }
    }
    if cfgPath != "" {
        cfg, err = onkyo.ReadConfig(cfgPath)
        if err != nil {
            log.Printf("Error reading config from %q: %v", cfgPath, err)
            cfg = onkyo.DefaultConfig()
        }
    }

    cfg.Log = onkyo.NewLogger(logLevel)

    // override some config settings from command line
    if host != "" {
        cfg.Host = host
    }
    if port != 0 {
        cfg.Port = port
    }

    cfg.Commands = onkyo.BasicCommands()

    device := onkyo.NewDevice(cfg)
    device.OnMessage(func(name, value string) {
        fmt.Printf("%v = %v\n", name, value)
    })
    return device
}

func contains(haystack []string, needle string) bool {
    for _, item := range(haystack) {
        if item == needle {
            return true
        }
    }
    return false
}
