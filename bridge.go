package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Bridge is a representation of the Philips Hue bridge device.
type Bridge struct {
	IpAddr   string
	Username string
	debug    bool
}

// CreateUser registers a new user on the bridge. The user will have
// to authenticate this request by pressing the blue link button
// on the physical bridge.
func (bridge *Bridge) CreateUser(deviceType string) error {
	// construct our json params
	params := map[string]string{"devicetype": deviceType}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// create a new user
	uri := fmt.Sprintf("http://%s/api", bridge.IpAddr)
	response, err := client.Post(uri, "text/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// extract username from the results
	var results []map[string]map[string]string
	json.NewDecoder(response.Body).Decode(&results)
	value := results[0]
	username := value["success"]["username"]

	// and create the new bridge object
	bridge.Username = username
	return nil
}

// NewBridge instantiates a bridge object. Use this method when you already
// know the ip address and username to use.
func NewBridge(ipAddr, username string) *Bridge {
	return &Bridge{IpAddr: ipAddr, Username: username}
}

// Debug enables the output of debug messages for every bridge request.
func (bridge *Bridge) Debug() *Bridge {
	bridge.debug = true
	return bridge
}

func (bridge *Bridge) toUri(path string) string {
	return fmt.Sprintf("http://%s/api/%s%s", bridge.IpAddr, bridge.Username, path)
}

func (bridge *Bridge) get(path string) (*http.Response, error) {
	uri := bridge.toUri(path)
	if bridge.debug {
		log.Printf("GET %s\n", uri)
	}
	return client.Get(uri)
}

func (bridge *Bridge) post(path string, body io.Reader) (*http.Response, error) {
	uri := bridge.toUri(path)
	if bridge.debug {
		log.Printf("POST %s\n", uri)
	}
	return client.Post(uri, "application/json", body)
}

func (bridge *Bridge) put(path string, body io.Reader) (*http.Response, error) {
	uri := bridge.toUri(path)
	if bridge.debug {
		log.Printf("PUT %s\n", uri)
	}
	request, err := http.NewRequest("PUT", uri, body)
	if err != nil {
		return nil, err
	}

	return client.Do(request)
}

func (bridge *Bridge) delete(path string) (*http.Response, error) {
	uri := bridge.toUri(path)
	if bridge.debug {
		log.Printf("DELETE %s\n", uri)
	}
	request, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(request)
}

// GetNewLights retrieves the list lights we've seen since
// the last scan. Returns the new lights, lastseen and any error
// that may have occurred as per:
// http://developers.meethue.com/1_lightsapi.html#12_get_new_lights
func (bridge *Bridge) GetNewLights() ([]*Light, string, error) {
	response, err := bridge.get("/lights/new")
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	results := make(map[string]interface{})
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, "", err
	}

	lastScan := results["lastscan"].(string)

	var lights []*Light
	for id, params := range results {
		if id != "lastscan" {
			value := params.(map[string]interface{})["name"]
			light := &Light{Id: id, Name: value.(string)}
			lights = append(lights, light)
		}
	}

	return lights, lastScan, nil
}

// FindLightById allows you to easily look up light if you know it's Id
func (bridge *Bridge) FindLightById(id string) (*Light, error) {
	lights, err := bridge.GetAllLights()
	if err != nil {
		return nil, err
	}

	for _, light := range lights {
		if light.Id == id {
			return light, nil
		}
	}

	return nil, errors.New("Unable to find light with id " + id)
}

// FindLightByName is a convenience method which
// returns the light with the given name.
func (bridge *Bridge) FindLightByName(name string) (*Light, error) {
	lights, err := bridge.GetAllLights()
	if err != nil {
		return nil, err
	}

	for _, light := range lights {
		if light.Name == name {
			return light, nil
		}
	}

	return nil, errors.New("Unable to find light with name " + name)
}

// Search starts a lookup for new devices on your bridge as per
// http://developers.meethue.com/1_lightsapi.html#13_search_for_new_lights
func (bridge *Bridge) Search() ([]Result, error) {
	response, err := bridge.post("/lights", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	return results, err
}

// GetAllLights retrieves all devices the bridge is aware of
func (bridge *Bridge) GetAllLights() ([]*Light, error) {
	// fetch all the lights
	response, err := bridge.get("/lights")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// deconstruct the json results
	var results map[string]Light
	err = json.NewDecoder(response.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	// and convert them into lights
	var lights []*Light
	for id, params := range results {
		light := Light{Id: id, Name: params.Name, bridge: bridge}
		lights = append(lights, &light)
	}

	return lights, nil
}
