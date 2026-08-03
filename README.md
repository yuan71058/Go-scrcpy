# Go-scrcpy

Go library for interfacing with Android devices via [scrcpy](https://github.com/Genymobile/scrcpy) protocol.

## Features

- Full scrcpy v4.1 protocol support (22 control message types + ws-scrcpy extensions)
- Multi-device concurrent control
- Video streaming (H.264, H.265, AV1, VP8, VP9)
- Audio streaming (Opus, AAC, FLAC, PCM)
- Touch, keyboard, and scroll input injection
- Clipboard bidirectional sync
- File push and APK install
- Screen recording (MP4/MKV) and screenshot (PNG/JPEG)
- Camera control (torch, zoom, switching)
- UHID virtual HID device support
- Automatic device discovery and tracking

## Installation

```bash
go get github.com/yuan71058/go-scrcpy
```

## Quick Start

### Single Device

```go
package main

import (
    "context"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
    "github.com/yuan71058/go-scrcpy/pkg/server"
)

func main() {
    adbClient := &adb.Client{ExecPath: "adb"}

    c, err := scrcpy.New("DEVICE_SERIAL", server.Options{
        VideoBitRate: 4000000,
        Audio:        true,
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := c.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    for frame := range c.VideoStream() {
        _ = frame // Process video frame
    }
}
```

### Multi-Device

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yuan71058/go-scrcpy/pkg/adb"
    "github.com/yuan71058/go-scrcpy/pkg/scrcpy"
    "github.com/yuan71058/go-scrcpy/pkg/server"
)

func main() {
    adbClient := &adb.Client{ExecPath: "adb"}
    mc := scrcpy.NewMulti(adbClient)
    defer mc.Close()

    watcher := scrcpy.NewDeviceWatcher(adbClient)
    watcher.OnAdd(func(serial string) {
        fmt.Printf("Device connected: %s\n", serial)
        c, err := mc.Add(serial, server.Options{
            VideoBitRate: 4000000,
            Audio:        true,
        })
        if err != nil {
            log.Printf("Failed to connect %s: %v", serial, err)
            return
        }

        go func() {
            for frame := range c.VideoStream() {
                _ = frame
            }
        }()
    })

    watcher.OnRemove(func(serial string) {
        mc.Remove(serial)
    })

    watcher.Start(context.Background())
}
```

## Documentation

- [Development Plan](DEVELOPMENT.md) - Complete project plan and architecture
- [Protocol Specification](PROTOCOL.md) - Binary protocol details

## Requirements

- Go >= 1.22
- ADB installed and in PATH
- Android device with USB debugging enabled
- scrcpy-server.jar (bundled or downloaded separately)

## License

MIT
