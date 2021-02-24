package tesla

import (
	"encoding/json"
	"errors"
	"strconv"
)

// CommandResponse is the response from the Tesla API after POSTing a command.
type CommandResponse struct {
	Response struct {
		Reason string `json:"reason"`
		Result bool   `json:"result"`
	} `json:"response"`
}

// AutoParkRequest are the required elements to POST an Autopark/Summon request for the vehicle.
type AutoParkRequest struct {
	VehicleID uint64  `json:"vehicle_id,omitempty"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Action    string  `json:"action,omitempty"`
}

// SentryData shows whether Sentry is on.
type SentryData struct {
	Mode string `json:"on"`
}

// AutoparkAbort causes the vehicle to abort the Autopark request.
func (v Vehicle) AutoparkAbort() error {
	return v.autoPark("abort")
}

// AutoparkForward causes the vehicle to pull forward.
func (v Vehicle) AutoparkForward() error {
	return v.autoPark("start_forward")
}

// AutoparkReverse causes the vehicle to go in reverse.
func (v Vehicle) AutoparkReverse() error {
	return v.autoPark("start_reverse")
}

// Performs the actual auto park/summon request for the vehicle
func (v Vehicle) autoPark(action string) error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/autopark_request"
	driveState, _ := v.DriveState()
	autoParkRequest := &AutoParkRequest{
		VehicleID: v.VehicleID,
		Lat:       driveState.Latitude,
		Lon:       driveState.Longitude,
		Action:    action,
	}
	body, _ := json.Marshal(autoParkRequest)

	_, err := v.sendCommand(apiUrl, body)
	return err
}

// EnableSentry enables Sentry Mode
func (v *Vehicle) EnableSentry() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/set_sentry_mode"
	sentryRequest := &SentryData{
		Mode: "true",
	}

	body, _ := json.Marshal(sentryRequest)
	_, err := v.sendCommand(apiUrl, body)
	return err
}

// TBD based on Github issue #7
// Toggles defrost on and off, locations values are 'front' or 'rear'
// func (v Vehicle) Defrost(location string, state bool) error {
// 	command := location + "_defrost_"
// 	if state {
// 		command += "on"
// 	} else {
// 		command += "off"
// 	}
// 	apiUrl := v.c.URL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/" + command
// 	fmt.Println(apiUrl)
// 	_, err := v.sendCommand(apiUrl, nil)
// 	return err
// }

// TriggerHomelink opens and closes the configured Homelink garage door of the vehicle
// keep in mind this is a toggle and the garage door state is unknown
// a major limitation of Homelink.
func (v Vehicle) TriggerHomelink() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/trigger_homelink"
	driveState, _ := v.DriveState()
	autoParkRequest := &AutoParkRequest{
		Lat: driveState.Latitude,
		Lon: driveState.Longitude,
	}
	body, _ := json.Marshal(autoParkRequest)

	_, err := v.sendCommand(apiUrl, body)
	return err
}

// Wakeup wakes up the vehicle when it is powered off.
func (v Vehicle) Wakeup() (*Vehicle, error) {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/wake_up"
	body, err := v.sendCommand(apiUrl, nil)
	if err != nil {
		return nil, err
	}
	vehicleResponse := &VehicleResponse{}
	if err := json.Unmarshal(body, vehicleResponse); err != nil {
		return nil, err
	}
	return vehicleResponse.Response, nil
}

// Opens the charge port so you may insert your charging cable
func (v Vehicle) OpenChargePort() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/charge_port_door_open"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Resets the PIN set for valet mode, if set
func (v Vehicle) ResetValetPIN() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/reset_valet_pin"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Sets the charge limit to the standard setting
func (v Vehicle) SetChargeLimitStandard() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/charge_standard"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Sets the charge limit to the max limit
func (v Vehicle) SetChargeLimitMax() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/charge_max_range"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Set the charge limit to a custom percentage
func (v Vehicle) SetChargeLimit(percent int) error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/set_charge_limit"
	theJson := `{"percent": ` + strconv.Itoa(percent) + `}`
	_, err := v.c.post(apiUrl, []byte(theJson))
	return err
}

// StartCharging starts the charging of the vehicle after you have inserted the
// charging cable
func (v Vehicle) StartCharging() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/charge_start"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Stop the charging of the vehicle
func (v Vehicle) StopCharging() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/charge_stop"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Flashes the lights of the vehicle
func (v Vehicle) FlashLights() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/flash_lights"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Honks the horn of the vehicle
func (v *Vehicle) HonkHorn() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/honk_horn"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Unlock the car's doors
func (v Vehicle) UnlockDoors() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/door_unlock"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Locks the doors of the vehicle
func (v Vehicle) LockDoors() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/door_lock"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

type tempRequest struct {
	DriverTemp    string `json:"driver_temp"`
	PassengerTemp string `json:"passenger_temp"`
}

// Sets the temperature of the vehicle, where you may set the driver
// zone and the passenger zone to seperate temperatures
func (v Vehicle) SetTemperature(driver float64, passenger float64) error {
	driveTemp := strconv.FormatFloat(driver, 'f', -1, 32)
	passengerTemp := strconv.FormatFloat(passenger, 'f', -1, 32)
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/set_temps"
	b, err := json.Marshal(&tempRequest{driveTemp, passengerTemp})
	if err != nil {
		return err
	}
	_, err = v.c.post(apiUrl, b)
	return err
}

// StartAirConditioning starts the air conditioning in the car
func (v Vehicle) StartAirConditioning() error {
	url := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/auto_conditioning_start"
	_, err := v.sendCommand(url, nil)
	return err
}

// Stops the air conditioning in the car
func (v Vehicle) StopAirConditioning() error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/auto_conditioning_stop"
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// The desired state of the panoramic roof. The approximate percent open
// values for each state are open = 100%, close = 0%, comfort = 80%, vent = %15, move = set %
func (v Vehicle) MovePanoRoof(state string, percent int) error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/sun_roof_control"
	theJson := `{"state": "` + state + `", "percent":` + strconv.Itoa(percent) + `}`
	_, err := v.c.post(apiUrl, []byte(theJson))
	return err
}

// Start starts the car by turning it on, requires the password to be sent
// again
func (v Vehicle) Start(password string) error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/remote_start_drive?password=" + password
	_, err := v.sendCommand(apiUrl, nil)
	return err
}

// Opens the trunk, where values may be 'front' or 'rear'
func (v Vehicle) OpenTrunk(trunk string) error {
	apiUrl := v.c.BaseURL + "/vehicles/" + strconv.FormatInt(v.ID, 10) + "/command/trunk_open" // ?which_trunk=" + trunk
	theJson := `{"which_trunk": "` + trunk + `"}`
	_, err := v.c.post(apiUrl, []byte(theJson))
	return err
}

// Sends a command to the vehicle
func (v *Vehicle) sendCommand(url string, reqBody []byte) ([]byte, error) {
	body, err := v.c.post(url, reqBody)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		response := &CommandResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return nil, err
		}
		if !response.Response.Result && response.Response.Reason != "" {
			return nil, errors.New(response.Response.Reason)
		}
	}
	return body, nil
}
