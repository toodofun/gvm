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

package java

import (
	"context"
	"testing"
)

func TestJava_Name(t *testing.T) {
	j := &Java{}
	if j.Name() != "java" {
		t.Errorf("expected java, got %s", j.Name())
	}
}

func TestJava_ListRemoteVersions(t *testing.T) {
	j := &Java{}
	versions, err := j.ListRemoteVersions(context.Background())
	if err != nil {
		t.Errorf("ListRemoteVersions error: %v", err)
	}

	if len(versions) == 0 {
		t.Errorf("expected remote versions, got 0")
	}
}

func TestJava_ListInstalledVersions(t *testing.T) {
	j := &Java{}
	versions, err := j.ListInstalledVersions(context.Background())
	if err != nil {
		t.Errorf("ListInstalledVersions error: %v", err)
	}
	// 如果没有安装版本，应该返回空列表而不是错误
	if len(versions) == 0 {
		t.Errorf("should return empty slice, not nil")
	}
}

func TestJava_GetDefaultVersion(t *testing.T) {
	j := &Java{}
	version := j.GetDefaultVersion(context.Background())
	if version == nil {
		t.Errorf("GetDefaultVersion should not return nil")
	}
}

func TestJava_Uninstall(t *testing.T) {
	j := &Java{}
	err := j.Uninstall(context.Background(), "non-existent-version")
	if err != nil {
		t.Errorf("Uninstall should not return error for non-existent version, got: %v", err)
	}
}
