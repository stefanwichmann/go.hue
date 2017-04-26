// MIT License
//
// Copyright (c) 2017 Stefan Wichmann
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package hue

import "encoding/json"
import "net/http"

const nupnpEndpoint = "https://www.meethue.com/api/nupnp"

type nupnpBridge struct {
	Serial string `json:"id"`
	IpAddr string `json:"internalipaddress"`
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
		respondingHosts <- bridge.IpAddr
	}
	return nil
}
