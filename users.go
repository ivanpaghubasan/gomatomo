package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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
