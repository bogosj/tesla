package tesla

type RegionResponse struct {
	Region          string `json:"region"`
	FleetApiBaseUrl string `json:"fleet_api_base_url"`
}

// UserRegion fetches the users region
func (c *Client) UserRegion() (*RegionResponse, error) {
	var regionResponse RegionResponse
	if err := c.getJSON(c.baseURL+"/users/region", &regionResponse); err != nil {
		return nil, err
	}
	return &regionResponse, nil
}
