package matomo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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
