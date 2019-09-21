# Onkyo Remote Control
Control a network enabled Onkyo stereo amplifier.

See:
- https://github.com/miracle2k/onkyo-eiscp

## Basics
We connect over TCP on port **

We can send and receive messages in *eISCP* format.

An eISCP message is an ISCP message wrapped for sending
over the network.

The messages are not human readable
and should be translated to/from friendly text.

We send *commands* or *queries*
we receive status messages.

All commands are asynchronous.
