package matomo

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"net/http"
	"net/url"
	"strings"
)

type MatomoClient struct {
	BaseURL    string
	TokenAuth  string
	ScriptHost string
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

func (m *MatomoClient) GenerateTrackingScript(siteID string) string {
	return fmt.Sprintf(`
        <script>
        var _paq = window._paq = window._paq || [];
        _paq.push(['trackPageView']);
        _paq.push(['enableLinkTracking']);
        (function() {
            var u="%s/";
            _paq.push(['setTrackerUrl', u+'matomo.php']);
            _paq.push(['setSiteId', '%s']);
            var d=document, g=d.createElement('script'), s=d.getElementsByTagName('script')[0];
            g.async=true; g.src="%s"; s.parentNode.insertBefore(g,s);
        })();
        </script>
        `, m.BaseURL, siteID, m.ScriptHost)
}

func (m *MatomoClient) ProvisionTelemetry(userName, userEmail, appName, appURL string) (siteID, userLogin, password, script string, err error) {
	userLogin = userName
	password = "password123"

	exists, err := m.UserExists(userLogin)
	if err != nil {
		return "", "", "", "", err
	}

	if !exists {
		if err := m.CreateUser(userLogin, password, userEmail); err != nil {
			return "", "", "", "", err
		}
	} else {
		password = ""
	}

	siteExist, err := m.SiteExists(appURL)
	if err != nil {
		return "", "", "", "", err
	}

	if siteExist {
		return "", "", "", "", fmt.Errorf("app url %s is already exists", appURL)
	}

	siteID, err = m.AddSite(appName, appURL)
	if err != nil {
		return "", "", "", "", err
	}

	if err := m.SetUserAccess(userLogin, siteID); err != nil {
		return "", "", "", "", err
	}

	script = m.GenerateTrackingScript(siteID)
	return siteID, userLogin, password, script, nil
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

func (m *MatomoClient) InviteUser(userLogin, userEmail string) error {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "UsersManager.inviteUser")
	params.Set("userLogin", userLogin)
	params.Set("email", userEmail)
	params.Set("format", "JSON")
	params.Set("initialIdSite", "1")
	params.Set("token_auth", m.TokenAuth)

	endpoint := fmt.Sprintf("%s/index.php", m.BaseURL)
	resp, err := http.PostForm(endpoint, params)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	_ = response
	fmt.Printf("response body: %s", string(body))

	return nil

}

func generateSecurePassword() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "snap12345"
	}
	return base64.StdEncoding.EncodeToString(b)
}
