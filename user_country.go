package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func (m *MatomoClient) GetCountryList(siteID string) ([]DataResponse, error) {
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

type AudienceByCountryResponse struct {
	Country    string  `json:"country"`
	PageViews  int64   `json:"pageViews"`
	BounceRate float64 `json:"bounceRate"`
}

func (m *MatomoClient) GetAudienceByCountry(siteID string) ([]AudienceByCountryResponse, error) {
	audienceList, err := m.GetDeviceList(siteID)
	if err != nil {
		return nil, err
	}

	var response []AudienceByCountryResponse
	for _, data := range audienceList {
		bounceRate := (data.BounceCount / data.NbVisits) * 100
		response = append(response, AudienceByCountryResponse{
			Country:    data.Label,
			PageViews:  data.NbVisits,
			BounceRate: float64(bounceRate),
		})
	}

	return response, nil
}

func (m *MatomoClient) GetMockAudienceByCountry() ([]AudienceByCountryResponse, error) {
	var audienceList []DataResponse
	file, err := os.Open("mock_country_list.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&audienceList); err != nil {
		return nil, err
	}

	var response []AudienceByCountryResponse
	for _, data := range audienceList {
		bounceRate := (data.BounceCount / data.NbVisits) * 100
		response = append(response, AudienceByCountryResponse{
			Country:    data.Label,
			PageViews:  data.NbVisits,
			BounceRate: float64(bounceRate),
		})
	}

	return response, nil
}
