package basic

import (
	"encoding/json"
	"strings"
)

// LowerUpperString demonstrates field-level custom marshaling logic.
type LowerUpperString string

func (v LowerUpperString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToUpper(string(v)))
}

func (v *LowerUpperString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*v = LowerUpperString(strings.ToLower(s))
	return nil
}

type User struct {
	ID    int              `json:"id"`
	Name  string           `json:"name"`
	Email string           `json:"email"`
	Nick  LowerUpperString `json:"nick,omitempty"`
}
