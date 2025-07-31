package utils

import "fmt"

// ValidateRequiredFields checks that all required fields are present in the data map
func ValidateRequiredFields(data map[string]interface{}, fields ...string) error {
	var missing []string

	for _, field := range fields {
		if ExtractString(data, field) == "" {
			missing = append(missing, field)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %v", missing)
	}

	return nil
}

// ValidateRequiredField checks that a single required field is present
func ValidateRequiredField(data map[string]interface{}, field string) error {
	if ExtractString(data, field) == "" {
		return fmt.Errorf("missing required field: %s", field)
	}
	return nil
}