// Copyright 2026 The Toodofun Authors
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

package exec

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SafeExecutor provides safe command execution, preventing command injection attacks
type SafeExecutor struct {
	// allowedCommands is a whitelist, only allowing predefined commands
	allowedCommands map[string]bool
}

// NewSafeExecutor creates a safe executor that only allows whitelisted commands
func NewSafeExecutor(allowedCommands []string) *SafeExecutor {
	whitelist := make(map[string]bool)
	for _, cmd := range allowedCommands {
		whitelist[cmd] = true
	}
	return &SafeExecutor{
		allowedCommands: whitelist,
	}
}

// Execute safely executes a command
func (e *SafeExecutor) Execute(ctx context.Context, cmd string, args ...string) error {
	// NEW: Validate command name doesn't contain dangerous characters
	for _, dangerousChar := range dangerousChars {
		if strings.Contains(cmd, dangerousChar) {
			return fmt.Errorf("command contains dangerous character '%s'", dangerousChar)
		}
	}

	// Check if command is in whitelist
	if !e.allowedCommands[cmd] {
		return fmt.Errorf("command not allowed: %s", cmd)
	}

	// Validate arguments don't contain shell metacharacters
	for _, arg := range args {
		for _, dangerousChar := range dangerousChars {
			if strings.Contains(arg, dangerousChar) {
				return fmt.Errorf("argument contains dangerous character '%s': %s", dangerousChar, arg)
			}
		}
	}

	// Use parameterized execution, not through shell
	execCmd := exec.CommandContext(ctx, cmd, args...)
	return execCmd.Run()
}

// ExecuteString parses and executes a command from string
func (e *SafeExecutor) ExecuteString(ctx context.Context, cmdString string) error {
	cmd, err := ParseCommand(cmdString)
	if err != nil {
		return err
	}

	return e.Execute(ctx, cmd.Name, cmd.Args...)
}
