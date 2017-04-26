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

import "time"
import "net"
import "strings"
import "errors"

const upnpTimeout = 3 * time.Second

// SSDP Payload - Make sure to keep linebreaks and indention untouched.
const ssdpPayload = `M-SEARCH * HTTP/1.1
HOST: 239.255.255.250:1900
ST: ssdp:all
MAN: ssdp:discover
MX: 2

`

func upnpDiscover(respondingHosts chan<- string) error {
	// Open listening port for incoming responses
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1900})
	if err != nil {
		return err
	}
	socket.SetDeadline(time.Now().Add(upnpTimeout))
	defer socket.Close()

	// Send out discovery request as broadcast
	rawBody := []byte(strings.Replace(ssdpPayload, "\n", "\r\n", -1))
	_, err = socket.WriteToUDP(rawBody, &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900})
	if err != nil {
		return err
	}

	// Loop over responses until timeout hits
	var origins []string // keep track of response origins (return each origin only once)
loop:
	for {
		// Read response
		buf := make([]byte, 8192)
		_, addr, err := socket.ReadFromUDP(buf)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				return err //legitimate error, not a timeout.
			}
			return nil // timeout
		}

		// Parse and validate response
		body := string(buf)
		valid, err := ssdpResponseValid(body, addr.IP)
		if err != nil || !valid {
			continue // Ignore response
		}

		// Filter responses from duplicate origins
		for _, origin := range origins {
			if origin == addr.IP.String() {
				continue loop // duplicate
			}
		}

		// Response seems valid and unique -> send to channel
		origins = append(origins, addr.IP.String())
		respondingHosts <- addr.IP.String()
	}
}

func ssdpResponseValid(body string, origin net.IP) (valid bool, err error) {
	/*
		Response example:

		HTTP/1.1 200 OK
		HOST: 239.255.255.250:1900
		EXT:
		CACHE-CONTROL: max-age=100
		LOCATION: http://192.168.178.241:80/description.xml
		SERVER: FreeRTOS/7.4.2 UPnP/1.0 IpBridge/1.10.0
		hue-bridgeid: 001788FFFE09A206
		ST: upnp:rootdevice
		USN: uuid:2f402f80-da50-11e1-9b23-00178809a206::upnp:rootdevice

		FROM: https://developers.meethue.com/documentation/changes-bridge-discovery
	*/

	// Validate header
	if !strings.Contains(body, "HTTP/1.1 200 OK") {
		return false, errors.New("Invalid SSDP response header")
	}

	lower := strings.ToLower(body)
	// Validate MUST fields (from UPnP Device Architecture 1.1)
	if !strings.Contains(lower, "usn") || !strings.Contains(lower, "st") {
		return false, errors.New("Invalid SSDP response")
	}

	// Hue bridges send string "IpBridge" in SERVER field
	// (see https://developers.meethue.com/documentation/hue-bridge-discovery)
	if !strings.Contains(lower, "ipbridge") {
		return false, errors.New("Origin is no hue bridge")
	}

	// Validate IP in LOCATION field
	if !strings.Contains(lower, "location") {
		return false, errors.New("Invalid hue bridge response")
	}
	s := strings.SplitAfter(lower, "location: ")
	location := strings.Split(s[1], "\n")[0]
	s = strings.SplitAfter(location, "http://")
	ip := strings.Split(s[1], ":")[0]

	if ip != origin.String() {
		return false, errors.New("Response and sender mismatch")
	}

	return true, nil
}
