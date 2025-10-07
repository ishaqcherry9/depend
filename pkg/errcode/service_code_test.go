package errcode

import (
	"context"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/sgorm/mysql"
	"github.com/ishaqcherry9/depend/pkg/utils"
	"testing"
	"time"
)

// go test -v -test.run TestRedisErrCode
func TestRedisErrCode(t *testing.T) {
	dsnStr := "127.0.0.1:6379"

	redisC, err := goredis.Init(dsnStr, []goredis.Option{}...)
	if err != nil {
		t.Fatal(err)
	}
	if redisC == nil {
		t.Fatal("redisC is nil")
	}

	val, err := redisC.Get(context.Background(), "testkey").Result()
	if err != nil {
		errInfo := WrapRedisErr(err)
		t.Fatalf("code:%d, msg:%s", errInfo.code, errInfo.msg)
	}

	t.Logf("val:%s", val)
}

func TestDBErrCode(t *testing.T) {
	dsnStr := "root:root@tcp(127.0.0.1:3306)/test?parseTime=true&loc=Local&charset=utf8,utf8mb4"
	dsn := utils.AdaptiveMysqlDsn(dsnStr)

	opts := []mysql.Option{}
	db, err := mysql.Init(dsn, opts...)
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Fatal("db is nil")
	}

	err = db.Model(&CgMember{}).
		Select("Memberid", "Member", "CreatedAt", "UpdatedAt").
		Create(&CgMember{
			Memberid:  123456789,
			Member:    "robert",
			CreatedAt: time.Now().Unix(),
			UpdatedAt: 0,
		}).Error

	if err != nil {
		errInfo := WrapDBErr(err)
		t.Fatalf("code:%d, msg:%s", errInfo.code, errInfo.msg)
	}

	t.Log("create success")
}

type CgMember struct {
	Memberid  int64  `gorm:"column:memberid;primaryKey;comment:用户ID" json:"memberid"` // 用户ID
	Member    string `gorm:"column:member;not null;comment:玩家名" json:"member"`        // 玩家名
	CreatedAt int64  `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt int64  `gorm:"column:updated_at;not null" json:"updatedAt"`
}
