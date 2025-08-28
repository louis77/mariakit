package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringArray []string

func (p StringArray) Value() (driver.Value, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (p *StringArray) Scan(value any) error {
	var data []byte

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", value)
	}

	err := json.Unmarshal(data, p)
	if err != nil {
		return err
	}

	return err
}
