package matomo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type UserCountryResponse struct {
	Label             string      `json:"label"`
	NbUniqVisitors    int         `json:"nb_uniq_visitors"`
	NbVisits          int64       `json:"nb_visits"`
	NbActions         json.Number `json:"nb_actions"`
	NbUsers           int         `json:"nb_users"`
	MaxActions        int         `json:"max_actions"`
	SumVisitLength    json.Number `json:"sum_visit_length"`
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

func (m *MatomoClient) GetCountryList(siteID string) ([]UserCountryResponse, error) {
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

	var response []UserCountryResponse

	json.Unmarshal([]byte(body), &response)

	return response, nil
}

type AudienceByCountryResponse struct {
	Country    string  `json:"country"`
	PageViews  int64   `json:"pageViews"`
	BounceRate float64 `json:"bounceRate"`
}

func (m *MatomoClient) GetAudienceByCountry(siteID string) ([]AudienceByCountryResponse, error) {
	audienceList, err := m.GetCountryList(siteID)
	if err != nil {
		return nil, err
	}

	var response []AudienceByCountryResponse
	for _, data := range audienceList {
		bounceCount, err := data.BounceCount.Float64()
		if err != nil {
			log.Fatalf("error on converting bounce count to float64 %v", err)
		}
		bounceRate := (bounceCount / float64(data.NbVisits)) * 100
		response = append(response, AudienceByCountryResponse{
			Country:    data.Label,
			PageViews:  data.NbVisits,
			BounceRate: bounceRate,
		})
	}

	return response, nil
}

func (m *MatomoClient) GetMockAudienceByCountry() ([]AudienceByCountryResponse, error) {
	var audienceList []UserCountryResponse
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
		bounceCount, err := data.BounceCount.Float64()
		if err != nil {
			log.Fatalf("error on converting bounce count to float64 %v", err)
		}

		bounceRate := (bounceCount / float64(data.NbVisits)) * 100
		response = append(response, AudienceByCountryResponse{
			Country:    data.Label,
			PageViews:  data.NbVisits,
			BounceRate: bounceRate,
		})
	}

	return response, nil
}
