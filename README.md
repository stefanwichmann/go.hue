# **go.hue** - An easy to use API to manage your Philips Hue
[![GoDoc](http://godoc.org/github.com/stefanwichmann/go.hue?status.png)](http://godoc.org/github.com/stefanwichmann/go.hue)
[![Build Status](https://travis-ci.org/stefanwichmann/go.hue.svg?branch=master)](https://travis-ci.org/stefanwichmann/go.hue)
[![Go Report Card](https://goreportcard.com/badge/github.com/stefanwichmann/go.hue)](https://goreportcard.com/report/github.com/stefanwichmann/go.hue)

For documentation, check out the link to godoc above.

# Features added to this fork
- Added configuration API
- Added scenes API
- Added additional bridge discovery method (LAN scanning)
- Added HTTPS support (needs bridge API version 1.24)
- Adapt API changes
- Fixed documentation issues
- Fixed ```go vet``` and ```go lint``` issues

# Examples
### Register a new device
To start using the hue API, you first need to register your device.
```go
package main
import "fmt"
import "github.com/stefanwichmann/go.hue"

func main() {
	bridges, _ := hue.DiscoverBridges(false)
	bridge := bridges[0] //Use the first bridge found

	//Remember to push the button on your hue first
	err := bridge.CreateUser("my nifty app")
	if err != nil {
		fmt.Printf("Device registration failed: %v\n", err)
	}
	fmt.Printf("Registered new device => %+v\n", bridge)
}
```

### Turn on all the lights
```go
package main
import "github.com/stefanwichmann/go.hue"

func main() {
	bridge := hue.NewBridge("your-ip-address", "your-device-name")
	lights, _ := bridge.GetAllLights()

	for _, light := range lights {
		light.On()
	}
}
```

### ***Disco Time!*** Switch all lights into colorloop
```go
package main
import "github.com/stefanwichmann/go.hue"

func main() {
	bridge := hue.NewBridge("your-ip-address", "your-device-name")
	lights, _ := bridge.GetAllLights()

	for _, light := range lights {
		light.ColorLoop()
	}
}
```

### Access lights by name
```go
package main
import "github.com/stefanwichmann/go.hue"

func main() {
	bridge := hue.NewBridge("your-ip-address", "your-device-name")
	light, _ := bridge.FindLightByName("Bathroom Light")
	light.On()
}
```
