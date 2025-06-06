package videos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GlobalClient is the shared YouTube client instance
var GlobalClient *YouTubeClient

type YouTubeClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

type Video struct {
	Title    string `json:"title"`
	VideoID  string `json:"videoId"`
	EmbedURL string `json:"embedUrl"`
}

func NewYouTubeClient(apiKey, baseURL string) *YouTubeClient {
	return &YouTubeClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// BuildSearchURL constructs the full URL for the search query
func (c *YouTubeClient) BuildSearchURL(query string, maxResults int) (string, error) {
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 5
	}

	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "video")
	params.Set("videoEmbeddable", "true")     // Only embeddable videos
	params.Set("videoCaption", "closedCaption") // Require captioned videos
	params.Set("safeSearch", "strict")         // Safe results
	params.Set("order", "relevance")
	params.Set("q", query)
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))
	params.Set("key", c.APIKey)

	return fmt.Sprintf("%s/search?%s", c.BaseURL, params.Encode()), nil
}

// SearchVideos calls the YouTube API and returns a list of embeddable captioned videos
func (c *YouTubeClient) SearchVideos(query string, maxResults int) ([]Video, error) {
	url, err := c.BuildSearchURL(query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call YouTube API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := make([]Video, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		videoID := item.ID.VideoID
		if videoID == "" {
			continue
		}
		videos = append(videos, Video{
			Title:    item.Snippet.Title,
			VideoID:  videoID,
			EmbedURL: fmt.Sprintf("https://www.youtube.com/embed/%s", videoID),
		})
		if len(videos) >= maxResults {
			break
		}
	}

	return videos, nil
}