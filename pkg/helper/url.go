package helper

import (
	"fmt"
	"net/url"
)

// CompleteURL 补全URL的域名
// baseURL 为基础URL
// rawURL 为待补全的URL
func CompleteURL(baseURL, rawURL string) (string, error) {
	// 解析基础URL（包含协议和域名）
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %v", err)
	}

	// 解析爬取的URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse raw URL: %v", err)
	}

	// 如果没有携带域名，则补全
	if u.Scheme == "" {
		u.Scheme = base.Scheme
	}
	if u.Host == "" {
		u.Host = base.Host
	}

	// 生成完整的URL]
	return u.String(), nil
}
