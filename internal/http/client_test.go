// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"resty.dev/v3"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClientSingleton(t *testing.T) {
	c1 := Default()
	c2 := Default()
	require.NotNil(t, c1)
	require.Equal(t, c1, c2, "Default should return the same singleton client")
}

func TestMakeCacheKey(t *testing.T) {
	c := &Client{}

	// 普通URL，无Query参数
	key, err := c.makeCacheKey("https://example.com/foo")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/foo", key)

	// 带乱序Query参数，确认排序生效
	key, err = c.makeCacheKey("https://example.com/foo?b=2&a=1")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/foo?a=1&b=2", key)

	// 错误URL
	_, err = c.makeCacheKey("http://%41:80/")
	assert.Error(t, err)
}

func TestGet_CacheHitAndMiss(t *testing.T) {
	c := &Client{
		cache: cache.New(defaultCacheTTL, defaultCacheTTL*2),
		resty: Default().resty, // 复用默认客户端的resty，不进行真正请求
	}

	ctx := context.Background()
	url := "https://example.com/resource"

	// 先写入缓存
	cacheKey, _ := c.makeCacheKey(url)
	expectedContent := []byte("cached content")
	c.cache.Set(cacheKey, expectedContent, cache.DefaultExpiration)

	// 命中缓存
	data, err := c.Get(ctx, url)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, data)

	// 模拟缓存未命中，用一个无法访问的URL模拟请求失败
	c.cache.Delete(cacheKey)
	_, err = c.Get(ctx, "http://invalid.url") // 预期失败
	assert.Error(t, err)
}

func TestHead(t *testing.T) {
	c := Default()
	ctx := context.Background()

	// 模拟一个正常返回 HEAD 请求的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("X-Test-Header", "value")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	// 测试正常请求
	headers, status, err := c.Head(ctx, server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "value", headers.Get("X-Test-Header"))

	// 模拟错误的地址（server 关闭后的地址等效于无效 URL）
	badServer := httptest.NewUnstartedServer(nil)
	_, _, err = c.Head(ctx, badServer.URL)
	assert.Error(t, err)
}

func TestClient_Get(t *testing.T) {
	ctx := context.Background()

	// 模拟响应内容
	content := []byte("hello world")

	// 启动本地 HTTP 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// 构建 Client
	c := &Client{
		resty: resty.New(),
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// 第一次请求，不命中缓存
	data, err := c.Get(ctx, server.URL)
	assert.NoError(t, err)
	assert.Equal(t, content, data)

	// 第二次请求，应命中缓存（可通过日志或缓存验证）
	dataCached, err := c.Get(ctx, server.URL)
	assert.NoError(t, err)
	assert.Equal(t, content, dataCached)
}

// Download 单测示例，简单测试创建目录和下载失败场景
func TestDownload(t *testing.T) {
	c := Default()
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 模拟一个 HTTP 文件服务器
	testContent := []byte("0123456789") // 10字节内容
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	// 测试目录创建失败（试图写入非法目录）
	_, err := c.Download(ctx, server.URL, "/root/invalid-dir", "file.txt")
	assert.Error(t, err)

	// 测试下载失败（模拟 404）
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer badServer.Close()

	_, err = c.Download(ctx, badServer.URL, tmpDir, "file.txt")
	assert.Error(t, err)

	// 测试成功路径
	filename := "testfile.dat"
	filepath := filepath.Join(tmpDir, filename)

	resultPath, err := c.Download(ctx, server.URL, tmpDir, filename)
	assert.NoError(t, err)
	assert.Equal(t, filepath, resultPath)

	// 文件确实被创建且大小为10字节
	fi, err := os.Stat(filepath)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testContent)), fi.Size())
}
