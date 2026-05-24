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

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/toodofun/gvm/internal/util/exec"
	"github.com/toodofun/gvm/internal/util/path"
)

// TestSecurityModuleIntegration tests the integration between command executor and path safety
func TestSecurityModuleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gvm-security-integration-*")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	t.Run("SafeCommandExecutionWithinAllowedPaths", func(t *testing.T) {
		// Create executor with allowed commands
		allowedCommands := []string{"echo", "pwd", "ls"}
		executor := exec.NewSafeExecutor(allowedCommands)

		// Test safe command execution
		ctx := context.Background()
		err := executor.Execute(ctx, "echo", "hello", "world")
		assert.NoError(t, err, "Safe command should execute successfully")

		// Test command not in whitelist
		err = executor.Execute(ctx, "rm", "-rf", tempDir)
		assert.Error(t, err, "Non-whitelisted command should fail")
		assert.Contains(t, err.Error(), "command not allowed")
	})

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		// Test path safety check
		basePath := tempDir
		targetPath := "safe_directory"

		isSafe, err := path.IsPathSafe(basePath, targetPath)
		assert.NoError(t, err, "Path safety check should not error")
		assert.True(t, isSafe, "Safe path should pass validation")

		// Create the safe directory
		safeDir := filepath.Join(basePath, targetPath)
		err = os.MkdirAll(safeDir, 0755)
		assert.NoError(t, err, "Should create safe directory")

		// Test path traversal attempt
		maliciousPath := "../../../etc/passwd"
		isSafe, err = path.IsPathSafe(basePath, maliciousPath)
		assert.Error(t, err, "Path traversal should be detected")
		assert.False(t, isSafe, "Traversal path should fail validation")
		assert.Contains(t, err.Error(), "path traversal")
	})

	t.Run("CommandInjectionPrevention", func(t *testing.T) {
		allowedCommands := []string{"echo"}
		executor := exec.NewSafeExecutor(allowedCommands)

		ctx := context.Background()

		// Test various injection attempts
		injectionAttempts := []string{
			"echo hello; rm -rf /",
			"echo hello | cat",
			"echo hello && echo hacked",
			"echo hello `whoami`",
			"echo hello $(whoami)",
			"echo hello \x00 malicious",
		}

		for _, attempt := range injectionAttempts {
			err := executor.ExecuteString(ctx, attempt)
			assert.Error(t, err, "Injection attempt should fail: %s", attempt)
			assert.Contains(t, err.Error(), "dangerous character", "Error should mention dangerous character")
		}
	})

	t.Run("CombinedSafetyChecks", func(t *testing.T) {
		allowedCommands := []string{"ls"}
		executor := exec.NewSafeExecutor(allowedCommands)

		ctx := context.Background()

		// Create a test directory structure
		testDir := filepath.Join(tempDir, "test_data")
		err := os.MkdirAll(testDir, 0755)
		assert.NoError(t, err)

		// Test safe ls command in safe path
		err = executor.Execute(ctx, "ls", testDir)
		assert.NoError(t, err, "Safe command in safe path should work")

		// Test command with unsafe argument
		err = executor.Execute(ctx, "ls", "../../../etc")
		assert.Error(t, err, "Command with path traversal in args should fail")
	})
}
