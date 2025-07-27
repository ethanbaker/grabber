package grabber_test

import (
	"testing"

	"github.com/ethanbaker/grabber"
)

// TestIdentifyPlatform tests the platform identification functionality
func TestIdentifyPlatform(t *testing.T) {
	g := grabber.New()

	tests := []struct {
		name     string
		url      string
		expected grabber.Platform
	}{
		// TikTok tests
		{
			name:     "TikTok video URL",
			url:      "https://www.tiktok.com/@user/video/1234567890",
			expected: grabber.Tiktok,
		},
		{
			name:     "TikTok vm URL",
			url:      "https://vm.tiktok.com/abcd1234",
			expected: grabber.Tiktok,
		},
		// Instagram tests
		{
			name:     "Instagram reel URL",
			url:      "https://www.instagram.com/reel/ABC123/",
			expected: grabber.InstagramReels,
		},
		// Twitter/X tests
		{
			name:     "Twitter URL",
			url:      "https://twitter.com/user/status/1234567890",
			expected: grabber.Twitter,
		},
		{
			name:     "X.com URL",
			url:      "https://x.com/user/status/1234567890",
			expected: grabber.Twitter,
		},
		// YouTube tests
		{
			name:     "YouTube video URL",
			url:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: grabber.YouTube,
		},
		{
			name:     "YouTube Shorts URL",
			url:      "https://www.youtube.com/shorts/dQw4w9WgXcQ",
			expected: grabber.YouTube,
		},
		{
			name:     "YouTube short URL (youtu.be)",
			url:      "https://youtu.be/dQw4w9WgXcQ",
			expected: grabber.YouTube,
		},
		{
			name:     "YouTube mobile URL",
			url:      "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: grabber.YouTube,
		},
		// Unknown platform tests
		{
			name:     "Unknown platform",
			url:      "https://www.example.com/video/123",
			expected: grabber.Unknown,
		},
		{
			name:     "Invalid URL",
			url:      "not-a-url",
			expected: grabber.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.IdentifyPlatform(tt.url)
			if result != tt.expected {
				t.Errorf("IdentifyPlatform(%s) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}
