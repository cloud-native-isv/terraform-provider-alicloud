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
