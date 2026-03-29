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
			name:    "pipe character",
			cmd:     "ls | rm -rf /",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "command substitution",
			cmd:     "ls $(rm -rf /)",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "semicolon",
			cmd:     "ls; rm -rf /",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "backtick",
			cmd:     "ls `rm -rf /`",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "ampersand",
			cmd:     "ls & rm",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "dollar sign",
			cmd:     "ls $HOME",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "parentheses",
			cmd:     "ls (rm)",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "redirect",
			cmd:     "ls > file",
			wantErr: true,
			errMsg:  "dangerous character",
		},
		{
			name:    "backslash",
			cmd:     "ls \\",
			wantErr: true,
			errMsg:  "dangerous character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.cmd)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, cmd)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cmd)
				assert.NotEmpty(t, cmd.Name)
			}
		})
	}
}

func TestSafeExecutor_Execute_RejectsUnauthorizedCommands(t *testing.T) {
	allowedCmds := []string{"ls", "tar"}
	executor := NewSafeExecutor(allowedCmds)

	tests := []struct {
		name        string
		cmd         string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "allowed command with args",
			cmd:     "ls",
			args:    []string{"-la", "/tmp"},
			wantErr: false,
		},
		{
			name:        "command not in whitelist",
			cmd:         "sh",
			args:        []string{"-c", "echo test"},
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "rm not allowed",
			cmd:         "rm",
			args:        []string{"-rf", "/"},
			wantErr:     true,
			errContains: "not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Execute(context.Background(), tt.cmd, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSafeExecutor_Execute_RejectsDangerousArguments(t *testing.T) {
	allowedCmds := []string{"ls", "tar"}
	executor := NewSafeExecutor(allowedCmds)

	tests := []struct {
		name        string
		cmd         string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "argument with pipe",
			cmd:         "ls",
			args:        []string{"|", "rm", "-rf", "/"},
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "argument with command substitution",
			cmd:         "ls",
			args:        []string{"$(rm -rf /)"},
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "argument with semicolon",
			cmd:         "tar",
			args:        []string{"-xzvf", "file.tar.gz;", "rm", "-rf", "/"},
			wantErr:     true,
			errContains: "dangerous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Execute(context.Background(), tt.cmd, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
