package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON[T any] struct {
	Data  T
	Valid bool
}

func (p JSON[T]) Value() (driver.Value, error) {
	if !p.Valid {
		return nil, nil
	}

	data, err := json.Marshal(p.Data)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (p *JSON[T]) Scan(value any) error {
	var data []byte

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported type for JSON: %T", value)
	}

	err := json.Unmarshal(data, &p.Data)
	if err != nil {
		return err
	}

	p.Valid = true

	return err
}
