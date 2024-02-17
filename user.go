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

type MeResponse struct {
	Response *Me `json:"response"`
}
type Me struct {
	Email           string `json:"email"`
	FullName        string `json:"full_name"`
	ProfileImageUrl string `json:"profile_image_url"`
}

// UserMe fetches the users me
func (c *Client) UserMe() (*Me, error) {
	var meResponse MeResponse
	if err := c.getJSON(c.baseURL+"/users/me", &meResponse); err != nil {
		return nil, err
	}
	return meResponse.Response, nil
}
