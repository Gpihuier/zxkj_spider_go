package helper

import (
	"testing"
)

func TestCompleteURL(t *testing.T) {
	baseURL := "https://www.sxsme.com.cn/"

	tests := []struct {
		name       string
		rawURL     string
		expected   string
		shouldFail bool
	}{
		{
			name:     "Relative path",
			rawURL:   "/xiazai/69176.html",
			expected: "https://www.sxsme.com.cn/xiazai/69176.html",
		},
		{
			name:     "Protocol-relative URL",
			rawURL:   "//sxsme.com/xiazai/69176.html",
			expected: "https://sxsme.com/xiazai/69176.html",
		},
		{
			name:     "Absolute URL with same domain",
			rawURL:   "https://www.sxsme.com.cn/game/1884.html",
			expected: "https://www.sxsme.com.cn/game/1884.html",
		},
		{
			name:     "Absolute URL with different domain",
			rawURL:   "https://downyi.com/page4",
			expected: "https://downyi.com/page4",
		},
		{
			name:       "Invalid URL",
			rawURL:     ":/invalid",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := CompleteURL(baseURL, tt.rawURL)
			if (err != nil) != tt.shouldFail {
				t.Fatalf("expected error: %v, got: %v", tt.shouldFail, err)
			}
			if !tt.shouldFail && actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}
