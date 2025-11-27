package trader

import (
	"fmt"
	"strconv"
)

// SafeFloat64 从map中安全提取float64值
func SafeFloat64(data map[string]interface{}, key string) (float64, error) {
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("key '%s' not found", key)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		// 尝试解析字符串为float64
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string '%s' as float64: %w", v, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("value for key '%s' is not a number (type: %T)", key, v)
	}
}

// SafeString 从map中安全提取字符串值
func SafeString(data map[string]interface{}, key string) (string, error) {
	value, ok := data[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// SafeInt 从map中安全提取int值
func SafeInt(data map[string]interface{}, key string) (int, error) {
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("key '%s' not found", key)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string '%s' as int: %w", v, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("value for key '%s' is not an integer (type: %T)", key, v)
	}
}

// SafeInt64 从map中安全提取int64值
func SafeInt64(data map[string]interface{}, key string) (int64, error) {
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("key '%s' not found", key)
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string '%s' as int64: %w", v, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("value for key '%s' is not an int64 (type: %T)", key, v)
	}
}

// SafeBool 从map中安全提取bool值
func SafeBool(data map[string]interface{}, key string) (bool, error) {
	value, ok := data[key]
	if !ok {
		return false, fmt.Errorf("key '%s' not found", key)
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("cannot parse string '%s' as bool: %w", v, err)
		}
		return parsed, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("value for key '%s' is not a boolean (type: %T)", key, v)
	}
}
