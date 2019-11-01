package types

import (
	"encoding/json"
	"errors"
	"time"
)

type MarshalableDuration time.Duration

func (m MarshalableDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(m).String())
}

func (m *MarshalableDuration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		d, err := time.ParseDuration(value)
		if err != nil {
			panic(err)
		}
		*m = MarshalableDuration(d)
	case float64:
		*m = MarshalableDuration(time.Duration(value))
	default:
		return errors.New("duration must be a string or integer")
	}

	return nil
}
