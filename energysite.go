package tesla

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CustomerPreferredExportRule string

const (
	PvOnly    CustomerPreferredExportRule = "pv_only"
	BatteryOk CustomerPreferredExportRule = "battery_ok"
)

type EnergySiteComponents struct {
	DisallowChargeFromGridWithSolarInstalled bool                        `json:"disallow_charge_from_grid_with_solar_installed"`
	CustomerPreferredExportRule              CustomerPreferredExportRule `json:"customer_preferred_export_rule"`
}

// this represents site_info endpoint
type EnergySite struct {
	ID                   string               `json:"id"`
	SiteName             string               `json:"site_name"`
	BackupReservePercent int64                `json:"backup_reserve_percent,omitempty"`
	DefaultRealMode      string               `json:"default_real_mode,omitempty"`
	Components           EnergySiteComponents `json:"components,omitempty"`

	productId int64
	c         *Client
}

type EnergySiteStatus struct {
	ResourceType      string  `json:"resource_type"`
	SiteName          string  `json:"site_name"`
	GatewayId         string  `json:"gateway_id"`
	EnergyLeft        float64 `json:"energy_left"`
	TotalPackEnergy   uint64  `json:"total_pack_energy"`
	PercentageCharged float64 `json:"percentage_charged"`
	BatteryType       string  `json:"battery_type"`
	BackupCapable     bool    `json:"backup_capable"`
	BatteryPower      int64   `json:"battery_power"`

	c *Client
}

type EnergySiteHistory struct {
	SerialNumber string                        `json:"serial_number"`
	Period       string                        `json:"period"`
	TimeSeries   []EnergySiteHistoryTimeSeries `json:"time_series"`

	c *Client
}

type EnergySiteHistoryTimeSeries struct {
	Timestamp                           time.Time `json:"timestamp"`
	SolarEnergyExported                 float64   `json:"solar_energy_exported"`
	GeneratorEnergyExported             float64   `json:"generator_energy_exported"`
	GridEnergyImported                  float64   `json:"grid_energy_imported"`
	GridServicesEnergyImported          float64   `json:"grid_services_energy_imported"`
	GridServicesEnergyExported          float64   `json:"grid_services_energy_exported"`
	GridEnergyExportedFromSolar         float64   `json:"grid_energy_exported_from_solar"`
	GridEnergyExportedFromGenerator     float64   `json:"grid_energy_exported_from_generator"`
	GridEnergyExportedFromBattery       float64   `json:"grid_energy_exported_from_battery"`
	BatteryEnergyExported               float64   `json:"battery_energy_exported"`
	BatteryEnergyImportedFromGrid       float64   `json:"battery_energy_imported_from_grid"`
	BatteryEnergyImportedFromSolar      float64   `json:"battery_energy_imported_from_solar"`
	BatteryEnergyImportedFromGenerator  float64   `json:"battery_energy_imported_from_generator"`
	ConsumerEnergyImportedFromGrid      float64   `json:"consumer_energy_imported_from_grid"`
	ConsumerEnergyImportedFromSolar     float64   `json:"consumer_energy_imported_from_solar"`
	ConsumerEnergyImportedFromBattery   float64   `json:"consumer_energy_imported_from_battery"`
	ConsumerEnergyImportedFromGenerator float64   `json:"consumer_energy_imported_from_generator"`
}

type SiteInfoResponse struct {
	Response *EnergySite `json:"response"`
}

type SiteStatusResponse struct {
	Response *EnergySiteStatus `json:"response"`
}

type SiteHistoryResponse struct {
	Response *EnergySiteHistory `json:"response"`
}

