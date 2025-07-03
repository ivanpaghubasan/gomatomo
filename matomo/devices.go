package matomo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type DevicesDetectionResponse struct {
	Label             string      `json:"label"`
	NbUniqVisitors    int         `json:"nb_uniq_visitors"`
	NbVisits          int64       `json:"nb_visits"`
	NbActions         json.Number `json:"nb_actions"`
	NbUsers           int         `json:"nb_users"`
	MaxActions        int         `json:"max_actions"`
	SumVisitLength    string      `json:"sum_visit_length"`
	BounceCount       json.Number `json:"bounce_count"`
	NbVisitsConverted json.Number `json:"nb_visits_converted"`
	Goals             any         `json:"goals"`
	NbConversions     int         `json:"nb_conversions"`
	Revenue           float32     `json:"revenue"`
	Code              string      `json:"us"`
	Logo              string      `json:"logo"`
	Segment           string      `json:"segment"`
	LogoHeight        int         `json:"logoHeight"`
}

func (m *MatomoClient) GetDeviceList(siteID string) ([]DevicesDetectionResponse, error) {
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

	var response []DevicesDetectionResponse

	json.Unmarshal([]byte(body), &response)

	return nil, nil
}

type SessionsByDeviceResponse struct {
	Device             string  `json:"device"`
	Visits             int64   `json:"visits"`
	AverageVisitLength float64 `json:"averageVisitLength"`
}

func (m *MatomoClient) GetSessionsByDevice(siteID string) ([]SessionsByDeviceResponse, error) {
	var deviceList []DevicesDetectionResponse
	deviceList, err := m.GetDeviceList(siteID)
	if err != nil {
		return nil, err
	}

	if len(deviceList) == 0 {
		file, err := os.Open("mock_device_list.json")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&deviceList); err != nil {
			return nil, err
		}
	}

	var response []SessionsByDeviceResponse
	for _, data := range deviceList {
		var sumVisitLength float64
		if data.SumVisitLength != "" {
			sumVisitLength, err = strconv.ParseFloat(data.SumVisitLength, 64)
			if err != nil {
				log.Fatalf("error on string conversion to float64 %v", err)
			}
		} else {
			sumVisitLength = 0
		}

		var averageVisitLength float64
		if data.NbVisits > 0 {
			averageVisitLength = (sumVisitLength / float64(data.NbVisits)) * 100
		} else {
			averageVisitLength = 0
		}

		response = append(response, SessionsByDeviceResponse{
			Device:             data.Label,
			Visits:             data.NbVisits,
			AverageVisitLength: averageVisitLength,
		})
	}

	return response, nil
}

func (m *MatomoClient) GetMockSessionsByDevice() ([]SessionsByDeviceResponse, error) {
	var deviceList []DevicesDetectionResponse

	file, err := os.Open("mock_device_list.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&deviceList); err != nil {
		return nil, err
	}

	var response []SessionsByDeviceResponse
	for _, data := range deviceList {
		var sumVisitLength float64
		if data.SumVisitLength != "" {
			sumVisitLength, err = strconv.ParseFloat(data.SumVisitLength, 64)
			if err != nil {
				log.Fatalf("error on string conversion to float64 %v", err)
			}
		} else {
			sumVisitLength = 0
		}

		var averageVisitLength float64
		if data.NbVisits > 0 {
			averageVisitLength = (sumVisitLength / float64(data.NbVisits)) * 100
		} else {
			averageVisitLength = 0
		}

		response = append(response, SessionsByDeviceResponse{
			Device:             data.Label,
			Visits:             data.NbVisits,
			AverageVisitLength: averageVisitLength,
		})
	}

	return response, nil
}
