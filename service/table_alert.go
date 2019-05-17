// 构建修改语句
package service

import (
	"fmt"
	"strings"
)

type AlterType int

const (
	alterTypeNo     AlterType = 0
	alterTypeCreate           = 1
	alterTypeDrop             = 2
	alterTypeAlter            = 3
)

// 获取修改类型
func (at AlterType) String() string {
	switch at {
	case alterTypeNo:
		return "not_change"
	case alterTypeCreate:
		return "create"
	case alterTypeDrop:
		return "drop"
	case alterTypeAlter:
		return "alter"
	default:
		return "unknow"
	}

}

type TableAlterData struct {
	Table      string
	Type       AlterType
	SQL        string
	SchemaDiff *SchemaDiff
}

func (ta *TableAlterData) String() string {
	relationTables := ta.SchemaDiff.RelationTables()
	fmtStr := `
-- Table : %s
-- Type  : %s
-- RealtionTables : %s
-- SQL   :
%s
`
	return fmt.Sprintf(fmtStr, ta.Table, ta.Type, strings.Join(relationTables, ","), ta.SQL)
}
