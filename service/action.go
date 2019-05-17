package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"struct_sync/logger"
	db "struct_sync/model"
	"time"
)

type InputMode int64

const (
	DbMode   InputMode = 0x1
	FileMode InputMode = 0x2
)

// Source db struct map
var gTableList map[string]*MySchema

// global config object
var globalSet *GlobalSet

// diff sql list
var gSqlList []string

// global config
type GlobalSet struct {
	SrcDbDsn       *DBSet    // source db connection info
	DestDbList     []*DBSet  // dest db list
	PageSize       int       // Page Size
	ChanNum        int       // Max chan num
	InputSql       string    // source db schema file
	OutputDir      string    // output path
	DropUnecessary bool      // delete tag
	InputMode      InputMode // input mode
	ExecuteSQL     bool      // execute mode
	SaveSQL        bool      // save sql
	TimeOut        string    // Execute SQL timeout
	LogLevel       int       // Log level
	LogPath        string    // default ${app}/log
	LogFileName    string    // log file name, ex: StructSync_20190101.log  or StructSync_${date}${time}.log
}

// db connection info
type DBSet struct {
	Host    string
	Port    string
	DbName  string
	User    string
	Pswd    string
	Charset string
	timeout string
}

// Sync result
type SyncRet struct {
	Id  string
	Ret int    // Execute result: 0 all failed 1 all success, 2 part success
}

/**
* Init Global config
 */
func InitGlobalSet(set *GlobalSet) {
	globalSet = set
	if globalSet.SaveSQL {
		globalSet.OutputDir += "/" + time.Now().Format("2006-01-02")
		_, err := os.Stat(globalSet.OutputDir)
		if nil != err {
			err = os.MkdirAll(globalSet.OutputDir, os.ModeDir|os.ModePerm)
			if nil != err {
				logger.Fatal("Create dir failed, dir =", globalSet.OutputDir, ",", err.Error())
			}
		}
	}
}

/**
* Init Src database
 */
func InitSrcDbSchema(dbSet *DBSet) {
	if dbSet.timeout == "" {
		dbSet.timeout = globalSet.TimeOut
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&timeout=%s",
		dbSet.User,
		dbSet.Pswd,
		dbSet.Host,
		dbSet.Port,
		dbSet.DbName,
		dbSet.Charset,
		dbSet.timeout)

	srcDb, err := db.NewMysqlDb(dsn)
	if nil != err {
		logger.Fatal("Init Source Database Failed!", err.Error())
		panic("Init Source Database Failed: " + err.Error())
	}

	defer srcDb.Close()

	tableNameList := srcDb.GetTableNames()
	if nil == tableNameList {
		logger.Fatal("Get Source Database Table List Failed")
		panic("Get Source Database Table List Failed")
	}

	gTableList = make(map[string]*MySchema, len(tableNameList))
	for _, tableName := range tableNameList {
		tableSchema, _ := srcDb.GetTableSchema(tableName)
		if len(tableSchema) < 1 {
			logger.Fatal("Get Source Database Table Schema Failed")
			panic("Get Source Database Table Schema Failed")
		}

		var tblSchema = ParseSchema(tableSchema)

		//Parser field info
		fldSchema, _ := srcDb.GetColumnsSchema(tableName)

		//tbl_schema.FieldSchemas = make(map[string]*db.FieldSchema)
		tblSchema.FieldSchemas = *fldSchema
		gTableList[tableName] = tblSchema
	}
}

/**
* begin check src & dest db difference
 */
func StartDatabaseSync() {
	if globalSet.InputMode == DbMode {
		InitSrcDbSchema(globalSet.SrcDbDsn)
	} else if globalSet.InputMode == FileMode {
		ParseSQLFile()
	}

	totalNum := len(globalSet.DestDbList)

	// Use chan sync database
	var syncChan chan SyncRet = make(chan SyncRet, globalSet.ChanNum) // Init chan with buffer
	for index, destDbSet := range globalSet.DestDbList {

		dbSet := destDbSet

		if globalSet.TimeOut != "" {
			dbSet.timeout = globalSet.TimeOut
		} else {
			dbSet.timeout = "600s"
		}

		if globalSet.InputMode == DbMode {
			go DiffOneDB(syncChan, dbSet, string(index))
		} else {
			go FillOneDB(syncChan, dbSet, string(index))
		}
	}

	// Wait for execute result
	for j := 0; j < totalNum; j++ {
		ret := <-syncChan
		logger.Info(ret)
	}
}

/**
* Diff One database
 */
