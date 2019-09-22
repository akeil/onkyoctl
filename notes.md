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

The messages are not human readable
and should be translated to/from friendly text.

We send *commands* or *queries*
we receive status messages.

All commands are asynchronous.

## Webthing
Command structure:

    GROUP + ARG

These can be mapped to *Properties*, *Actions* and *Events*.

### Properties
- the QSTN command is only used to send queries
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

| Command  | Group         | Arg      | Property  | Action     |
|----------|---------------|----------|-----------|------------|
| PWR00    | power         | off      | power     |            |
| PWR01    | power         | on       | power     |            |
| PWRQSTN  | power         | query    | power     |            |
| AMT00    | audio-mute    | off      | mute      |            |
| AMT01    | audio-mute    | on       | mute      |            |
| AMTTG    | audio-mute    | toggle   |           | toggleMute |
| AMTQSTN  | audio-mute    | query    | mute      |            |
| SPA00    | speaker-a     | off      | speaker-a |            |
| SPA01    | speaker-a     | on       | speaker-a |            |
| SPAQSTN  | speaker-a     | query    | speaker-a |            |
| SPB00    | speaker-b     | off      | speaker-b |            |
| SPB01    | speaker-b     | on       | speaker-b |            |
| SPBQSTN  | speaker-b     | query    | speaker-b |            |
| MVLnn    | master-volume |          | volume    |            |
| MVLUP    | master-volume | up       |           | volumeUp   |
| MVLDOWN  | master-volume | down     |           | volumeDown |
| MVLUP1   | master-volume | up 1db   |           |            |
| MVLDOWN1 | master-volume | down 1db |           |            |
| MVLQSTN  | master-volume | query    | volume    |            |
| DIF00    | display-mode  | mode 0   |           |            |
| DIF01    | display-mode  | mode 1   |           |            |
| DIF02    | display-mode  | mode 2   |           |            |
| DIF03    | display-mode  | mode 3   |           |            |
| DIFTG    | display-mode  | toggle   |           |            |
| DIFQSTN  | display-mode  | query    |           |            |
| DIM00    | dimmer-level  | bright   | dimmer    |            |
| DIM01    | dimmer-level  | dim      | dimmer    |            |
| DIM02    | dimmer-level  | dark     | dimmer    |            |
| DIM03    | dimmer-level  | off      | dimmer    |            |
| DIM08    | dimmer-level  | LED off  |           |            |
| DIMQSTN  | dimmer-level  | query    | dimmer    |            |
| SLI00    | input-select  | video1   | input     |            |
| SLI01    | input-select  | cbl/sat  | input     |            |
| SLI02    | input-select  | game     | input     |            |
| SLI03    | input-select  | aux1     | input     |            |
| SLI04    | input-select  | aux2     | input     |            |
| SLI05    | input-select  | pc       | input     |            |
| SLI06    | input-select  | video7   | input     |            |
| SLI07    | input-select  | extra1   | input     |            |
| SLI08    | input-select  | extra2   | input     |            |
| SLI09    | input-select  | extra3   | input     |            |
| SLI10    | input-select  | dvd      | input     |            |
| SLI11    | input-select  | strm-box | input     |            |
| SLI12    | input-select  | tv       | input     |            |
| SLI20    | input-select  | tape1    | input     |            |
| SLI21    | input-select  | tape2    | input     |            |
| SLIxx    | input-select  | (more)   | input     |            |
| SLIQSTN  | input-select  | query    | input     |            |
| LMD00    | listen-mode   | stereo   |           |            |
| LMDSTEREO | listen-mode   | stereo   |           |            |
| LMD01    | listen-mode   | direct   |           |            |
| LMDxx    | listen-mode   | (more)   |           |            |
| LMD11    | listen-mode   | pure     |           |            |
| LMDQSTN  | listen-mode   | query    |           |            |
| APD00    | auto-powerdown | off      | auto-standby |            |
| APD01    | auto-powerdown | on       | auto-standby |            |
| APDUP    | auto-powerdown | wrap     | auto-standby |            |
| APDQSTN  | auto-powerdown | query    | auto-standby |            |