// SiteCommandResponse is the response from the Tesla API after POSTing a command.
type SiteCommandResponse struct {
	Response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"response"`
}

// return fetches the energy site for the given product ID
func (c *Client) EnergySite(productID int64) (*EnergySite, error) {
	siteInfoResponse := &SiteInfoResponse{}
	if err := c.getJSON(c.baseURL+"/energy_sites/"+strconv.FormatInt(productID, 10)+"/site_info", siteInfoResponse); err != nil {
		return nil, err
	}
	siteInfoResponse.Response.c = c
	siteInfoResponse.Response.productId = productID
	return siteInfoResponse.Response, nil
}

func (s *EnergySite) EnergySiteStatus() (*EnergySiteStatus, error) {
	siteStatusResponse := &SiteStatusResponse{}
	if err := s.c.getJSON(s.statusPath(), siteStatusResponse); err != nil {
		return nil, err
	}
	siteStatusResponse.Response.c = s.c
	return siteStatusResponse.Response, nil
}

type HistoryPeriod string

const (
	HistoryPeriodDay   HistoryPeriod = "day"
	HistoryPeriodWeek  HistoryPeriod = "week"
	HistoryPeriodMonth HistoryPeriod = "month"
	HistoryPeriodYear  HistoryPeriod = "year"
)

func (s *EnergySite) EnergySiteHistory(period HistoryPeriod) (*EnergySiteHistory, error) {
	historyResponse := &SiteHistoryResponse{}
	if err := s.c.getJSON(s.historyPath(period), historyResponse); err != nil {
		return nil, err
	}
	historyResponse.Response.c = s.c
	return historyResponse.Response, nil
}

func (s *EnergySite) basePath() string {
	return strings.Join([]string{s.c.baseURL, "energy_sites", strconv.FormatInt(s.productId, 10)}, "/")
}

func (s *EnergySite) statusPath() string {
	return strings.Join([]string{s.basePath(), "site_status"}, "/")
}

func (s *EnergySite) historyPath(period HistoryPeriod) string {
	v := url.Values{}
	v.Set("kind", "energy")
	v.Set("period", string(period))

	return strings.Join([]string{s.basePath(), "history"}, "/") + fmt.Sprintf("?%s", v.Encode())
}

func (s *EnergySite) SetGridCharging(enabled bool) error {
	url := s.basePath() + "/grid_import_export"
	payload := fmt.Sprintf(`{"disallow_charge_from_grid_with_solar_installed":%t}`, !enabled)

	if err := s.sendCommandExpectingEmptyResponse(url, []byte(payload)); err != nil {
		return err
	}

	return nil
}

func (s *EnergySite) SetExportRule(exportRule CustomerPreferredExportRule) error {
	if exportRule != PvOnly && exportRule != BatteryOk {
		return errors.New("invalid arg supplied to SetExportRule")
	}

	url := s.basePath() + "/grid_import_export"
	payload := fmt.Sprintf(`{"customer_preferred_export_rule": "%s"}`, exportRule)

	if err := s.sendCommandExpectingEmptyResponse(url, []byte(payload)); err != nil {
		return err
	}

	return nil
}

func (s *EnergySite) SetBatteryReserve(percent uint64) error {
	url := s.basePath() + "/backup"
	payload := fmt.Sprintf(`{"backup_reserve_percent":%d}`, percent)
	body, err := s.sendCommand(url, []byte(payload))
	if err != nil {
		return err
	}

	response := SiteCommandResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if response.Response.Code != 201 {
		return fmt.Errorf("batteryReserve failed: %s", response.Response.Message)
	}

	return nil
}

// Sends a command to the vehicle
func (s *EnergySite) sendCommand(url string, reqBody []byte) ([]byte, error) {
	body, err := s.c.post(url, reqBody)
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

func (s *EnergySite) sendCommandExpectingEmptyResponse(url string, reqBody []byte) error {
	body, err := s.c.post(url, reqBody)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		response := &CommandEmptyResponse{}
		if err := json.Unmarshal(body, response); err != nil {
			return err
		}
		if response.Response != "" {
			return errors.New(fmt.Sprintf("Unexpected response: %s", body))
		}
	}
	return nil
}
