package tesla

type RegionResponse struct {
	Response *Region `json:"response"`
}
type Region struct {
	Region          string `json:"region"`
	FleetApiBaseUrl string `json:"fleet_api_base_url"`
}

// UserRegion fetches the users region
func (c *Client) UserRegion() (*Region, error) {
	var regionResponse RegionResponse
	if err := c.getJSON(c.baseURL+"/users/region", &regionResponse); err != nil {
		return nil, err
	}
	return regionResponse.Response, nil
}
