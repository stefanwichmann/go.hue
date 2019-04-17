package hue

import (
	"strconv"
)

// Light encapsulates the controls for a specific philips hue light
type Light struct {
	Id         string
	Name       string
	Attributes LightAttributes
	bridge     *Bridge
}

// LightState encapsulates all attributes for a specific philips hue light state
type LightState struct {
	Hue       int       `json:"hue"`
	On        bool      `json:"on"`
	Effect    string    `json:"effect"`
	Alert     string    `json:"alert"`
	Bri       int       `json:"bri"`
	Sat       int       `json:"sat"`
	Ct        int       `json:"ct"`
	Xy        []float32 `json:"xy"`
	Reachable bool      `json:"reachable"`
	ColorMode string    `json:"colormode"`
}

// SetLightState encapsulates all attributes to set a light to a specific state
type SetLightState struct {
	// On/Off state of the light. On=true, Off=false
	On string

	// The brightness value to set the light to.
	// Brightness is a scale from 1 (the minimum the light is capable of) to 254 (the maximum).
	// Note: a brightness of 1 is not off.
	Bri string

	// The hue value to set light to.
	// The hue value is a wrapping value between 0 and 65535.
	// Both 0 and 65535 are red, 25500 is green and 46920 is blue.
	Hue string

	// Saturation of the light. 254 is the most saturated (colored) and 0 is the least saturated (white).
	Sat string

	// The x and y coordinates of a color in CIE color space.
	// The first entry is the x coordinate and the second entry is the y coordinate. Both x and y must be between 0 and 1.
	// If the specified coordinates are not in the CIE color space, the closest color to the coordinates will be chosen.
	Xy []float32

	// The Mired Color temperature of the light. 2012 connected lights are capable of 153 (6500K) to 500 (2000K).
	Ct string

	// The alert effect, is a temporary change to the bulb’s state, and has one of the following values:
	// “none” – The light is not performing an alert effect.
	// “select” – The light is performing one breathe cycle.
	// “lselect” – The light is performing breathe cycles for 15 seconds or until an "alert": "none" command is received.
	//
	// Note that this contains the last alert sent to the light and not its current state.
	// i.e. After the breathe cycle has finished the bridge does not reset the alert to "none".
	Alert string

	// The dynamic effect of the light. Currently “none” and “colorloop” are supported. Other values will generate an error of type 7.
	// Setting the effect to colorloop will cycle through all hues using the current brightness and saturation settings.
	Effect string

	// The duration of the transition from the light’s current state to the new state.
	// This is given as a multiple of 100ms and defaults to 4 (400ms).
	// For example, setting transitiontime:10 will make the transition last 1 second.
	TransitionTime string
}

// LightAttributes encapsulates all attributes (hardware and state) for a specific philips hue light
type LightAttributes struct {
	State            LightState `json:"state"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	ModelId          string     `json:"modelid"`
	UniqueId         string     `json:"uniqueid"`
	ManufacturerName string     `json:"manufacturername"`
	ProductName      string     `json:"productname"`
	SoftwareVersion  string     `json:"swversion"`
}

// GetLightAttributes retrieves light attributes and state as per
// http://developers.meethue.com/1_lightsapi.html#14_get_light_attributes_and_state
func (light *Light) GetLightAttributes() (*LightAttributes, error) {
	var result LightAttributes
	err := light.bridge.get("/lights/"+light.Id, &result)
	if err != nil {
		return nil, err
	}

	// update locally cached attributes
	light.Attributes = result
	return &result, nil
}

// SetName sets the name of a light as per
// http://developers.meethue.com/1_lightsapi.html#15_set_light_attributes_rename
func (light *Light) SetName(newName string) ([]Result, error) {
	params := map[string]string{"name": newName}
	var results []Result
	err := light.bridge.put("/lights/"+light.Id, &params, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// On is a convenience method to turn on a light and set its effect to "none"
func (light *Light) On() ([]Result, error) {
	state := SetLightState{
		On:     "true",
		Effect: "none",
	}
	return light.SetState(state)
}

// Off is a convenience method to turn off a light
func (light *Light) Off() ([]Result, error) {
	state := SetLightState{On: "false"}
	return light.SetState(state)
}

// ColorLoop is a convenience method to turn on a light and have it begin
// a colorloop effect
func (light *Light) ColorLoop() ([]Result, error) {
	state := SetLightState{
		On:     "true",
		Effect: "colorloop",
	}
	return light.SetState(state)
}

// SetState sets the state of a light as per
// http://developers.meethue.com/1_lightsapi.html#16_set_light_state
func (light *Light) SetState(state SetLightState) ([]Result, error) {
	params := make(map[string]interface{})

	if state.On != "" {
		value, _ := strconv.ParseBool(state.On)
		params["on"] = value
	}
	if state.Bri != "" {
		params["bri"], _ = strconv.Atoi(state.Bri)
	}
	if state.Hue != "" {
		params["hue"], _ = strconv.Atoi(state.Hue)
	}
	if state.Sat != "" {
		params["sat"], _ = strconv.Atoi(state.Sat)
	}
	if state.Xy != nil {
		params["xy"] = state.Xy
	}
	if state.Ct != "" {
		params["ct"], _ = strconv.Atoi(state.Ct)
	}
	if state.Alert != "" {
		params["alert"] = state.Alert
	}
	if state.Effect != "" {
		params["effect"] = state.Effect
	}
	if state.TransitionTime != "" {
		params["transitiontime"], _ = strconv.Atoi(state.TransitionTime)
	}

	var results []Result
	err := light.bridge.put("/lights/"+light.Id+"/state", &params, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
