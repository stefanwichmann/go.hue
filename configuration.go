package hue

import "encoding/json"

// Configuration contains all basic information about the hue bridge itself.
type Configuration struct {
	Name             string                 `json:"name"`
	ZigbeeChannel    int                    `json:"zigbeechannel"`
	SoftwareUpdate   map[string]interface{} `json:"swupdate"`
	Whitelist        map[string]interface{} `json:"whitelist"`
	APIVersion       string                 `json:"apiversion"`
	SoftwareVersion  string                 `json:"swversion"`
	Proxyaddress     string                 `json:"proxyaddress"`
	Proxyport        int                    `json:"proxyport"`
	Linkbutton       bool                   `json:"linkbutton"`
	IPAddress        string                 `json:"ipadress"`
	Mac              string                 `json:"mac"`
	Netmask          string                 `json:"netmask"`
	Gateway          string                 `json:"gateway"`
	DHCP             bool                   `json:"dhcp"`
	Portalservices   bool                   `json:"bool"`
	UTC              string                 `json:"UTC"`
	Localtime        string                 `json:"localtime"`
	Timezone         string                 `json:"timezone"`
	ModelId          string                 `json:"modelid"`
	BridgeId         string                 `json:"bridgeid"`
	FactoryNew       bool                   `json:"factorynew"`
	ReplacesBridgeId string                 `json:"replacesbridgeid"`
	DatastoreVersion string                 `json:"datastoreversion"`
}

// Configuration return all basic information about the hue bridge itself.
func (bridge *Bridge) Configuration() (*Configuration, error) {
	response, err := bridge.get("/config")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result Configuration
	err = json.NewDecoder(response.Body).Decode(&result)
	return &result, err
}
