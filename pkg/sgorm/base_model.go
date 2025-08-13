package sgorm

import (
	"reflect"
	"time"

	"github.com/huandu/xstrings"
	"gorm.io/gorm"

	"github.com/ishaqcherry9/depend/pkg/sgorm/dbclose"
)

type DB = gorm.DB

var ErrRecordNotFound = gorm.ErrRecordNotFound

const (
	DBDriverMysql = "mysql"
	DBDriverTidb  = "tidb"
)

type Model struct {
	ID        uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

type Model2 struct {
	ID        uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

type KV = map[string]interface{}

func GetTableName(object interface{}) string {
	tableName := ""

	typeof := reflect.TypeOf(object)
	switch typeof.Kind() {
	case reflect.Ptr:
		tableName = typeof.Elem().Name()
	case reflect.Struct:
		tableName = typeof.Name()
	default:
		return tableName
	}

	return xstrings.ToSnakeCase(tableName)
}

func CloseDB(db *gorm.DB) error {
	return dbclose.Close(db)
}
