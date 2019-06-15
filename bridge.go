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
	"sync"
	"time"
)

// Bridge is a representation of the Philips Hue bridge device.
type Bridge struct {
	IpAddr               string
	Username             string
	debug                bool
	useHTTPS             bool
	delayBetweenRequests time.Duration
	lastRequestTimestamp time.Time
	lock                 *sync.Mutex
	client               *http.Client
}

// CreateUser registers a new user on the bridge. The user will have
// to authenticate this request by pressing the blue link button
// on the physical bridge.
func (bridge *Bridge) CreateUser(deviceType string) error {
	params := map[string]string{"devicetype": deviceType}
	var results []map[string]map[string]string

	err := bridge.do("POST", bridge.baseURL(), &params, &results)
	if err != nil {
		return err
	}

	bridge.lock.Lock()
	defer bridge.lock.Unlock()

	value := results[0]
	bridge.Username = value["success"]["username"]
	return nil
}

// NewBridge instantiates a bridge object. Use this method when you already
// know the ip address and username to use.
func NewBridge(ipAddr, username string) *Bridge {
	return &Bridge{IpAddr: ipAddr, Username: username, debug: false, useHTTPS: false, delayBetweenRequests: 0, client: newTimeoutClient(), lock: &sync.Mutex{}}
}

// Debug enables the output of debug messages for every bridge request.
func (bridge *Bridge) Debug() *Bridge {
	bridge.lock.Lock()
	defer bridge.lock.Unlock()

	bridge.debug = true
	return bridge
}

// EnableHTTPS controls the use of an encrypted communication (requires bridge software version 1.24 or later)
func (bridge *Bridge) EnableHTTPS(enable bool) {
	bridge.lock.Lock()
	defer bridge.lock.Unlock()

	bridge.useHTTPS = enable
}

// EnableRateLimiting will only allow requests in the rate of the given paramter duration. If requests are issued faster, the function will wait for the specified time and execute the request afterwards.
func (bridge *Bridge) EnableRateLimiting(delayBetweenRequests time.Duration) {
	bridge.lock.Lock()
	defer bridge.lock.Unlock()

	bridge.delayBetweenRequests = delayBetweenRequests
}

func (bridge *Bridge) baseURL() string {
	if bridge.useHTTPS {
		return fmt.Sprintf("https://%s/api", bridge.IpAddr)
	}
	return fmt.Sprintf("http://%s/api", bridge.IpAddr)
}

func (bridge *Bridge) toURI(path string) string {
	if bridge.Username != "" {
		return fmt.Sprintf("%s/%s%s", bridge.baseURL(), bridge.Username, path)
	}
	return fmt.Sprintf("%s%s", bridge.baseURL(), path)
}

func (bridge *Bridge) get(path string, result interface{}) error {
	return bridge.do("GET", bridge.toURI(path), nil, result)
}

func (bridge *Bridge) post(path string, request interface{}, result interface{}) error {
	return bridge.do("POST", bridge.toURI(path), request, result)
}

func (bridge *Bridge) put(path string, request interface{}, result interface{}) error {
	return bridge.do("PUT", bridge.toURI(path), request, result)
}

func (bridge *Bridge) delete(path string, result interface{}) error {
	return bridge.do("DELETE", bridge.toURI(path), nil, result)
}

func (bridge *Bridge) do(method string, url string, request interface{}, result interface{}) error {
	bridge.lock.Lock()
	defer bridge.lock.Unlock()

	if bridge.delayBetweenRequests > 0 {
		// Enforce rate limit
		now := time.Now()
		nextRequest := bridge.lastRequestTimestamp.Add(bridge.delayBetweenRequests)
		if now.Before(nextRequest) {
			waitTime := time.Until(nextRequest)
			if bridge.debug {
				log.Printf("RATE LIMIT: Waiting %s until executing the next request", waitTime)
			}
			time.Sleep(waitTime)
		}
	}

	// Marshal request struct to JSON
	var body io.Reader
	var requestData []byte

	if request != nil {
		requestData, err := json.Marshal(request)
		if err != nil {
			return err
		}
		body = bytes.NewReader(requestData)
	}

	// Create HTTP request with JSON body
	httpRequest, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	// Execute request
	if bridge.debug {
		log.Printf("[%s] Request to %s (Body: %s)\n", method, url, requestData)
	}

	httpResponse, err := bridge.client.Do(httpRequest)
	bridge.lastRequestTimestamp = time.Now()
	if httpResponse != nil {
		defer httpResponse.Body.Close()
		defer io.Copy(ioutil.Discard, httpResponse.Body)
	}
	if err != nil {
		return err
	}

	// Decode response JSON to struct
	if result != nil {
		responseData, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			return err
		}

		if bridge.debug {
			log.Printf("[%s] Response to %s (Body: %s)\n", method, url, responseData)
		}

		err = json.Unmarshal(responseData, result)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetNewLights retrieves the list lights we've seen since
// the last scan. Returns the new lights, lastseen and any error
// that may have occurred as per:
// http://developers.meethue.com/1_lightsapi.html#12_get_new_lights
func (bridge *Bridge) GetNewLights() ([]*Light, string, error) {
	results := make(map[string]interface{})
	err := bridge.get("/lights/new", &results)
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
	var results []Result
	err := bridge.post("/lights", nil, &results)
	if err != nil {
		return nil, err
	}
	return results, err
}

// GetAllLights retrieves all devices the bridge is aware of
func (bridge *Bridge) GetAllLights() ([]*Light, error) {
	var result map[string]LightAttributes
	err := bridge.get("/lights", &result)
	if err != nil {
		return nil, err
	}

	// and convert them into lights
	var lights []*Light
	for id, attributes := range result {
		light := Light{Id: id, Name: attributes.Name, Attributes: attributes, bridge: bridge}
		lights = append(lights, &light)
	}

	return lights, nil
}
