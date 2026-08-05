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

package exec

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand_RejectsShellMetacharacters(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "normal command",
			cmd:     "ls -la",
			wantErr: false,
		},
		{
			name:    "command with pipe",
			cmd:     "cat file | grep test",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with semicolon",
			cmd:     "ls; rm -rf /",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with ampersand",
			cmd:     "sleep 1 &",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with backtick",
			cmd:     "echo `date`",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with dollar sign",
			cmd:     "echo $HOME",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with newline",
			cmd:     "ls\ncat",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
		{
			name:    "command with redirect",
			cmd:     "cat < file",
			wantErr: true,
			errMsg:  "shell metacharacter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommand(context.Background(), tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
