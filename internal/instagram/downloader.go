package instagram

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type APIRequest struct {
	URL string `json:"url"`
}

type APIResponse struct {
	Status   string `json:"status"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// ExtractURL gets the first Instagram URL found in the text.
func ExtractURL(text string) string {
	re := regexp.MustCompile(`(?i)https?://(www\.)?instagram\.com/(reel|reels|p)/[a-zA-Z0-9_-]+`)
	return re.FindString(text)
}

// FetchVideo calls the external API to get the download link and then downloads the video data.
// It tries up to 3 times before returning an error.
func FetchVideo(instaURL string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		log.Info().Str("url", instaURL).Int("attempt", attempt).Msg("Starting Instagram video fetch")

		videoData, err := fetchVideoOnce(instaURL)
		if err == nil {
			return videoData, nil
		}

		lastErr = err
		log.Warn().Err(err).Int("attempt", attempt).Msg("Attempt failed")
		if attempt < 3 {
			time.Sleep(2 * time.Second) // Wait a bit before retrying
		}
	}

	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

func fetchVideoOnce(instaURL string) ([]byte, error) {
	apiReq := APIRequest{URL: instaURL}
	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   45 * time.Second, // Robust timeout for 24/7 operation
	}

	apiURL := os.Getenv("INSTAGRAM_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("INSTAGRAM_API_URL environment variable is not set")
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding API response: %w", err)
	}

	if apiResp.URL == "" {
		return nil, fmt.Errorf("API did not return a URL: %+v", apiResp)
	}

	log.Debug().Str("video_url", apiResp.URL).Msg("Downloading video from API response")

	// 2. Download the video
	videoResp, err := httpClient.Get(apiResp.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading video from %s: %w", apiResp.URL, err)
	}
	defer videoResp.Body.Close()

	videoData, err := io.ReadAll(videoResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading video data: %w", err)
	}

	log.Info().Int("bytes", len(videoData)).Msg("Successfully downloaded Instagram video")
	return videoData, nil
}
