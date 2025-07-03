package matomo

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"net/http"
	"net/url"
)

type MatomoClient struct {
	BaseURL    string
	TokenAuth  string
	ScriptHost string
}

type DataResponse struct {
	Label             string  `json:"label"`
	NbUniqVisitors    int64   `json:"nb_uniq_visitors"`
	NbVisits          int64   `json:"nb_visits"`
	NbActions         int64   `json:"nb_actions"`
	NbUsers           int64   `json:"nb_users"`
	MaxActions        int64   `json:"max_actions"`
	SumVisitLength    int64   `json:"sum_visit_length"`
	BounceCount       int64   `json:"bounce_count"`
	NbVisitsConverted int64   `json:"nb_visits_converted"`
	Goals             any     `json:"goals"`
	NbConversions     int64   `json:"nb_conversions"`
	Revenue           float64 `json:"revenue"`
	Code              string  `json:"us"`
	Logo              string  `json:"logo"`
	Segment           string  `json:"segment"`
	LogoHeight        int32   `json:"logoHeight"`
}

func InitClient(baseURL, tokenAuth, scriptHost string) *MatomoClient {
	return &MatomoClient{
		BaseURL:    baseURL,
		TokenAuth:  tokenAuth,
		ScriptHost: scriptHost,
	}
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

func generateSecurePassword() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "snap12345"
	}
	return base64.StdEncoding.EncodeToString(b)
}
