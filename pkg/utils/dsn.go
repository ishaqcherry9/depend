package utils

import (
	"strings"
)

func AdaptiveMysqlDsn(dsn string) string {
	return strings.ReplaceAll(dsn, "mysql://", "")
}

func DeleteBrackets(str string) string {
	start := strings.Index(str, "@(")
	end := strings.LastIndex(str, ")/")

	if start == -1 || end == -1 {
		return str
	}

	addr := str[start+2 : end]
	return strings.Replace(str, "@("+addr+")/", "@"+addr+"/", 1)
}
