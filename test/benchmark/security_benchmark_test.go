// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package benchmark

import (
	"context"
	"testing"

	"github.com/toodofun/gvm/internal/core/validation"
	"github.com/toodofun/gvm/internal/util/exec"
	"github.com/toodofun/gvm/internal/util/path"
)

func BenchmarkSafeExecutor_Execute(b *testing.B) {
	executor := exec.NewSafeExecutor([]string{"echo"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(ctx, "echo", "test")
	}
}

func BenchmarkParseCommand(b *testing.B) {
	cmdString := "tar -xzf file.tar.gz -C /tmp"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.ParseCommand(cmdString)
	}
}

func BenchmarkValidateVersion(b *testing.B) {
	version := "1.20.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidateVersion(version)
	}
}

func BenchmarkValidatePath(b *testing.B) {
	pathStr := "/usr/local/go/bin"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidatePath(pathStr)
	}
}

func BenchmarkIsPathSafe(b *testing.B) {
	basePath := "/tmp/gvm"
	targetPath := "versions/go1.20.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path.IsPathSafe(basePath, targetPath)
	}
}
