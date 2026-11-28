package repositories

import (
	"database/sql/driver"
	"strings"
)

type postgresTextArray []string

func (values postgresTextArray) Value() (driver.Value, error) {
	var result strings.Builder
	result.WriteByte('{')
	for i, value := range values {
		if i > 0 {
			result.WriteByte(',')
		}
		result.WriteByte('"')
		for _, character := range value {
			if character == '\\' || character == '"' {
				result.WriteByte('\\')
			}
			result.WriteRune(character)
		}
		result.WriteByte('"')
	}
	result.WriteByte('}')
	return result.String(), nil
}
