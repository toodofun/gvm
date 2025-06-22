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
	"os"
	"path/filepath"
	"testing"

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

	// 测试正常请求，使用一个确定存在的URL
	headers, status, err := c.Head(ctx, "https://httpbin.org/get")
	assert.NoError(t, err)
	assert.Greater(t, status, 0)
	assert.NotNil(t, headers)

	// 测试请求错误
	_, _, err = c.Head(ctx, "http://invalid.url")
	assert.Error(t, err)
}

// Download 单测示例，简单测试创建目录和下载失败场景
func TestDownload(t *testing.T) {
	c := Default()
	ctx := context.Background()

	tmpDir := t.TempDir()

	// 测试目录创建失败（试图写入非法目录）
	_, err := c.Download(ctx, "https://example.com/file", "/root/invalid-dir", "file.txt")
	assert.Error(t, err)

	// 测试下载失败（无效URL）
	_, err = c.Download(ctx, "http://invalid.url/file", tmpDir, "file.txt")
	assert.Error(t, err)

	// 测试成功路径，模拟一个小文件下载（使用 httpbin.org）
	url := "https://httpbin.org/bytes/10"
	filename := "testfile.dat"
	filepath := filepath.Join(tmpDir, filename)

	resultPath, err := c.Download(ctx, url, tmpDir, filename)
	assert.NoError(t, err)
	assert.Equal(t, filepath, resultPath)

	// 文件确实被创建且大小为10字节
	fi, err := os.Stat(filepath)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), fi.Size())
}
