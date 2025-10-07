package errcode

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	//101001–101099
	RedisErrNil            = NewError(101001, "redis: key not found")
	RedisErrClosed         = NewError(101002, "redis: client is closed")
	RedisErrPoolExhausted  = NewError(101003, "redis: connection pool exhausted")
	RedisErrPoolTimeout    = NewError(101004, "redis: connection pool timeout")
	RedisErrContextTimeout = NewError(101005, "redis: operation timeout")
	RedisErrExtra          = NewError(101099, "")

	//102001–102099
	DBErrNotFound                = NewError(102001, "db: record not found")
	DBErrDuplicateEntry          = NewError(102002, "db: duplicate key conflict")
	DBErrDeadlock                = NewError(102003, "db: deadlock detected")
	DBErrLockWaitTimeout         = NewError(102004, "db: lock wait timeout exceeded")
	DBErrConnRefused             = NewError(102005, "db: connection refused")
	DBErrConnLost                = NewError(102006, "db: connection lost")
	DBErrSyntax                  = NewError(102007, "db: sql syntax error")
	DBErrAccessDenied            = NewError(102008, "db: access denied")
	DBErrInvalidTransaction      = NewError(102009, "db: invalid transaction")
	DBErrNotImplemented          = NewError(102010, "db: feature not implemented")
	DBErrMissingWhereClause      = NewError(102011, "db: missing WHERE clause")
	DBErrUnsupportedRelation     = NewError(102012, "db: unsupported relation")
	DBErrPrimaryKeyRequired      = NewError(102013, "db: primary key required")
	DBErrModelValueRequired      = NewError(102014, "db: model value required")
	DBErrModelAccessibleRequired = NewError(102015, "db: model accessible fields required")
	DBErrSubQueryRequired        = NewError(102016, "db: subquery required")
	DBErrInvalidData             = NewError(102017, "db: invalid data")
	DBErrUnsupportedDriver       = NewError(102018, "db: unsupported driver")
	DBErrRegistered              = NewError(102019, "db: already registered")
	DBErrInvalidField            = NewError(102020, "db: invalid field")
	DBErrEmptySlice              = NewError(102021, "db: empty slice found")
	DBErrDryRunModeUnsupported   = NewError(102022, "db: dry run mode unsupported")
	DBErrInvalidDB               = NewError(102023, "db: invalid db")
	DBErrInvalidValue            = NewError(102024, "db: invalid value (should be pointer to struct or slice)")
	DBErrInvalidValueOfLength    = NewError(102025, "db: invalid association values, length doesn't match")
	DBErrPreloadNotAllowed       = NewError(102026, "db: preload not allowed when count is used")
	DBErrForeignKeyViolated      = NewError(102027, "db: foreign key constraint violated")
	DBErrCheckConstraintViolated = NewError(102028, "db: check constraint violated")
	DBErrContextTimeout          = NewError(102029, "db: operation timeout")
	DBErrExtra                   = NewError(102099, "")

	//todo pulsar ...

	//108001–108099
	YunDunParamError  = NewError(108001, "yundun: param error")
	YunDunRequestErr  = NewError(108002, "yundun: request error")
	YunDunResponseErr = NewError(108003, "yundun: response error")
	YunDunErrExtra    = NewError(108099, "")
)

func WrapRedisErr(err error) *Error {
	if err == nil {
		return nil
	}

	//redis自身 & context
	switch {
	case errors.Is(err, redis.Nil):
		return RedisErrNil
	case errors.Is(err, redis.ErrClosed):
		return RedisErrClosed
	case errors.Is(err, redis.ErrPoolExhausted):
		return RedisErrPoolExhausted
	case errors.Is(err, redis.ErrPoolTimeout):
		return RedisErrPoolTimeout

	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return RedisErrContextTimeout
	}

	//防并发覆盖，返回新Error
	return RedisErrExtra.WithMsg(err.Error())
}

func WrapDBErr(err error) *Error {
	if err == nil {
		return nil
	}

	//gorm自身 & context
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return DBErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return DBErrDuplicateEntry
	case errors.Is(err, gorm.ErrInvalidTransaction):
		return DBErrInvalidTransaction
	case errors.Is(err, gorm.ErrNotImplemented):
		return DBErrNotImplemented
	case errors.Is(err, gorm.ErrMissingWhereClause):
		return DBErrMissingWhereClause
	case errors.Is(err, gorm.ErrUnsupportedRelation):
		return DBErrUnsupportedRelation
	case errors.Is(err, gorm.ErrPrimaryKeyRequired):
		return DBErrPrimaryKeyRequired
	case errors.Is(err, gorm.ErrModelValueRequired):
		return DBErrModelValueRequired
	case errors.Is(err, gorm.ErrModelAccessibleFieldsRequired):
		return DBErrModelAccessibleRequired
	case errors.Is(err, gorm.ErrSubQueryRequired):
		return DBErrSubQueryRequired
	case errors.Is(err, gorm.ErrInvalidData):
		return DBErrInvalidData
	case errors.Is(err, gorm.ErrUnsupportedDriver):
		return DBErrUnsupportedDriver
	case errors.Is(err, gorm.ErrRegistered):
		return DBErrRegistered
	case errors.Is(err, gorm.ErrInvalidField):
		return DBErrInvalidField
	case errors.Is(err, gorm.ErrEmptySlice):
		return DBErrEmptySlice
	case errors.Is(err, gorm.ErrDryRunModeUnsupported):
		return DBErrDryRunModeUnsupported
	case errors.Is(err, gorm.ErrInvalidDB):
		return DBErrInvalidDB
	case errors.Is(err, gorm.ErrInvalidValue):
		return DBErrInvalidValue
	case errors.Is(err, gorm.ErrInvalidValueOfLength):
		return DBErrInvalidValueOfLength
	case errors.Is(err, gorm.ErrPreloadNotAllowed):
		return DBErrPreloadNotAllowed
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return DBErrForeignKeyViolated
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return DBErrCheckConstraintViolated

	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return DBErrContextTimeout
	}

	//mysql自身
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return DBErrDuplicateEntry
		case 1213:
			return DBErrDeadlock
		case 1205:
			return DBErrLockWaitTimeout
		case 1045:
			return DBErrAccessDenied
		case 2002:
			return DBErrConnRefused
		case 2013:
			return DBErrConnLost
		case 1064: // ER_PARSE_ERROR
			return DBErrSyntax
		}
	}

	//防并发覆盖，返回新Error
	return DBErrExtra.WithMsg(err.Error())
}
