package alicloud

import (
	"encoding/json"
	"fmt"
)

func NormalizeLogtailConfigJson(jsonStr string) (string, error) {
	return normalizeJsonString(jsonStr)
}

func ValidateLogtailConfigJsonObject(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}
	var i interface{}
	if err := json.Unmarshal([]byte(jsonStr), &i); err != nil {
		return err
	}
	if _, ok := i.(map[string]interface{}); !ok {
		return fmt.Errorf("expected JSON object")
	}
	return nil
}

func validateLogtailConfigJsonObjectString(v interface{}, k string) (ws []string, errors []error) {
	jsonStr, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("%s must be a JSON string", k))
		return
	}

	if err := ValidateLogtailConfigJsonObject(jsonStr); err != nil {
		errors = append(errors, fmt.Errorf("%s: %s", k, err.Error()))
	}
	return
}
