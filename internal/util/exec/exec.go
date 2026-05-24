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
	"fmt"
	"strings"
)

// Command represents a parsed safe command
type Command struct {
	Name string
	Args []string
}

// dangerousChars contains all dangerous shell metacharacters
var dangerousChars = []string{
	"|", "&", ";", "$", "(", ")", "<", ">", "`", "\\",
	"\n", "\r", "\x00", // CRITICAL: control characters
	"{", "}", "[", "]", // shell expansions
	"!", "#", "~", // shell features
	"\"", "'", // quotes
}

// ParseCommand safely parses a command string, rejecting shell metacharacters
// This is a critical security function - it prevents command injection attacks
func ParseCommand(cmdString string) (*Command, error) {
	// Check for dangerous characters
	for _, char := range dangerousChars {
		if strings.Contains(cmdString, char) {
			return nil, fmt.Errorf("command contains dangerous character: %s", char)
		}
	}

	// Split command and arguments
	parts := strings.Fields(cmdString)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return &Command{
		Name: parts[0],
		Args: parts[1:],
	}, nil
}
