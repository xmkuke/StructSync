package service

import (
	"fmt"
	"struct_sync/common"
	"struct_sync/logger"
	"struct_sync/model"
	"strings"
)

type SchemaSync struct {
	DestDb *model.MysqlDb
	DbSet  *DBSet
}

/**
* Create dest db connection
 */
func NewSchemaSync(dbset *DBSet) *SchemaSync {
	// root:123456@tcp(127.0.0.1:3306)/sbsp?charset=utf8
	dest_dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&timeout=%s",
		dbset.User,
		dbset.Pswd,
		dbset.Host,
		dbset.Port,
		dbset.DbName,
		dbset.Charset,
		dbset.timeout)
	sc := &SchemaSync{DbSet: dbset}
	db, err := model.NewMysqlDb(dest_dsn)
	if nil != err {
		sc.addFatalLog("NewSchemaSync", fmt.Sprintln("Connect db failed: ", err.Error()))
		return nil
	}

	sc.DestDb = db

	return sc
}

/**
* Append log info
 */
func (sc *SchemaSync) addInfoLog(action, content string) {
	var ctx = fmt.Sprintf("%s@%s %s", sc.DbSet.DbName, sc.DbSet.Host, content)
	logger.Info(ctx)
}

/**
* Append warn log
 */
func (sc *SchemaSync) addWarnLog(action, content string) {
	var ctx = fmt.Sprintf("%s@%s %s", sc.DbSet.DbName, sc.DbSet.Host, content)
	logger.Warn(ctx)
}

/**
* Append error log
 */
func (sc *SchemaSync) addErrorLog(action, content string) {
	var ctx = fmt.Sprintf("%s@%s %s", sc.DbSet.DbName, sc.DbSet.Host, content)
	logger.Error(ctx)
}

/**
* Append fatal log
 */
func (sc *SchemaSync) addFatalLog(action, content string) {
	var ctx = fmt.Sprintf("%s@%s %s", sc.DbSet.DbName, sc.DbSet.Host, content)
	logger.Fatal(ctx)
}

/**
* Check field difference
 */
func (sc *SchemaSync) getSchemaDiff(alert *TableAlterData) string {
	ssource := alert.SchemaDiff.Source
	dsource := alert.SchemaDiff.Dest
	table := alert.Table

	var alterLines []string

	// Compare field difference, use schema, not the create sql info
	for name, dt := range ssource.FieldSchemas {
		var alertSQL = ""
		s, _ := ssource.Fields[name]
		if destDt, has := dsource.FieldSchemas[name]; has {
			if dt.FieldLen != destDt.FieldLen ||
				dt.FieldType != destDt.FieldType ||
				dt.DefaultValue != destDt.DefaultValue { // exist, but diff
				alertSQL = fmt.Sprintf("CHANGE `%s` %s", name, s)
			}
		} else { // not exist, append field to dest
			alertSQL = "ADD " + s
			fmt.Println("Souce Table: ", table, " Field:", s)
		}

		if "" != alertSQL {
			alterLines = append(alterLines, alertSQL)
			sc.addWarnLog("getSchemaDiff",
				fmt.Sprint("[COLUMN.ALTER] ", table+"."+name, ", SQL=", alertSQL))
		} else {
			sc.addInfoLog("getSchemaDiff",
				fmt.Sprint("[COLUMN.ALTER] ", table+"."+name, " Same"))
		}
	}

	// Delete fields that are not in the source db
	if globalSet.DropUnecessary {
		for name := range dsource.Fields {
			if _, has := ssource.Fields[name]; !has {
				dropSQL := fmt.Sprintf("DROP `%s`", name)
				alterLines = append(alterLines, dropSQL)
				sc.addWarnLog("getSchemaDiff",
					fmt.Sprint("[COLUMN.DROP] ", table+"."+name, ", SQL=", dropSQL))
			} else {
				sc.addInfoLog("getSchemaDiff", fmt.Sprint("[COLUMN.DROP] ", table, ".", name, " Same"))
			}
		}
	}

	// Compare index
	for indexName, idx := range ssource.IndexAll {
		var alertSQL = ""
		if dIdx, has := dsource.IndexAll[indexName]; has {
			if idx.SQL != dIdx.SQL {
				alertSQL = idx.alterAddSQL(true)
				fmt.Println("Index Check: ", table, idx.SQL, dIdx.SQL)
			}
		} else {
			alertSQL = idx.alterAddSQL(false)
		}

		if "" != alertSQL {
			alterLines = append(alterLines, alertSQL)
			sc.addWarnLog("getSchemaDiff",
				fmt.Sprint("[INDEX.ALTER] ", table+"."+indexName, ", SQL=", alertSQL))
		} else {
			sc.addInfoLog("getSchemaDiff",
				fmt.Sprint("[INDEX.ALTER] ", table+"."+indexName, " Same"))
		}
	}

	// Delete index that are not in the source db
	if globalSet.DropUnecessary {
		for indexName, dIdx := range dsource.IndexAll {
			var dropSQL string
			if _, has := ssource.IndexAll[indexName]; !has {
				dropSQL = dIdx.alterDropSQL()
			}

			if dropSQL != "" {
				alterLines = append(alterLines, dropSQL)
				sc.addWarnLog("getSchemaDiff",
					fmt.Sprint("[INDEX.DROP] ", table+"."+indexName, ", SQL=", dropSQL))
			} else {
				sc.addInfoLog("getSchemaDiff",
					fmt.Sprint("[INDEX.DROP] ", table+"."+indexName, " Same"))
			}
		}
	}

	// Compare foreign key
	for foreignName, idx := range ssource.ForeignAll {
		var alterSQL = ""
		if dIdx, has := dsource.ForeignAll[foreignName]; has {
			if idx.SQL != dIdx.SQL {
				alterSQL = idx.alterAddSQL(true)
			}
		} else {
			alterSQL = idx.alterAddSQL(false)
		}
		if alterSQL != "" {
			alterLines = append(alterLines, alterSQL)
			sc.addWarnLog("getSchemaDiff",
				fmt.Sprint("[FOREIGN_KEY.ALTER] ", table+"."+foreignName, ", SQL=", alterSQL))
		} else {
			sc.addInfoLog("getSchemaDiff",
				fmt.Sprint("[FOREIGN_KEY.ALTER] ", table+"."+foreignName, " Same"))
		}
	}

	// Delete foreign key that are not in the source db
	if globalSet.DropUnecessary {
		for foreignName, dIdx := range dsource.ForeignAll {
			var dropSQL = ""
			if _, has := ssource.ForeignAll[foreignName]; !has {
				dropSQL = dIdx.alterDropSQL()
			}

			if dropSQL != "" {
				alterLines = append(alterLines, dropSQL)
				sc.addWarnLog("getSchemaDiff",
					fmt.Sprint("[FOREIGN_KEY.DROP] ", table+"."+foreignName, ", SQL=", dropSQL))
			} else {
				sc.addInfoLog("getSchemaDiff",
					fmt.Sprint("[FOREIGN_KEY.DROP] ", table+"."+foreignName, " Same"))
			}
		}
	}

	// Compare extend info
	for name, dt := range ssource.Extend {
		var alertSQL = ""
		if destDt, has := dsource.Extend[name]; has { // exists, but diff
			if dt != destDt {
				if name == "CHARSET" {
					alertSQL += "DEFAULT "
				}
				alertSQL += name + "=" + dt
			}

		} else { // not exists, append
			if name == "CHARSET" {
				alertSQL += "DEFAULT "
			}
			alertSQL += name + "=" + dt
		}

		if "" != alertSQL {
			alterLines = append(alterLines, alertSQL)
			sc.addWarnLog("getSchemaDiff",
				fmt.Sprint("[EXTEND.ALTER] ", table+"."+name, ", SQL=", alertSQL))

		} else {
			sc.addInfoLog("getSchemaDiff",
				fmt.Sprint("[EXTEND.ALTER] ", table+"."+name, " Same"))
		}
	}

	return strings.Join(alterLines, ",\n")
}

