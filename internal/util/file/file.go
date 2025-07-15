package file

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadJSONFile(filename string, target interface{}) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from file %s: %w", filename, err)
	}
	return nil
}

func WriteJSONFile(filename string, source interface{}) error {
	data, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file %s: %w", filename, err)
	}
	return nil
}
