package query

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Eq               = "eq"
	Neq              = "neq"
	Gt               = "gt"
	Gte              = "gte"
	Lt               = "lt"
	Lte              = "lte"
	Like             = "like"
	In               = "in"
	NotIN            = "notin"
	IsNull           = "isnull"
	IsNotNull        = "isnotnull"
	AND       string = "and"
	OR        string = "or"
)

var expMap = map[string]string{
	Eq:        " = ",
	Neq:       " <> ",
	Gt:        " > ",
	Gte:       " >= ",
	Lt:        " < ",
	Lte:       " <= ",
	Like:      " LIKE ",
	In:        " IN ",
	NotIN:     " NOT IN ",
	IsNull:    " IS NULL ",
	IsNotNull: " IS NOT NULL ",

	"=":           " = ",
	"!=":          " <> ",
	">":           " > ",
	">=":          " >= ",
	"<":           " < ",
	"<=":          " <= ",
	"not in":      " NOT IN ",
	"is null":     " IS NULL ",
	"is not null": " IS NOT NULL ",
}

var logicMap = map[string]string{
	AND: " AND ",
	OR:  " OR ",

	"&":   " AND ",
	"&&":  " AND ",
	"|":   " OR ",
	"||":  " OR ",
	"AND": " AND ",
	"OR":  " OR ",

	"and:(": " AND ",
	"and:)": " AND ",
	"or:(":  " OR ",
	"or:)":  " OR ",
}

type rulerOptions struct {
	whitelistNames map[string]bool
	validateFn     func(columns []Column) error
}

type RulerOption func(*rulerOptions)

