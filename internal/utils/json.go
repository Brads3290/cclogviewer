package utils

import (
	"encoding/json"
	"fmt"
)

// UnmarshalToMap unmarshals JSON data into a map[string]interface{}
func UnmarshalToMap(data json.RawMessage) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON: %w", err)
	}
	return result, nil
}

// UnmarshalToSlice unmarshals JSON data into a []interface{}
func UnmarshalToSlice(data json.RawMessage) ([]interface{}, error) {
	var result []interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON: %w", err)
	}
	return result, nil
}