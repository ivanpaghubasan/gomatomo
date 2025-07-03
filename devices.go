package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (m *MatomoClient) GetDeviceList(siteID string) ([]DataResponse, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "DevicesDetection.getType")
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
		return nil, fmt.Errorf("error on reading response body devicesdetection.gettype: %v ", string(body))
	}

	var response []DataResponse

	json.Unmarshal([]byte(body), &response)

	return nil, nil
}

type SessionsByDeviceResponse struct {
	Device             string  `json:"device"`
	Visits             int64   `json:"visits"`
	AverageVisitLength float64 `json:"averageVisitLength"`
}

func (m *MatomoClient) GetSessionsByDevice(siteID string) ([]SessionsByDeviceResponse, error) {
	deviceList, err := m.GetDeviceList(siteID)
	if err != nil {
		return nil, err
	}

	var response []SessionsByDeviceResponse
	for _, data := range deviceList {
		averageVisitLength := (data.SumVisitLength / data.NbVisits) * 100
		response = append(response, SessionsByDeviceResponse{
			Device:             data.Label,
			Visits:             data.NbVisits,
			AverageVisitLength: float64(averageVisitLength),
		})
	}

	return response, nil
}
