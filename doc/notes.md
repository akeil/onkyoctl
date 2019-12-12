# Onkyo Remote Control
Control a network enabled Onkyo stereo amplifier.

See:
- https://github.com/miracle2k/onkyo-eiscp

## Basics
We connect over TCP on port *60128*

We can send and receive messages in *eISCP* format.

eISCP stands for *Integra Serial Control Protocol over Ethernet*.
An eISCP message is an ISCP message wrapped for sending
over the network.
`ISCP` messages look like this:

    PWR01   // Power on
    SPA01   // Speaker A off
    MVL3A   // Volume 29.0

The `eISCP` format adds a header and encodes the messages as bytes.

The messages are not (easily) human readable
and should be translated to/from friendly text.

We send *commands* or *queries*
we receive status messages.

All commands are asynchronous.

Command structure:

    GROUP + ARG

These can be mapped to *Properties*, *Actions* and *Events*.

### Properties
- the `QSTN` command is only used to send queries
  we need these to initialize property states
- Some of the on/off properties can be mapped to booleans.
- an additonal *toggle* command is an *Action*
  which changes a *Property*

The `DIM` commands can be mapped to a "slider"
with four levels (0=off...3=bright).

When we receive a command (`MVL52`),
we look for a property matching the group (`MVL`)
and make it apply the value.
Application of the value may fail if not all possible messages are supported.

### Actions
When we want to execute an action such as "volumeUp",
we need to send the associated command `MVLUP`.

For actions with no parameters, the command is always the same.

The response to an action may be the changed property:
```
-> MVLUP
<- MVL52
```

### Events
Some selected commands may be mapped to events:
- power up
- power down
- update available/started/complete
- connected
- connection lost


## Command Overview

