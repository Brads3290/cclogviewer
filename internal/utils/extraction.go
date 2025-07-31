package utils

// ExtractString safely extracts a string value from a map
func ExtractString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ExtractBool safely extracts a boolean value from a map
func ExtractBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ExtractInt safely extracts an integer value from a map
func ExtractInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

// ExtractFloat64 safely extracts a float64 value from a map
func ExtractFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return 0
}

// ExtractMap safely extracts a map value from a map
func ExtractMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// ExtractSlice safely extracts a slice value from a map
func ExtractSlice(data map[string]interface{}, key string) []interface{} {
	if val, ok := data[key]; ok {
		if s, ok := val.([]interface{}); ok {
			return s
		}
	}
	return nil
}