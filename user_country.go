package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)



func (m *MatomoClient) GetCountry(siteID string) ([]DataResponse, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "UserCountry.getCountry")
	params.Set("format", "JSON")
	params.Set("idSite", siteID)
	params.Set("period", "day")
	params.Set("date", "yesterday")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)
	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error on reading response body usercountry.getcountry: %v ", string(body))
	}

	var response []DataResponse

	json.Unmarshal([]byte(body), &response)

	return response, nil
}
