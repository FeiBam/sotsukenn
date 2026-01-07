package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sotsukenn/go/types"
)

// FrigateClient handles communication with Frigate API
type FrigateClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewFrigateClient creates a new Frigate API client
func NewFrigateClient(baseURL string) *FrigateClient {
	return &FrigateClient{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Login authenticates with Frigate API and returns token from set-cookie
func (fc *FrigateClient) Login(username, password string) (string, error) {
	loginReq := types.FrigateLoginRequest{
		User:     username,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := fmt.Sprintf("%s/api/login", fc.BaseURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("frigate login failed: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Extract frigate_token from set-cookie header
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "frigate_token" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("no frigate_token cookie found in response")
}

// VerifyToken checks if the provided token is valid using Bearer token
func (fc *FrigateClient) VerifyToken(token string) (bool, error) {
	url := fmt.Sprintf("%s/api/auth", fc.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Use Authorization: Bearer <token> header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 200-299 status codes indicate valid token
	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

// GetCameras retrieves all cameras from Frigate
func (fc *FrigateClient) GetCameras(token string) (types.CamerasResponse, error) {
	url := fmt.Sprintf("%s/api/cameras", fc.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get cameras: %s (status: %d)", string(body), resp.StatusCode)
	}

	var cameras types.CamerasResponse
	if err := json.NewDecoder(resp.Body).Decode(&cameras); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return cameras, nil
}

// GetStreamURL returns the URL for a camera stream
// streamType can be: "mjpg" (MJPEG), "mp4" (H.264), "rtsp" (RTSP)
func (fc *FrigateClient) GetStreamURL(cameraName, streamType string) string {
	return fmt.Sprintf("%s/%s.%s", fc.BaseURL, cameraName, streamType)
}

// GetSnapshotURL returns the URL for a camera snapshot
func (fc *FrigateClient) GetSnapshotURL(cameraName string) string {
	return fmt.Sprintf("%s/api/%s/snapshot.jpg", fc.BaseURL, cameraName)
}

// GetSnapshot retrieves a snapshot image from the specified camera using Bearer token authentication
func (fc *FrigateClient) GetSnapshot(cameraName, token string) ([]byte, error) {
	url := fc.GetSnapshotURL(cameraName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get snapshot: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetStreams retrieves all available streams from Frigate
func (fc *FrigateClient) GetStreams(token string) (types.GetStreamsResponse, error) {
	url := fmt.Sprintf("%s/api/streams", fc.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get streams: %s (status: %d)", string(body), resp.StatusCode)
	}

	var streams types.GetStreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&streams); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return streams, nil
}

// GetGo2RTCStreams retrieves all go2rtc streams using HTTP Basic Auth
// username and password are used for Basic Authentication
func (fc *FrigateClient) GetGo2RTCStreams(username, password string) (types.Go2RTCStreamsResponse, error) {
	url := fmt.Sprintf("%s/api/go2rtc/streams", fc.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth
	req.SetBasicAuth(username, password)

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get go2rtc streams: %s (status: %d)", string(body), resp.StatusCode)
	}

	var streams types.Go2RTCStreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&streams); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return streams, nil
}

// GetGo2RTCStreamURL returns the go2rtc stream URL for a specific stream name
// format can be: "mp4", "mjpg", "webrtc", "rtsp"
func (fc *FrigateClient) GetGo2RTCStreamURL(streamName, format string) string {
	return fmt.Sprintf("%s/api/go2rtc/%s.%s", fc.BaseURL, streamName, format)
}

// GetGo2RTCStreamsWithToken retrieves all go2rtc streams using Bearer token
func (fc *FrigateClient) GetGo2RTCStreamsWithToken(token string) (types.Go2RTCStreamsResponse, error) {
	url := fmt.Sprintf("%s/api/go2rtc/streams", fc.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Bearer token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get go2rtc streams: %s (status: %d)", string(body), resp.StatusCode)
	}

	var streams types.Go2RTCStreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&streams); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return streams, nil
}

// GetLatestSnapshotURL returns the URL for the latest snapshot from a camera
func (fc *FrigateClient) GetLatestSnapshotURL(cameraName string) string {
	return fmt.Sprintf("%s/api/%s/latest.jpg", fc.BaseURL, cameraName)
}

// GetLatestSnapshot retrieves the latest snapshot image from the specified camera
// Returns image data, Content-Type, and error
func (fc *FrigateClient) GetLatestSnapshot(cameraName, token string) ([]byte, string, error) {
	url := fc.GetLatestSnapshotURL(cameraName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to get snapshot: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	return data, contentType, nil
}

// GetMJPEGStreamURL returns the MJPEG stream URL for a camera
// Note: Authentication should be done via Authorization header with Bearer token
func (fc *FrigateClient) GetMJPEGStreamURL(cameraName string) string {
	return fmt.Sprintf("%s/api/%s", fc.BaseURL, cameraName)
}
