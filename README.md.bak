# OnkyoCTL
`onkyoctl` is a library and command line tool to control
[Onkyo](https://www.onkyo.com/) devices over the network.

You can switch the receiver on and off (from and to standby mode),
change the volume, input source and switch between speaker sets A/B.

**Note:**
This code has been developed and tested with a single receiver model
([TX-8250](https://www.intl.onkyo.com/products/hi-fi_components/receivers/tx-8250/index.html)).

## Library Usage
The `Device` type is used to control a receiver.
To set it up, a `Config` object is required,
which at least needs the IP address or hostname for the receiver:
```go
c := onkyoctl.NewDefaultConfig()
c.Host = "192.168.1.2"
d := onkyoctl.NewDevice(c)
d.Start()
defer d.Stop()
```

### Single Commands
If you only want to send a single command, use the *AutoConnect* config setting
to connect automatically.

```go
c := onkyoctl.NewDefaultConfig()
c.Host = "192.168.1.2"
c.AutoConnect = true            // connect as soon as required
d := onkyoctl.NewDevice(c)

d.SendCommand("volume", 25)     // will automatically connect
d.Stop()
```

We still need to `Stop()` the device if we want to disconnect
after the command is sent.

### Receive Status Changes
The commands do not return an immediate response.
Instead, we need to observe the receiver for status changes
to determine whether a command was successful.

Register the `OnMessage` callback to receive notifications for all messages
sent by the receiver:

```go
d.OnMessage(func(name, value string) {
    // name is e.g. "volume"
    // value is e.g. "45"
})
```

### Continuous Connection
The receiver supports a long-living connection over which we can send several
commands and receive messages for status updates.
Unfortunately, the device supports *exactly one* client connection. If another
client connects, the receiver will terminate our connection and serve the new
client instead.

The `onkyoctl` library deals with this by allowing an automatic reconnect
after a short hold-off time.
This is also useful to (re-)connect to the receiver as soon as it becomes
available on the network (for example after it was switched off).

```go
c := onkyoctl.NewDefaultConfig()
// ...
c.AllowReconnect = true
c.ReconnectSeconds = 10

d := onkyoctl.NewDevice(c)
d.Start()
defer d.Stop()
// ...
```

If `AllowReconnect` is *true*, the device will reconnect when the connection is
lost. Commands that were issued while the device is disconnected are **queued**
and will be sent as soon as we are reconnected.

The `OnConnected` and `OnDisconnected` callbacks can be used to react to
changes in the connection status:

```go
///...
d.OnConnected(func() {
    // do something when the device (re-)connects
})
d.OnDisconnected(func(){
    // do something when we lose connection
})
```

**Warning:** The reconnect behavior means that as soon as we reconnect, we will
cause the receiver to disconnect any other client that is currently connected.
To avoid completely blocking other clients, use `ReconnectSeconds` to give
other clients sufficient time to complete their tasks.
If you have other clients that need a constant connection to the receiver,
this will not work.

## Command Line Usage
The command line tool supports three subcommands.

The default command is `do` and you do not need to spell it out.
It takes a space-separated list of <command> <param> pairs and sends these
to the device. It does not wait for a reply.

```shell
$ onkyoctl power on volume up speaker-a on
````

Use the `status` command to query properties of the device.
When called without arguments, a default set of properties is queried.

```shell
$ onkyoctl status power volume
power: on
volume: 23.5
```
or without arguments:

```shell
$ onkyoctl status
power: on
volume: 23.5
mute: off
speaker-a: on
speaker-b: off
input: game
```

The `watch` command connects to the device and prints out any status messages
it receives. Use `ctrl + c` to quit.

```shell
$ onkyoctl watch
volume: 23.5
volume: 26
volume: 29.5
volume: 32
```

## Configuration
For command line usage, the configuration file is expected at:
`~/.config/onkyoctl.ini`.

It looks like this:
```ini
# IP address of the onkyo device
# you will probably want to set this
Host = 192.168.1.2

# Port number (default: 60128)
Port = 60123

# Reconnect after connection loss?
AllowReconnect = false
ReconnectSeconds = 5

# Reconnect when a message needs to be sent?
AutoConnect = false
```

When used as a library, the `Config` struct is used to configure a `Device`.
Use `ReadConfig(path)` to populate it from an *.ini* file or set individual
options directly.

## Similar Projects

- https://github.com/miracle2k/onkyo-eiscp
- https://sites.google.com/a/webarts.ca/toms-blog/Blog/new-blog-items/javaeiscp-integraserialcontrolprotocol
