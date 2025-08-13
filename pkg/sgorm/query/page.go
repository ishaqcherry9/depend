package query

import "strings"

var defaultMaxSize = 1000

func SetMaxSize(maxValue int) {
	if maxValue < 10 {
		maxValue = 10
	}
	defaultMaxSize = maxValue
}

type Page struct {
	page  int
	limit int
	sort  string
}

func (p *Page) Page() int {
	return p.page
}

func (p *Page) Limit() int {
	return p.limit
}

func (p *Page) Size() int {
	return p.limit
}

func (p *Page) Sort() string {
	return p.sort
}

func (p *Page) Offset() int {
	return p.page * p.limit
}

func DefaultPage(page int) *Page {
	if page < 0 {
		page = 0
	}
	return &Page{
		page:  page,
		limit: 20,
		sort:  "id DESC",
	}
}

func NewPage(page int, limit int, columnNames string) *Page {
	if page < 0 {
		page = 0
	}
	if limit > defaultMaxSize || limit < 1 {
		limit = defaultMaxSize
	}

	return &Page{
		page:  page,
		limit: limit,
		sort:  getSort(columnNames),
	}
}

func getSort(columnNames string) string {
	columnNames = strings.Replace(columnNames, " ", "", -1)
	if columnNames == "" {
		return "id DESC"
	}

	names := strings.Split(columnNames, ",")
	strs := make([]string, 0, len(names))
	for _, name := range names {
		if name[0] == '-' && len(name) > 1 {
			strs = append(strs, name[1:]+" DESC")
		} else {
			strs = append(strs, name+" ASC")
		}
	}

	return strings.Join(strs, ", ")
}
