package sgorm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type Bool bool
type BitBool = Bool

func (b *Bool) Scan(value interface{}) error {
	if value == nil {
		*b = false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*b = len(v) == 1 && v[0] == 1
	case bool:
		*b = Bool(v)
	default:
		return fmt.Errorf("unsupported type: %T for Bool", value)
	}
	return nil
}

func (b Bool) Value() (driver.Value, error) {
	switch currentDriver {
	case "postgresql", "postgres":
		return bool(b), nil
	default:
		if b {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	}
}

var currentDriver string

func SetDriver(driverName string) {
	currentDriver = driverName
}

type TinyBool bool

func (b *TinyBool) Scan(value interface{}) error {
	var nb sql.NullBool
	if err := nb.Scan(value); err != nil {
		return fmt.Errorf("failed to scan TinyBool: %w", err)
	}
	*b = TinyBool(nb.Bool)
	return nil
}

func (b TinyBool) Value() (driver.Value, error) {
	if b {
		return int64(1), nil
	}
	return int64(0), nil
}
