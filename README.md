# denon-api

This is a simple REST API server for a Denon AVR1912 (also AVR2112CI) receiver.

It connects to the receiver via the telnet interface, and communicates using the
[Denon AVR control protocol](https://assets.denon.com/DocumentMaster/MASTER/AVR2112CI_AVR1912_PROTOCOL_V740.pdf).

The API server itself uses [Gin](https://github.com/gin-gonic/gin).

## Use

```
DENON_AVR_HOST=1.2.3.4
go run github.com/seanrees/denon-api/cmd/server -avr $DENON_AVR_HOST`
```

This server default to port 8080, and you can see a demonstration remote controller at http://localhost:8080 once started.

Run the server with `-help` to see the flags to override the listen host/port, etc.

## Commands

URL | Method | Action
--- | ------ | ------
`/input` | GET | Current Input Mode (e.g; TV)
`/input/VAL` | PUT | Set input to `VAL`
`/mode` | GET | Current surround sound mode (e.g; `MCH STEREO`)
`/mode/VAL` | PUT | Set surround mode to `VAL` (e.g; `DIRECT`, `MCH STEREO`)
`/power` | GET | Current power mode, one of `ON` or `STANDBY`
`/power/on` | PUT | Power the AVR on
`/power/standby` | PUT | Set the AVR to Standby mode
`/volume` | GET | Current volume in dB
`/volume/down` | POST | Lower the volume by 0.5 dB
`/volume/up` | POST | Raise the volume by 0.5 dB
`/volume/VAL` | PUT | Set the volume to `VAL` dB