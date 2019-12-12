package onkyoctl

import (
	"sync"
	"time"
)

const protocol = "tcp"

// Callback is the type for message callback functions.
type Callback func(name, value string)

// Device is an Onkyo device.
type Device struct {
	Host           string
	Port           int
	log            Logger
	commands       CommandSet
	callback       Callback
	onConnect      func()
	onDisconnect   func()
	wait           *sync.WaitGroup
	autoConnect    bool
	allowReconnect bool
	reconnectTime  time.Duration
	client         *client
}

// NewDevice sets up a new Onkyo device.
func NewDevice(cfg *Config) *Device {
	commands := cfg.Commands
	if commands == nil {
		commands = emptyCommands()
	}

	log := cfg.Log
	if log == nil {
		log = NewLogger(NoLog)
	}

	d := &Device{
		Host:           cfg.Host,
		Port:           cfg.Port,
		log:            log,
		commands:       commands,
		wait:           &sync.WaitGroup{},
		autoConnect:    cfg.AutoConnect,
		allowReconnect: cfg.AllowReconnect,
		reconnectTime:  time.Duration(cfg.ReconnectSeconds) * time.Second,
		client:         newClient(cfg.Host, cfg.Port, log),
	}

	d.client.handler = d.handleReceived
	d.client.connectionCB = d.connectionChanged
	return d
}

// OnMessage sets the handler for received messages to the given function.
// This will replace any existing handler.
func (d *Device) OnMessage(callback Callback) {
	d.callback = callback
}

// OnDisconnected is called when the device is disconnected.
func (d *Device) OnDisconnected(callback func()) {
	d.onDisconnect = callback
}

// OnConnected is called when the deivce is (re-)connected.
func (d *Device) OnConnected(callback func()) {
	d.onConnect = callback
}

// Start connects to the device and starts receiving messages.
func (d *Device) Start() {
	d.client.Start()
	d.client.Connect()
}

// Stop disconnects from the device and stop message processing.
func (d *Device) Stop() {
	d.log.Info("Stop device [%v:%v]", d.Host, d.Port)
	d.client.Stop()
}

// SendCommand sends an "friendly" command (e.g. "power off") to the device.
//
// This method calls `SendISCP()` behind the scenes.
func (d *Device) SendCommand(name string, param interface{}) error {
	command, err := d.commands.CreateCommand(name, param)
	if err != nil {
		return err
	}

	return d.SendISCP(command, 0)
}

// Query sends a QSTN command for the given friendly name.
//
// This method calls `SendISCP()` behind the scenes.
func (d *Device) Query(name string) error {
	q, err := d.commands.CreateQuery(name)
	if err != nil {
		return err
	}
	return d.SendISCP(q, 0)
}

// SendISCP sends a raw ISCP command to the device.
//
// You must `Start()` before you can send messages.
// The device may lose its connection after start. With AutoConnect
// set to true, attempts to connect and returns an error only if that fails.
// Without autoconnect, an error is returned if the device is not connected.
//
// The message is send asynchronously. Use a non-zero timeout to wait until
// the message is sent.
// Note that the message may still be sent even if `ErrTimeout` is returned.
func (d *Device) SendISCP(cmd ISCPCommand, timeout time.Duration) error {
	if d.autoConnect {
		// if already connected, this does nothing
		d.client.Connect()
	}
	d.client.WaitConnect(timeout)

	return d.client.Send(cmd, timeout)
}

func (d *Device) connectionChanged(s ConnectionState) {
	d.log.Debug("Connection state changed to %q", s)
	if s == Connected && d.onConnect != nil {
		d.onConnect()
	}

	if s == Disconnected && d.onDisconnect != nil {
		d.onDisconnect()
		if d.allowReconnect {
			//TODO: not when we Stop()'ed
			d.log.Debug("Schedule reconnect")
			go func() {
				time.Sleep(d.reconnectTime)
				d.client.Connect()
			}()
		}
	}
}

func (d *Device) handleReceived(cmd ISCPCommand) {
	name, value, err := d.commands.ReadCommand(cmd)
	if err != nil {
		d.log.Warning("Error reading %q: %v", cmd, err)
		return
	}
	d.log.Debug("Received '%v %v'", name, value)
	if d.callback != nil {
		d.callback(name, value)
	}
}

// BasicCommands creates a command set with some commonly used commands.
func BasicCommands() CommandSet {
	commands := []Command{
		Command{
			Name:      "power",
			Group:     "PWR",
			ParamType: "onOff",
		},
		Command{
			Name:      "volume",
			Group:     "MVL",
			ParamType: "intRangeEnum",
			Lower:     0,
			Upper:     100,
			Scale:     2,
			Lookup: map[string]string{
				"UP":   "up",
				"DOWN": "down",
			},
		},
		Command{
			Name:      "mute",
			Group:     "AMT",
			ParamType: "onOffToggle",
		},
		Command{
			Name:      "speaker-a",
			Group:     "SPA",
			ParamType: "onOff",
		},
		Command{
			Name:      "speaker-b",
			Group:     "SPB",
			ParamType: "onOff",
		},
		Command{
			Name:      "dimmer",
			Group:     "DIM",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "bright",
				"01": "dim",
				"02": "dark",
				"03": "off",
				"08": "led-off",
			},
		},
		Command{
			Name:      "display",
			Group:     "DIF",
			ParamType: "enumToggle",
			Lookup: map[string]string{
				"00": "default",
				"01": "listening",
				"02": "source",
				"03": "mode-4",
			},
		},
		Command{
			Name:      "input",
			Group:     "SLI",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "video-1",
				"01": "cbl-sat",
				"02": "game",
				"03": "aux1",
				"20": "tv-tape",
			},
		},
		Command{
			Name:      "listen-mode",
			Group:     "LMD",
			ParamType: "enum",
			Lookup: map[string]string{
				"00":     "stereo",
				"STEREO": "stereo",
				"01":     "direct",
				"11":     "pure",
			},
		},
		Command{
			Name:      "update",
			Group:     "UPD",
			ParamType: "enum",
			Lookup: map[string]string{
				"00": "no-new-firmware",
				"01": "new-firmware",
			},
		},
	}
	return NewBasicCommandSet(commands)
}

func emptyCommands() CommandSet {
	return NewBasicCommandSet(make([]Command, 0))
}
