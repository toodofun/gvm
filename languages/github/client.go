package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Release struct {
	Name       string    `json:"name"`
	Prerelease bool      `json:"prerelease"`
	Assets     *[]Assets `json:"assets"`
}

type Assets struct {
	DownloadURL string `json:"browser_download_url"`
	Name        string `json:"name"`
}

// Client GitHub API 客户端
type Client struct {
	httpClient *http.Client
	token      string
}

// NewGitHubClient 创建新的 GitHub 客户端
func NewGitHubClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) GetAllReleases(ctx context.Context, owner, repo string) ([]Release, error) {
	var allReleases []Release
	page := 1
	perPage := 100

	for {
		releases, hasNext, err := c.getReleasesPage(ctx, owner, repo, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to get releases page %d: %w", page, err)
		}

		allReleases = append(allReleases, releases...)

		if !hasNext || len(releases) == 0 {
			break
		}

		page++
	}

	return allReleases, nil
}

// getReleasesPage 获取指定页的 releases
func (c *Client) getReleasesPage(ctx context.Context, owner, repo string, page, perPage int) ([]Release, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d&per_page=%d",
		owner, repo, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	// 设置认证头
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	// 设置 User-Agent
	req.Header.Set("User-Agent", "Go-GitHub-Releases-Fetcher/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, false, err
	}

	// 检查是否有下一页
	hasNext := c.hasNextPage(resp.Header.Get("Link"))

	return releases, hasNext, nil
}

// hasNextPage 检查 Link header 中是否有下一页
func (c *Client) hasNextPage(linkHeader string) bool {
	if linkHeader == "" {
		return false
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		if strings.Contains(link, `rel="next"`) {
			return true
		}
	}
	return false
}
