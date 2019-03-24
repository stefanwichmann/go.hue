package hue

import (
	"errors"
	"fmt"
	"github.com/stefanwichmann/lanscan"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const discoveryTimeout = 3 * time.Second

// DiscoverBridges is a two-step approach trying to find your hue bridges.
// First it will try to discover bridges in your network using UPnP and it
// will utilize the hue api (https://www.meethue.com/api/nupnp) to
// fetch a list of known bridges at the current location in parallel.
// Should this fail it will automatically scan all hosts in your local
// network and identify any bridges you have running.
// If the parameter discoverAllBridges is true the discovery will wait for all
// bridges to respond. When set to false, this method will return as soon as it
// found the first bridge in your network.
func DiscoverBridges(discoverAllBridges bool) ([]Bridge, error) {
	hostChannel := make(chan string, 10)
	bridgeChannel := make(chan string, 10)

	// Start UPnP and N-UPnP discovery in parallel
	go upnpDiscover(hostChannel)
	go nupnpDiscover(hostChannel)
	go validateBridges(hostChannel, bridgeChannel)

	var bridges = []Bridge{}
	scanStarted := false
loop:
	for {
		select {
		case bridge, more := <-bridgeChannel:
			if !more && len(bridges) > 0 {
				return bridges, nil
			}
			if !more {
				break loop
			}
			bridges = append(bridges, *NewBridge(bridge, ""))
			if !discoverAllBridges {
				return bridges, nil
			}
		case <-time.After(discoveryTimeout):
			if len(bridges) > 0 {
				return bridges, nil
			}
			if !scanStarted {
				// UPnP and N-UPnP didn't discover any bridges.
				// Start a LAN scan and feed results to hostChannel.
				scanLocalNetwork(hostChannel)
				scanStarted = true
				continue // Loop again with same timeout
			}
			break loop
		}
	}

	// Nothing found
	return bridges, errors.New("Bridge discovery failed")
}

func scanLocalNetwork(hostChannel chan<- string) {
	hosts, err := lanscan.ScanLinkLocal("tcp4", 80, 20, discoveryTimeout-1*time.Second)
	if err == nil {
		for _, host := range hosts {
			hostChannel <- host
		}
	}
	close(hostChannel)
}

func validateBridges(candidates <-chan string, bridges chan<- string) {
	for candidate := range candidates {
		resp, err := http.Get(fmt.Sprintf("http://%s/description.xml", candidate))
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// make sure it's a hue bridge
		str := string(body)
		if !strings.Contains(str, "<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>") {
			continue
		}
		if !strings.Contains(str, "<manufacturer>Royal Philips Electronics</manufacturer>") {
			continue
		}
		if !strings.Contains(str, "<modelURL>http://www.meethue.com</modelURL>") {
			continue
		}

		// Candidate seems to be a valid hue bridge
		bridges <- candidate
	}
	close(bridges)
}
