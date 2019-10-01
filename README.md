# OnkyoCTL
onkyoctl is a library and command line app to control Onkyo devices over
the network.

## Command Line Usage
The command line tool supports three subcommands.

The default command is `do` and you do not need to spell it out.
It takes a space-separated list of <command> <param> pairs and sends these
to the device. It does not wait for a reply.
```
$ onkyoctl power on volume up speaker-a on
````

Use the `status` command to query properties of the device.
When called without arguments, a default set of properties is queried.
```
$ onkyoctl status power volume
power: on
volume: 23.5
```
or without arguments:
```
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
```
$ onkyoctl watch
volume: 23.5
volume: 26
volume: 29.5
volume: 32
...
```

## Configuration
For command line usage, the configuration file is expected at:
`~/.config/onkyoctl.ini`.

It looks like this:
```ini
# IP address of the onkyo device
# you will probably want to set this
Host = 192.168.1.123

# Port number (default: 60128)
Port = 60123

# Connection timeout in seconds
ConnectTimout = 10

# Reconnect after connection loss?
AllowReconnect = false
ReconnectSeconds = 5

# Reconnect when a message needs to be sent?
AutoConnect = false
```

When used as a library, the `Config` struct is used to configure a `Device`.
Use `ReadConfig(path)` to populate it from an *.ini* file or set individual
options directly.

## Connection Issues
At least the one unit this was tested on would only hold *one* TCP connection
at a time. This means, when you connect to the device while another client is
already connected, the connection to the *other* client will be closed.
Likewise, when another client connects, we will lose our connection.

This may become a problem for long-running processes which require a constant
connection to monitor the device.
We have two options to deal with this:

`AutoConnect`: when set to *true*, the `Device` will attempt to reconnect
as soon as a message needs to be sent.

`AllowReconnect`: after we lose connection, waits for `ReconnectSeconds` and
then attempts to reconnect.
Use with care - if other clients behave the same way.

The `OnConnected` and `OnDisconnected` are called when the connection status
changes.

## Credits
- https://github.com/miracle2k/onkyo-eiscp
- https://sites.google.com/a/webarts.ca/toms-blog/Blog/new-blog-items/javaeiscp-integraserialcontrolprotocol