func (o *rulerOptions) apply(opts ...RulerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithWhitelistNames(whitelistNames map[string]bool) RulerOption {
	return func(o *rulerOptions) {
		o.whitelistNames = whitelistNames
	}
}

func WithValidateFn(fn func(columns []Column) error) RulerOption {
	return func(o *rulerOptions) {
		o.validateFn = fn
	}
}

type Params struct {
	Page  int    `json:"page" form:"page" binding:"gte=0"`
	Limit int    `json:"limit" form:"limit" binding:"gte=1"`
	Sort  string `json:"sort,omitempty" form:"sort" binding:""`

	Columns []Column `json:"columns,omitempty" form:"columns"`
}

type Column struct {
	Name  string      `json:"name" form:"name"`
	Exp   string      `json:"exp" form:"exp"`
	Value interface{} `json:"value" form:"value"`
	Logic string      `json:"logic" form:"logic"`
}

func (c *Column) checkExp() (string, error) {
	symbol := "?"
	if c.Exp == "" {
		c.Exp = Eq
	}
	if v, ok := expMap[strings.ToLower(c.Exp)]; ok {
		c.Exp = v
		switch c.Exp {
		case " LIKE ":
			val, ok1 := c.Value.(string)
			if !ok1 {
				return symbol, fmt.Errorf("invalid value type '%s'", c.Value)
			}
			l := len(val)
			if l > 2 {
				val2 := val[1 : l-1]
				val2 = strings.ReplaceAll(val2, "%", "\\%")
				val2 = strings.ReplaceAll(val2, "_", "\\_")
				val = string(val[0]) + val2 + string(val[l-1])
			}
			if strings.HasPrefix(val, "%") ||
				strings.HasPrefix(val, "_") ||
				strings.HasSuffix(val, "%") ||
				strings.HasSuffix(val, "_") {
				c.Value = val
			} else {
				c.Value = "%" + val + "%"
			}
		case " IN ", " NOT IN ":
			val, ok1 := c.Value.(string)
			if ok1 {
				values := []interface{}{}
				ss := strings.Split(val, ",")
				for _, s := range ss {
					s = strings.TrimSpace(s)
					if strings.HasPrefix(s, "\"") {
						values = append(values, strings.Trim(s, "\""))
						continue
					} else if strings.HasPrefix(s, "'") {
						values = append(values, strings.Trim(s, "'"))
						continue
					}
					value, err := strconv.Atoi(s)
					if err == nil {
						values = append(values, value)
					} else {
						values = append(values, s)
					}
				}
				c.Value = values
			}
			symbol = "(?)"
		case " IS NULL ", " IS NOT NULL ":
			c.Value = nil
			symbol = ""
		}
	} else {
		return symbol, fmt.Errorf("unsported exp type '%s'", c.Exp)
	}

	if c.Logic == "" {
		c.Logic = AND
	} else {
		logic := strings.ToLower(c.Logic)
		if _, ok := logicMap[logic]; ok {
			c.Logic = logic
		} else {
			return symbol, fmt.Errorf("unsported logic type '%s'", c.Logic)
		}
	}

	return symbol, nil
}

func (p *Params) ConvertToPage() (order string, limit int, offset int) {
	page := NewPage(p.Page, p.Limit, p.Sort)
	order = page.sort
	limit = page.limit
	offset = page.page * page.limit
	return
}

func (p *Params) ConvertToGormConditions(opts ...RulerOption) (string, []interface{}, error) {
	str := ""
	args := []interface{}{}
	l := len(p.Columns)
	if l == 0 {
		return "", nil, nil
	}

	isUseIN := true
	if l == 1 {
		isUseIN = false
	}
	field := p.Columns[0].Name

	o := rulerOptions{}
	o.apply(opts...)
	if o.validateFn != nil {
		err := o.validateFn(p.Columns)
		if err != nil {
			return "", nil, err
		}
	}

	for i, column := range p.Columns {
		if column.Name == "" || (o.whitelistNames != nil && !o.whitelistNames[column.Name]) {
			return "", nil, fmt.Errorf("field name '%s' is not allowed", column.Name)
		}

		if column.Value == nil {
			v := expMap[strings.ToLower(column.Exp)]
			if v != " IS NULL " && v != " IS NOT NULL " {
				return "", nil, fmt.Errorf("field 'value' cannot be nil")
			}
		} else {
			column.Value = convertValue(column.Value)
		}

		symbol, err := column.checkExp()
		if err != nil {
			return "", nil, err
		}

		if i == l-1 {
			switch column.Logic {
			case "or:)", "and:)":
				str += column.Name + column.Exp + symbol + " ) "
			default:
				str += column.Name + column.Exp + symbol
			}
		} else {
			switch column.Logic {
			case "or:(", "and:(":
				str += " ( " + column.Name + column.Exp + symbol + logicMap[column.Logic]
			case "or:)", "and:)":
				str += column.Name + column.Exp + symbol + " ) " + logicMap[column.Logic]
			default:
				str += column.Name + column.Exp + symbol + logicMap[column.Logic]
			}
		}
		if column.Value != nil {
			args = append(args, column.Value)
		}

		if isUseIN {
			if field != column.Name {
				isUseIN = false
				continue
			}
			if column.Exp != expMap[Eq] {
				isUseIN = false
			}
		}
	}

	if isUseIN {
		str = field + " IN (?)"
		args = []interface{}{args}
	}

	return str, args, nil
}

func convertValue(v interface{}) interface{} {
	s, ok := v.(string)
	if !ok {
		return v
	}

	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "\"") {
		s2 := strings.Trim(s, "\"")
		if _, err := strconv.Atoi(s2); err == nil {
			return s2
		}
		return s
	}
	intVal, err := strconv.Atoi(s)
	if err == nil {
		return intVal
	}
	boolVal, err := strconv.ParseBool(s)
	if err == nil {
		return boolVal
	}
	floatVal, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return floatVal
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05.999999999Z0700",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05.999999999 -07:00",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return v
}

type Conditions struct {
	Columns []Column `json:"columns" form:"columns" binding:"min=1"`
}

func (c *Conditions) ConvertToGorm(opts ...RulerOption) (string, []interface{}, error) {
	p := &Params{Columns: c.Columns}
	return p.ConvertToGormConditions(opts...)
}

func (c *Conditions) CheckValid() error {
	if len(c.Columns) == 0 {
		return fmt.Errorf("field 'columns' cannot be empty")
	}

	return nil
}
