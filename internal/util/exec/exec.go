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
var dangerousChars = []string{"|", "&", ";", "$", "(", ")", "<", ">", "`", "\\"}

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