func DiffOneDB(syncChan chan SyncRet, dbSet *DBSet, id string) {
	syncRet := SyncRet{Id: id, Ret: 0}
	schemaSync := NewSchemaSync(dbSet)
	if nil == schemaSync {
		fmt.Println(dbSet.Host, dbSet.DbName, "Database connection fail")
		syncChan <- syncRet
		return
	}
	defer schemaSync.DestDb.Close()

	fmt.Println(dbSet.Host+"#"+dbSet.DbName, "Begin Sync...")
	changedTables := make(map[string][]*TableAlterData)
	for table, _ := range gTableList {
		sd := schemaSync.getAlterDataByTable(table)
		if sd.Type != alterTypeNo {
			relationTables := sd.SchemaDiff.RelationTables()
			groupKey := "multi"
			if 0 == len(relationTables) {
				groupKey = "single_" + table
			}

			if _, has := changedTables[groupKey]; !has {
				changedTables[groupKey] = make([]*TableAlterData, 0)
			}

			changedTables[groupKey] = append(changedTables[groupKey], sd)
		} else {
			var s = fmt.Sprintf("%s@%s TABLE %s Same", dbSet.DbName, dbSet.Host, table)
			logger.Info(s)
		}
	}

	//Check Unecessary
	if globalSet.DropUnecessary {
		destTableList := schemaSync.DestDb.GetTableNames()

		for _, table := range destTableList {
			if gTableList[table] == nil {
				alter := &TableAlterData{Table: table, Type: alterTypeDrop}
				dropSQL := fmt.Sprintf("DROP TABLE `%s`", table)
				alter.SQL = dropSQL
				groupKey := "single_" + table
				changedTables[groupKey] = make([]*TableAlterData, 0)
				changedTables[groupKey] = append(changedTables[table], alter)

				schemaSync.addWarnLog("DiffOneDB",
					fmt.Sprint("[TABLE.DROP] ", table, ", SQL=", dropSQL))
			}
		}
	}

	numOk := 0
	numFailed := 0
	canRunTypePref := "single"
	var hFile *os.File

	if globalSet.SaveSQL {
		var err error
		fileName := globalSet.OutputDir + fmt.Sprintf("/%s@%s#%s.sql", dbSet.DbName, dbSet.Host, dbSet.Port)
		hFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
		if nil != err { // File open fail
			//bSave = false
			logger.Warn("Create file failed: ", fileName, ",", err.Error())
		}

		defer hFile.Close()
	}

runSync:
	for typeName, sds := range changedTables {
		if !strings.HasPrefix(typeName, canRunTypePref) {
			continue
		}

		var sqlList []string
		for _, sd := range sds {
			sql := strings.TrimRight(sd.SQL, ";")
			sqlList = append(sqlList, sql)
		}

		sql := strings.Join(sqlList, ";\n") + ";\n"
		if globalSet.ExecuteSQL { // Execute SQL
			var ret error = schemaSync.SyncSQL2Dest(sql, sqlList)
			if ret == nil {
				numOk++
			} else {
				numFailed++
			}
		}

		if globalSet.SaveSQL {
			hFile.WriteString(sql)
		}
	}

	if canRunTypePref == "single" {
		canRunTypePref = "multi"
		goto runSync
	}

	syncRet.Ret = 1
	if globalSet.ExecuteSQL {
		if numOk == 0 {
			syncRet.Ret = 0
		} else if numFailed > 0 {
			syncRet.Ret = 2
		}
		logger.Info("All sql execute done, succeed", numOk, ", failed:", numFailed)
	}

	fmt.Println(dbSet.Host+"#"+dbSet.DbName, "End Sync！")
	syncChan <- syncRet

	return
}

/**
* Parser SQL and execute
 */
func ParseSQLFile() {
	hFile, err := os.Open(globalSet.InputSql)
	if nil != err {
		logger.Fatal("Parse SQL File Failed", err.Error())
		panic("Parse SQL File Failed: " + err.Error())
	}

	defer hFile.Close()
	var sqlList []string
	var content string
	buf := bufio.NewReader(hFile)
	for {
		line, err := buf.ReadString('\n')
		if nil != err {
			if err == io.EOF {
				break
			}

			logger.Fatal("Parse SQL File Failed:", err.Error())
			panic("Parse SQL File Failed: " + err.Error())
		}

		line = strings.TrimSpace(line)
		if "" == line { // Empty Line
			continue
		}

		if strings.HasPrefix(line, "-- ") { // Comment Line
			continue
		}

		content += line
		lastChar := line[len(line)-1:]
		if lastChar == ";" {
			sqlList = append(sqlList, content)
			content = ""
		}
	}

	if content != "" { // Maybe line miss ';'
		sqlList = append(sqlList, content+";")
	}

	if nil == sqlList {
		logger.Fatal("SQL File Empty")
		panic("SQL File Empty")
	}

	gSqlList = sqlList
}

/**
* Fill database with file input
 */
func FillOneDB(syncChan chan SyncRet, dbSet *DBSet, id string) {
	syncRet := SyncRet{Id: id, Ret: 0}
	schemaSync := NewSchemaSync(dbSet)
	if nil == schemaSync { //  连接数据库失败
		fmt.Println(dbSet.Host, dbSet.DbName, "Database connection fail")
		syncChan <- syncRet
		return
	}
	defer schemaSync.DestDb.Close()

	fmt.Println(dbSet.Host+"#"+dbSet.DbName, "Begin Sync...")

	// Execute SQL
	numOk := 0
	numFailed := 0
	syncRet.Ret = 1
	if globalSet.ExecuteSQL {
		var ret error = schemaSync.SyncSQL2Dest(strings.Join(gSqlList, ""), gSqlList)
		if ret != nil {
			syncRet.Ret = 0
			numFailed++
		} else {
			numOk++
		}

		logger.Info("all sql execute done, ", numOk, ", failed:", numFailed)
	}

	fmt.Println(dbSet.Host+"#"+dbSet.DbName, "End Sync！")
	syncChan <- syncRet

	return
}
