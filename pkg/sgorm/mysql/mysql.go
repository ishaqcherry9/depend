package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/ishaqcherry9/depend/pkg/sgorm/dbclose"
	"github.com/ishaqcherry9/depend/pkg/sgorm/glog"
)

func Init(dsn string, opts ...Option) (*gorm.DB, error) {
	o := defaultOptions()
	o.apply(opts...)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(o.maxIdleConns)
	sqlDB.SetMaxOpenConns(o.maxOpenConns)
	sqlDB.SetConnMaxLifetime(o.connMaxLifetime)

	db, err := gorm.Open(mysqlDriver.New(mysqlDriver.Config{Conn: sqlDB}), gormConfig(o))
	if err != nil {
		return nil, err
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	if o.enableTrace {
		err = db.Use(otelgorm.NewPlugin())
		if err != nil {
			return nil, fmt.Errorf("using gorm opentelemetry, err: %v", err)
		}
	}

	if len(o.slavesDsn) > 0 {
		err = db.Use(rwSeparationPlugin(o))
		if err != nil {
			return nil, err
		}
	}

	for _, plugin := range o.plugins {
		err = db.Use(plugin)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func InitTidb(dsn string, opts ...Option) (*gorm.DB, error) {
	return Init(dsn, opts...)
}

func gormConfig(o *options) *gorm.Config {
	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: o.disableForeignKey,
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
	}

	if o.isLog {
		if o.gLog == nil {
			config.Logger = logger.Default.LogMode(o.logLevel)
		} else {
			config.Logger = glog.NewCustomGormLogger(o.gLog, o.requestIDKey, o.logLevel)
		}
	} else {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	if o.slowThreshold > 0 {
		config.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: o.slowThreshold,
				Colorful:      true,
				LogLevel:      logger.Warn,
			},
		)
	}

	return config
}

func rwSeparationPlugin(o *options) gorm.Plugin {
	slaves := []gorm.Dialector{}
	for _, dsn := range o.slavesDsn {
		slaves = append(slaves, mysqlDriver.New(mysqlDriver.Config{
			DSN: dsn,
		}))
	}

	masters := []gorm.Dialector{}
	for _, dsn := range o.mastersDsn {
		masters = append(masters, mysqlDriver.New(mysqlDriver.Config{
			DSN: dsn,
		}))
	}

	return dbresolver.Register(dbresolver.Config{
		Sources:  masters,
		Replicas: slaves,
		Policy:   dbresolver.RandomPolicy{},
	})
}

func Close(db *gorm.DB) error {
	return dbclose.Close(db)
}
