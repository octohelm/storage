package json

import (
	"database/sql/driver"
	"fmt"

	"github.com/go-json-experiment/json"
)

func scanValue(dbValue any, value any) error {
	switch v := dbValue.(type) {
	case []byte:
		bytes := v
		if len(bytes) > 0 {
			return json.Unmarshal(bytes, value)
		}
		return nil
	case string:
		str := v
		if str == "" {
			return nil
		}
		return json.Unmarshal([]byte(str), value)
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot sql.Scan() from: %#v", value)
	}
}

func toValue(value any) (driver.Value, error) {
	if zeroCheck, ok := value.(interface {
		IsZero() bool
	}); ok {
		if zeroCheck.IsZero() {
			return "", nil
		}
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	str := string(bytes)
	if str == "null" {
		return "", nil
	}
	return str, nil
}
