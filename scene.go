package hue

import "encoding/json"
import "bytes"
import "fmt"
import "errors"

// Scene represents a Hue scene saved on the bridge.
type Scene struct {
	bridge      *Bridge
	Id          string                 `json:"-"`
	Name        string                 `json:"name"`
	Lights      []string               `json:"lights"`
	Owner       string                 `json:"owner"`
	Recycle     bool                   `json:"recycle"`
	Locked      bool                   `json:"locked"`
	Appdata     map[string]interface{} `json:"appdata"`
	Picture     string                 `json:"picture"`
	LastUpdated string                 `json:"lastupdated"`
	Version     int                    `json:"version"`
	LightStates map[string]LightState  `json:"lightstates"`
}

// CreateScene contains all necessary attributes to create a new scene on the bridge.
type CreateScene struct {
	Name           string                 `json:"name,omitempty"`
	Lights         []string               `json:"lights,omitempty"`
	Recycle        bool                   `json:"recycle"`
	TransitionTime int                    `json:"transistiontime,omitempty"`
	Appdata        map[string]interface{} `json:"appdata,omitempty"`
	Picture        string                 `json:"picture,omitempty"`
}

// ModifyScene contains all attributes to be changed on a given scene.
type ModifyScene struct {
	Name            string   `json:"name,omitempty"`
	Lights          []string `json:"lights,omitempty"`
	StoreLightState bool     `json:"storelightstate,omitempty"`
}

type ModifyLightState struct {
	On               bool      `json:"on,omitempty"`
	Brightness       uint8     `json:"bri,omitempty"`
	Hue              uint16    `json:"hue,omitempty"`
	Saturation       uint8     `json:"sat,omitempty"`
	Xy               []float32 `json:"xy,omitempty"`
	ColorTemperature uint16    `json:"ct,omitempty"`
	Effect           string    `json:"effect,omitempty"`
	TransitionTime   uint16    `json:"transistiontime,omitempty"`
}

// CreateScene stores a new scene with the given attributes on the bridge.
// In addition to the given information the current light states of all referenced
// lights will be part of the scene.
func (bridge *Bridge) CreateScene(scenedata CreateScene) ([]Result, error) {
	data, err := json.Marshal(scenedata)
	if err != nil {
		return nil, err
	}

	response, err := bridge.post("/scenes/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	return results, err
}

// AllScenes returns all scenes currently saved on the bridge.
func (bridge *Bridge) AllScenes() ([]*Scene, error) {
	var scenes []*Scene
	response, err := bridge.get("/scenes")
	if err != nil {
		return scenes, err
	}
	defer response.Body.Close()

	// deconstruct the json results
	var results map[string]Scene
	err = json.NewDecoder(response.Body).Decode(&results)
	if err != nil {
		return scenes, err
	}

	// and convert them into scenes
	for id, scene := range results {
		scene := scene
		scene.Id = id
		scene.bridge = bridge
		if err != nil {
			return scenes, err
		}
		scenes = append(scenes, &scene)
	}

	return scenes, nil
}

// SceneByID looks up the scene with the given ID on the bridge.
func (bridge *Bridge) SceneByID(id string) (*Scene, error) {
	response, err := bridge.get(fmt.Sprintf("/scenes/%s", id))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// deconstruct the json results
	var result Scene
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	result.Id = id
	result.bridge = bridge

	return &result, nil
}

// SceneByName looks up the scene with the given name on the bridge.
func (bridge *Bridge) SceneByName(name string) (*Scene, error) {
	scenes, err := bridge.AllScenes()
	if err != nil {
		return nil, err
	}

	for _, scene := range scenes {
		if scene.Name == name {
			return bridge.SceneByID(scene.Id) // second request to fill lightstates
		}
	}

	return nil, errors.New("Unable to find scene with name " + name)
}

// Modify adjusts a saved scene according to the given attributes.
func (scene *Scene) Modify(modifyScene ModifyScene) ([]Result, error) {
	data, err := json.Marshal(modifyScene)
	if err != nil {
		return nil, err
	}

	response, err := scene.bridge.put("/scenes/"+scene.Id, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	return results, err
}

// ModifyLightStates adjusts the saved light states of all lights in the given scene on the bridge.
func (scene *Scene) ModifyLightStates(lightstate ModifyLightState) ([]Result, error) {
	var results []Result

	for _, light := range scene.Lights {
		result, err := scene.ModifyLightState(light, lightstate)
		results = append(results, result...)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// ModifyLightState adjusts the saved light state of the given light in the given scene on the bridge.
func (scene *Scene) ModifyLightState(lightID string, lightstate ModifyLightState) ([]Result, error) {
	data, err := json.Marshal(lightstate)
	if err != nil {
		return nil, err
	}

	response, err := scene.bridge.put(fmt.Sprintf("/scenes/%s/lightstates/%s", scene.Id, lightID), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	return results, err
}

// Delete will remove the given scene from the bridge.
func (scene *Scene) Delete() ([]Result, error) {
	response, err := scene.bridge.delete("/scenes/" + scene.Id)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	fmt.Printf("RES: %v\n", results)
	return results, err
}

// Activate will recall the given scene according to it's state on the bridge.
func (scene *Scene) Activate() ([]Result, error) {
	request := map[string]string{"scene": scene.Id}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := scene.bridge.put("/groups/0/action", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var results []Result
	err = json.NewDecoder(response.Body).Decode(&results)
	return results, err
}
