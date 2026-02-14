package mlservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/creafly/logger"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type AnomalyCheckRequest struct {
	UserID                 string  `json:"user_id"`
	TenantID               string  `json:"tenant_id"`
	IPAddress              string  `json:"ip_address"`
	UserAgent              string  `json:"user_agent"`
	Endpoint               string  `json:"endpoint"`
	Method                 string  `json:"method"`
	RequestCountLastMinute int     `json:"request_count_last_minute"`
	RequestCountLastHour   int     `json:"request_count_last_hour"`
	UniqueIPsLastHour      int     `json:"unique_ips_last_hour"`
	FailedAuthAttempts     int     `json:"failed_auth_attempts"`
	GeoCountry             *string `json:"geo_country,omitempty"`
	GeoCity                *string `json:"geo_city,omitempty"`
	UsualCountry           *string `json:"usual_country,omitempty"`
}

type AnomalyCheckResponse struct {
	IsAnomaly         bool      `json:"is_anomaly"`
	AnomalyScore      float64   `json:"anomaly_score"`
	RiskLevel         RiskLevel `json:"risk_level"`
	Reasons           []string  `json:"reasons"`
	RecommendedAction string    `json:"recommended_action"`
}

type UpdateProfileRequest struct {
	UserID     string  `json:"user_id"`
	TenantID   string  `json:"tenant_id"`
	IPAddress  string  `json:"ip_address"`
	UserAgent  string  `json:"user_agent"`
	GeoCountry *string `json:"geo_country,omitempty"`
	Timestamp  *string `json:"timestamp,omitempty"`
}

type UserBehaviorProfile struct {
	UserID             string   `json:"user_id"`
	TenantID           string   `json:"tenant_id"`
	UsualIPs           []string `json:"usual_ips"`
	UsualCountries     []string `json:"usual_countries"`
	UsualUserAgents    []string `json:"usual_user_agents"`
	UsualActiveHours   [2]int   `json:"usual_active_hours"`
	AvgRequestsPerHour float64  `json:"avg_requests_per_hour"`
	LastSeen           *string  `json:"last_seen"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

type Config struct {
	BaseURL string
	Timeout time.Duration
	Enabled bool
}

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		enabled: cfg.Enabled,
	}
}

func (c *Client) CheckAnomaly(ctx context.Context, req *AnomalyCheckRequest) (*AnomalyCheckResponse, error) {
	if !c.enabled {
		return nil, nil
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/anomaly/check",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Name", "identity")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("ML anomaly detection request failed")
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Warn().
			Int("status", resp.StatusCode).
			Msg("ML anomaly detection returned non-200 status")
		return nil, nil
	}

	var anomalyResp AnomalyCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&anomalyResp); err != nil {
		logger.Log.Warn().Err(err).Msg("Failed to decode ML anomaly response")
		return nil, nil
	}

	return &anomalyResp, nil
}

func (c *Client) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UserBehaviorProfile, error) {
	if !c.enabled {
		return nil, nil
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/anomaly/profile",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Name", "identity")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("ML profile update request failed")
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Warn().
			Int("status", resp.StatusCode).
			Msg("ML profile update returned non-200 status")
		return nil, nil
	}

	var profile UserBehaviorProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		logger.Log.Warn().Err(err).Msg("Failed to decode ML profile response")
		return nil, nil
	}

	return &profile, nil
}

func (c *Client) GetProfile(ctx context.Context, tenantID, userID string) (*UserBehaviorProfile, error) {
	if !c.enabled {
		return nil, nil
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/anomaly/profile/%s/%s", c.baseURL, tenantID, userID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Service-Name", "identity")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("ML profile get request failed")
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK {
		var profile *UserBehaviorProfile
		if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
			return nil, nil
		}
		return profile, nil
	}

	return nil, nil
}

func (c *Client) IsEnabled() bool {
	return c.enabled
}

func ShouldBlockLogin(resp *AnomalyCheckResponse) bool {
	if resp == nil {
		return false
	}
	return resp.RecommendedAction == "block"
}

func ShouldRequire2FA(resp *AnomalyCheckResponse) bool {
	if resp == nil {
		return false
	}
	return resp.RecommendedAction == "require_2fa"
}

func ShouldReview(resp *AnomalyCheckResponse) bool {
	if resp == nil {
		return false
	}
	return resp.RecommendedAction == "review"
}
