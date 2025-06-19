package http

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/schollz/progressbar/v3"
	"gvm/internal/log"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"resty.dev/v3"
	"strings"
	"sync"
	"time"
)

const (
	defaultCacheTTL = time.Minute * 5
)

var (
	client *Client
	once   sync.Once
)

type Client struct {
	resty *resty.Client
	cache *cache.Cache
}

func Default() *Client {
	once.Do(func() {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			Proxy:                 http.ProxyFromEnvironment, // 启用系统代理
		}
		c := resty.New().
			SetTransport(transport).
			SetRetryCount(3).
			SetRetryWaitTime(2 * time.Second).
			SetRetryMaxWaitTime(10 * time.Second).
			AddRetryHooks(func(response *resty.Response, err error) {
				log.Logger.Warnf("Retrying request to %s, attempt %d", response.Request.URL, response.Request.Attempt)
			})

		client = &Client{
			resty: c,
			cache: cache.New(defaultCacheTTL, defaultCacheTTL*2),
		}
	})

	return client
}

func (c *Client) makeCacheKey(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.RawQuery = parsed.Query().Encode() // 排序 query 参数
	return parsed.String(), nil
}

func (c *Client) Get(url string) ([]byte, error) {
	key, err := c.makeCacheKey(url)
	if err != nil {
		return nil, err
	}
	if val, found := c.cache.Get(key); found {
		log.Logger.Debugf("[cache] hit: %s", key)
		return val.([]byte), nil
	}

	resp, err := c.resty.R().Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Logger.Warnf("Close body error: %s", err)
		}
	}(resp.Body)
	if resp.IsError() {
		return nil, fmt.Errorf("response status: %s", resp.Status())
	}
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.cache.Set(key, res, defaultCacheTTL)

	return res, nil
}

func (c *Client) Head(url string) (http.Header, int, error) {
	resp, err := c.resty.R().Head(url)
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, resp.StatusCode(), nil
	}
	return resp.Header(), resp.StatusCode(), nil
}

func (c *Client) Download(url, destPath, filename string) (string, error) {
	file := path.Join(destPath, filename)
	if err := os.MkdirAll(path.Dir(file), os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", path.Dir(file), err)
	}

	var (
		out          *os.File
		err          error
		existingSize int64
	)

	if fi, err := os.Stat(file); err == nil {
		existingSize = fi.Size()
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat failed: %w", err)
	}

	supportsRange := false
	totalSize := int64(0)

	checkClient := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(2)

	resp, err := checkClient.R().
		SetHeader("Range", "bytes=0-1"). // 改为0-1，更标准
		Head(url)

	if err == nil && resp.StatusCode() == http.StatusPartialContent {
		supportsRange = true
		contentRange := resp.Header().Get("Content-Range")
		if contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				_, _ = fmt.Sscanf(parts[1], "%d", &totalSize)
			}
		}
	}

	if supportsRange && totalSize > 0 && existingSize >= totalSize {
		//logrus.Info("File already fully downloaded")
		return file, nil
	}

	if !supportsRange {
		existingSize = 0
	}

	if existingSize > 0 {
		out, err = os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		out, err = os.Create(file)
	}
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Logger.Warnf("Failed to close file %s: %+v", file, err)
		}
	}(out)

	downloadClient := resty.New().
		SetTimeout(0). // 下载不设超时
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second)

	request := downloadClient.R().
		SetDoNotParseResponse(true)

	if supportsRange && existingSize > 0 {
		request.SetHeader("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	resp, err = request.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Logger.Warnf("Failed to close response body: %+v", err)
		}
	}(resp.RawResponse.Body)

	switch resp.StatusCode() {
	case http.StatusOK:
		// 完整下载，正常情况
	case http.StatusPartialContent:
		// 断点续传，正常情况
	case http.StatusRequestedRangeNotSatisfiable: // 416错误
		// Range请求超出文件大小，说明文件已经完整下载
		if existingSize > 0 {
			log.Logger.Info("File already fully downloaded")
			return file, nil
		}
		return "", fmt.Errorf("range not satisfiable: %d", resp.StatusCode())
	default:
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if !supportsRange {
		if size := resp.Header().Get("Content-Length"); size != "" {
			_, _ = fmt.Sscanf(size, "%d", &totalSize)
		}
	}

	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(log.Logger.Out),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	_ = bar.Set64(existingSize)

	writer := io.MultiWriter(out, bar)
	_, err = io.Copy(writer, resp.RawResponse.Body)
	if err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}

	return file, nil
}
