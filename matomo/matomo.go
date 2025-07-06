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
	"strings"
)

type MatomoClient struct {
	BaseURL   string
	TokenAuth string
}

func NewClient(baseURL, tokenAuth string) *MatomoClient {
	return &MatomoClient{
		BaseURL:   baseURL,
		TokenAuth: tokenAuth,
	}
}

func (m *MatomoClient) AddSite(siteName, siteURL string) (string, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "SitesManager.addSite")
	params.Add("urls[]", siteURL)
	params.Set("siteName", siteName)
	params.Set("format", "json")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)
	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || strings.Contains(string(body), "error") {
		return "", fmt.Errorf("matomo addsite failed: %s", string(body))
	}
	var response struct {
		Value int `json:"value"`
	}
	_ = json.Unmarshal([]byte(body), &response)

	return strconv.Itoa(response.Value), nil
}

func (m *MatomoClient) SiteExists(siteURL string) (bool, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "SitesManager.getSitesIdFromSiteUrl")
	params.Set("url", siteURL)
	params.Set("format", "JSON")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)

	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("matomo api error: %s", string(body))
	}
	var response []struct {
		IdSite int `json:"idsite"`
	}

	_ = json.Unmarshal([]byte(body), &response)

	if len(response) > 0 {
		return true, nil
	}
	return false, nil
}

func (m *MatomoClient) CreateUser(userLogin, password, email string) error {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "UsersManager.addUser")
	params.Set("userLogin", userLogin)
	params.Set("password", password)
	params.Set("email", email)
	params.Set("alias", userLogin)
	params.Set("token_auth", m.TokenAuth)

	return m.callMatomo(params)
}

func (m *MatomoClient) UserExists(userEmail string) (bool, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "UsersManager.userEmailExists")
	params.Set("userEmail", userEmail)
	params.Set("format", "JSON")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)

	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("matomo api useremailexists error: %s", string(body))
	}

	var response struct {
		Value bool `json:"value"`
	}
	_ = json.Unmarshal([]byte(body), &response)

	return bool(response.Value), nil
}

func (m *MatomoClient) SetUserAccess(userLogin, siteID string) error {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "UsersManager.setUserAccess")
	params.Set("userLogin", userLogin)
	params.Set("access", "view")
	params.Set("idSites", siteID)
	params.Set("token_auth", m.TokenAuth)

	return m.callMatomo(params)
}

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

func (m *MatomoClient) GetTotalUsers(siteID string, days int) (int64, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "VisitsSummary.getUniqueVisitors")
	params.Set("idSite", siteID)
	params.Set("period", "day")
	params.Set("date", fmt.Sprintf("last%d", days))
	params.Set("format", "JSON")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)
	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error on reading response body VisitsSummary.getUniqueVisitors: %v ", string(body))
	}

	var data map[string]int64
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("error unmarshalling response: %v", err)
	}
	var total int64
	for _, v := range data {
		total += v
	}

	return total, nil
}

type Visit struct {
	IDVisit     string `json:"idVisit"`
	VisitorID   string `json:"visitorId"`
	VisitorType string `json:"visitorType"` // "new" or "returning"
	LastAction  int64  `json:"lastActionTimestamp"`
	VisitServer string `json:"serverTimePretty"`
}

func (m *MatomoClient) GetActiveUsers(siteID string, days int) (int, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "Live.getLastVisitsDetails")
	params.Set("idSite", siteID)
	params.Set("period", "range")
	params.Set("date", fmt.Sprintf("last%d", days))
	params.Set("format", "JSON")
	params.Set("token_auth", m.TokenAuth)

	resp, err := http.PostForm(fmt.Sprintf("%s/index.php", m.BaseURL), params)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("matomo api getlastvisitlastdetails error: %s", string(body))
	}

	var visits []Visit
	if err := json.Unmarshal([]byte(body), &visits); err != nil {
		return 0, fmt.Errorf("parse error: %v\nBody: %s", err, string(body))
	}

	uniqueVisitors := map[string]bool{}
	for _, v := range visits {
		uniqueVisitors[v.VisitorID] = true
	}

	return len(uniqueVisitors), nil
}

func (m *MatomoClient) GetNewUsers(siteID string, days int) (int, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "Live.getLastVisitsDetails")
	params.Set("idSite", siteID)
	params.Set("period", "range")
	params.Set("date", fmt.Sprintf("last%d", days))
	params.Set("format", "JSON")
	params.Set("token_auth", m.TokenAuth)

	resp, err := http.PostForm(fmt.Sprintf("%s/index.php", m.BaseURL), params)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("matomo api getlastvisitlastdetails error: getnewusers : %s", string(body))
	}

	var visits []Visit
	if err := json.Unmarshal([]byte(body), &visits); err != nil {
		return 0, fmt.Errorf("parse error: getnewusers : %v\nBody: %s", err, string(body))
	}

	newUsers := map[string]bool{}
	for _, v := range visits {
		if v.VisitorType == "new" {
			newUsers[v.VisitorID] = true
		}
	}

	return len(newUsers), nil
}

func (m *MatomoClient) ProvisionTelemetry(userName, userEmail, appName, appURL string) (siteID, userLogin, password string, err error) {
	userLogin = userName
	password = "password123"

	exists, err := m.UserExists(userLogin)
	if err != nil {
		return "", "", "", err
	}

	if !exists {
		if err := m.CreateUser(userLogin, password, userEmail); err != nil {
			return "", "", "", err
		}
	} else {
		password = ""
	}

	siteExist, err := m.SiteExists(appURL)
	if err != nil {
		return "", "", "", err
	}

	if siteExist {
		return "", "", "", fmt.Errorf("app url %s is already exists", appURL)
	}

	siteID, err = m.AddSite(appName, appURL)
	if err != nil {
		return "", "", "", err
	}

	if err := m.SetUserAccess(userLogin, siteID); err != nil {
		return "", "", "", err
	}

	return siteID, userLogin, password, nil
}

func (m *MatomoClient) callMatomo(params url.Values) error {
	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)
	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("matomo api error: %s", string(body))
	}

	return nil
}