| Command   | Group         | Arg      | Property  | Action     |
|-----------|---------------|----------|-----------|------------|
| PWR00     | power         | off      | power     |            |
| PWR01     | power         | on       | power     |            |
| PWRQSTN   | power         | query    | power     |            |
| AMT00     | audio-mute    | off      | mute      |            |
| AMT01     | audio-mute    | on       | mute      |            |
| AMTTG     | audio-mute    | toggle   |           | toggleMute |
| AMTQSTN   | audio-mute    | query    | mute      |            |
| SPA00     | speaker-a     | off      | speaker-a |            |
| SPA01     | speaker-a     | on       | speaker-a |            |
| SPAQSTN   | speaker-a     | query    | speaker-a |            |
| SPB00     | speaker-b     | off      | speaker-b |            |
| SPB01     | speaker-b     | on       | speaker-b |            |
| SPBQSTN   | speaker-b     | query    | speaker-b |            |
| MVLnn     | master-volume |          | volume    |            |
| MVLUP     | master-volume | up       |           | volumeUp   |
| MVLDOWN   | master-volume | down     |           | volumeDown |
| MVLUP1    | master-volume | up 1db   |           |            |
| MVLDOWN1  | master-volume | down 1db |           |            |
| MVLQSTN   | master-volume | query    | volume    |            |
| DIF00     | display-mode  | mode 0   |           |            |
| DIF01     | display-mode  | mode 1   |           |            |
| DIF02     | display-mode  | mode 2   |           |            |
| DIF03     | display-mode  | mode 3   |           |            |
| DIFTG     | display-mode  | toggle   |           |            |
| DIFQSTN   | display-mode  | query    |           |            |
| DIM00     | dimmer-level  | bright   | dimmer    |            |
| DIM01     | dimmer-level  | dim      | dimmer    |            |
| DIM02     | dimmer-level  | dark     | dimmer    |            |
| DIM03     | dimmer-level  | off      | dimmer    |            |
| DIM08     | dimmer-level  | LED off  |           |            |
| DIMQSTN   | dimmer-level  | query    | dimmer    |            |
| SLI00     | input-select  | video1   | input     |            |
| SLI01     | input-select  | cbl/sat  | input     |            |
| SLI02     | input-select  | game     | input     |            |
| SLI03     | input-select  | aux1     | input     |            |
| SLI04     | input-select  | aux2     | input     |            |
| SLI05     | input-select  | pc       | input     |            |
| SLI06     | input-select  | video7   | input     |            |
| SLI07     | input-select  | extra1   | input     |            |
| SLI08     | input-select  | extra2   | input     |            |
| SLI09     | input-select  | extra3   | input     |            |
| SLI10     | input-select  | dvd      | input     |            |
| SLI11     | input-select  | strm-box | input     |            |
| SLI12     | input-select  | tv       | input     |            |
| SLI20     | input-select  | tape1    | input     |            |
| SLI21     | input-select  | tape2    | input     |            |
| SLIxx     | input-select  | (more)   | input     |            |
| SLIQSTN   | input-select  | query    | input     |            |
| LMD00     | listen-mode   | stereo   |           |            |
| LMDSTEREO | listen-mode   | stereo   |           |            |
| LMD01     | listen-mode   | direct   |           |            |
| LMDxx     | listen-mode   | (more)   |           |            |
| LMD11     | listen-mode   | pure     |           |            |
| LMDQSTN   | listen-mode     | query                   |                |            |
| APD00     | auto-powerdown  | off                     | auto-standby   |            |
| APD01     | auto-powerdown  | on                      | auto-standby   |            |
| APDUP     | auto-powerdown  | wrap                    | auto-standby   |            |
| APDQSTN   | auto-powerdown  | query                   | auto-standby   |            |
| MOT00     | music-optimizer | off                     | optimizer      |            |
| MOT01     | music-optimizer | on                      | optimizer      |            |
| MOTUP     | music-optimizer | wrap                    | optimizer      |            |
| MOTQSTN   | music-optimizer | query                   | optimizer      |            |
| UPD00     | update          | no-new-firmware         | update         |            |
| UPD01     | update          | new-firmware            | update         |            |
| UPDNET    | update          | start-net               |                | startNetUpdate |
| UPDUSB    | update          | start-usb               |                | startUSBUpdate |
| UPDCMP    | update          | update-complete         | update         |           |
| UPDQSTN   | update          | query                   | update         |           |
| NDS---    | net-usb-status  | none                    | net-usb-status |           |
| NDSE--    | net-usb-status  | ethernet                | net-usb-status |           |
| NDSW--    | net-usb-status  | wifi                    | net-usb-status |           |
| NDS-i-    | net-usb-status  | front-iPhone            | net-usb-status |           |
| NDS-M-    | net-usb-status  | front-memory            | net-usb-status |           |
| NDS-W-    | net-usb-status  | front-wifi-adaptor      | net-usb-status |           |
| NDS-B-    | net-usb-status  | front-bluetooth-adaptor | net-usb-status |           |
| NDS-x-    | net-usb-status  | front-disabled          | net-usb-status |           |
| NDS--i    | net-usb-status  | rear-iPhone             | net-usb-status |           |
| NDS--M    | net-usb-status  | rear-memory             | net-usb-status |           |
| NDS--W    | net-usb-status  | rear-wifi-adaptor       | net-usb-status |           |
| NDS--B    | net-usb-status  | rear-bluetooth-adaptor  | net-usb-status |           |
| NDS--x    | net-usb-status  | rear-disabled           | net-usb-status |           |
| TFRB+nT+n | tone-front     | bass +2, treble +1      | tone-front     |           |
| TFRBnn    | tone-front     | +/-nn dB                | tone-bass      |           |
| TFRTnn    | tone-front     | +/-nn dB                | tone-treble    |           |
| TFRBUP    | tone-front     | up                      | tone-bass      |           |
| TFRBDOWN  | tone-front     | down                    | tone-bass      |           |
| TFRTUP    | tone-front     | up                      | tone-treble    |           |
| TFRTDOWN  | tone-front     | down                    | tone-treble    |           |
| NSBOFF    | network-standby | off                    | network-standby |           |
| NSBON     | network-standby | on                     | network-standby |           |
| NSBQSTN   | network-standby | query                  | network-standby |           |


### Volume (MVL)
Assume always(?) 8-bit integer
that is `0x00` through `0xFF` (0-255).

The allowed range is 0..200 (guess)
the "actual" value on the display has 0.5 steps and is 0..100 - so divide by 2.

It seems that we need to send the hex-string for this.

E.g. for `26`
we send `1A` (string)
effectively `0x31 0x41` (bytes).

### Tone Front (TFR)
Values look like this when the current status is broadcast:
```
TFRB+2T+1   Bass +2dB, Treble +1dB
FRB+3T00    Bass +3dB, Treble +0dB
TFRB+3T-1   Bass +2dB, Treble -1dB
```

for writing, we can (must) address bass and treble separately:
```
TFRB-5  Bass -5
TFRT+4  Treble +4
```

It would be more intuitive to treat these as *two* properties: bass and treble.

For writing, we must support two properties:

- TFR[B] / bass
- TFR[T] / treble

and for reading we must generate updates for two properties
from one ISCP message.