/**
* Get Alter Database table info
 */
func (sc *SchemaSync) getAlterDataByTable(table string) *TableAlterData {
	alter := &TableAlterData{Table: table, Type: alterTypeNo}
	var srcSchema = gTableList[table].SchemaRaw
	var srcSchema2 = gTableList[table].SchemaRawNoInc
	destSchema, err := sc.DestDb.GetTableSchema(table)
	if nil != err {
		sc.addWarnLog("getAlterDataByTable", fmt.Sprint("GetTableSchema Failed!", err.Error()))
	}

	var destSchema2 = common.RemoveAutoIncrement(destSchema)
	if srcSchema2 == destSchema2 { // struct is same
		return alter
	}

	alter.SchemaDiff = newSchemaDiff(table, destSchema, gTableList[table])
	if destSchema != "" {
		fldSchema, _ := sc.DestDb.GetColumnsSchema(table)
		if fldSchema != nil {
			alter.SchemaDiff.Dest.FieldSchemas = *fldSchema
		}
	}
	if srcSchema == "" && globalSet.DropUnecessary {
		alter.Type = alterTypeDrop
		alter.SQL = fmt.Sprintf("DROP TABLE `%s`;\n", table)
		return alter
	}

	if destSchema == "" { // dest table is not exists
		alter.Type = alterTypeCreate
		alter.SQL = srcSchema + ";"
		return alter
	}

	diff := sc.getSchemaDiff(alter)
	if "" != diff {
		alter.Type = alterTypeAlter
		alter.SQL = fmt.Sprintf("ALTER TABLE `%s` %s;", table, diff)
	}

	return alter
}

/**
* Execute adjust sql to dest db
 */
func (sc *SchemaSync) SyncSQL2Dest(sql string, sqlList []string) error {
	sc.addWarnLog("SyncSQL2Dest", fmt.Sprintln("Exec Sql:\n", sql))
	sql = strings.TrimSpace(sql)
	if "" == sql {
		sc.addWarnLog("SyncSQL2Dest", "sql empty, skip")
		return nil
	}

	t := NewMyTimer()
	_, err := sc.DestDb.SqlExec(sql)
	if nil != err && nil != sqlList && len(sqlList) > 0 {
		tx, errTx := sc.DestDb.Db.Begin()
		if nil == errTx {
			for _, sql := range sqlList {
				_, err = tx.Exec(sql)
				if err != nil {
					break
				}
			}

			if err == nil {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}
	}

	t.Stop()
	if nil != err {
		sc.addErrorLog("SyncSQL2Dest", fmt.Sprintln("excute sql failed,", err.Error()))
		return err
	}

	sc.addInfoLog("SyncSQL2Dest", fmt.Sprintln("Execute sql succeed, used:", t.UsedSecond()))
	return err
}
