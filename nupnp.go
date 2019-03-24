package hue

import "encoding/json"
import "net/http"

const nupnpEndpoint = "https://www.meethue.com/api/nupnp"

type nupnpBridge struct {
	Serial string `json:"id"`
	IPAddr string `json:"internalipaddress"`
}

func nupnpDiscover(respondingHosts chan<- string) error {
	response, err := http.Get(nupnpEndpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var bridges []nupnpBridge
	err = json.NewDecoder(response.Body).Decode(&bridges)
	if err != nil {
		return err
	}

	for _, bridge := range bridges {
		respondingHosts <- bridge.IPAddr
	}
	return nil
}
