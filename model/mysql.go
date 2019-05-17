package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

type FieldSchema struct {
	FieldName    string
	FieldType    string
	FieldLen     int
	AllowNull    bool
	DefaultValue string
}

type MysqlDb struct {
	Db *sql.DB
}

/**
* Create new connection
 */
func NewMysqlDb(dsn string) (*MysqlDb, error) {
	db, err := sql.Open("mysql", dsn)
	if nil != err {
		return nil, err
	}

	err = db.Ping()
	if nil != err {
		return nil, err
	}

	return &MysqlDb{Db: db}, nil
}

/**
* close connection
 */
func (this *MysqlDb) Close() {
	if nil != this.Db {
		this.Db.Close()
		this.Db = nil
	}
}

/**
* Query data from mysql
 */
func (this *MysqlDb) Query(sql string) (*sql.Rows, error) {
	if nil == this.Db {
		return nil, fmt.Errorf("invalid database connection")
	}

	if len(sql) < 1 {
		return nil, fmt.Errorf("inavlid SQL")
	}

	rows, err := this.Db.Query(sql)
	if nil != err {
		return nil, err
	}

	return rows, nil
}

/**
* Query one row data
 */
func (this *MysqlDb) QueryRow(sql string) (*sql.Row, error) {
	if nil == this.Db {
		return nil, fmt.Errorf("invalid database connection")
	}

	if len(sql) < 1 {
		return nil, fmt.Errorf("inavlid SQL")
	}

	row := this.Db.QueryRow(sql)

	return row, nil
}

/**
* Query all tables from the databases
 */
func (this *MysqlDb) GetTableNames() []string {
	rows, err := this.Query("show tables")
	if nil != err {
		return nil
	}

	defer rows.Close()
	var table_list []string
	for rows.Next() {
		var table_name string
		err = rows.Scan(&table_name)
		if nil != err {
			return nil
		}

		table_list = append(table_list, table_name)
	}

	return table_list
}

/**
* Query the table schema (create info/sql)
 */
func (this *MysqlDb) GetTableSchema(tableName string) (string, error) {
	rows, err := this.Query(fmt.Sprintf("show create table `%s`", tableName))
	if nil != err {
		return "", err
	}

	defer rows.Close()
	var schema string = ""
	for rows.Next() {
		var name string
		err = rows.Scan(&name, &schema)
		if nil != err {
			return "", err
		}
	}

	return schema, nil
}

/*
*  Parser the fieldtype info, split field type and length
 */
func (this *MysqlDb) SplitFileType(fieldType string) (string, int, error) {
	if strings.Contains(fieldType, "(") {
		var s = strings.Split(fieldType, "(")
		var sType = s[0]
		if len(s) > 1 && strings.Count(s[1], "") > 0 {
			var s_len = strings.ReplaceAll(s[1], ")", "")
			len, err := strconv.Atoi(s_len)
			if err != nil {
				return sType, len, nil
			}
		}
		return sType, 0, nil
	} else {
		return fieldType, 0, nil
	}
}

/**
*  Get all field info from the table
 */
func (this *MysqlDb) GetColumnsSchema(tableName string) (*map[string]*FieldSchema, error) {
	var schemas = make(map[string]*FieldSchema)
	recCount, ret, err := this.SqlQuery(fmt.Sprintf("show columns from `%s`", tableName))

	if err != nil {
		return nil, err
	}
	if recCount > 0 {
		for _, row := range ret {
			var schema = new(FieldSchema)
			f_type, f_len, _ := this.SplitFileType(row["Type"])

			schema.AllowNull = row["Null"] != "NO"
			schema.FieldName = row["Field"]
			schema.FieldLen = f_len
			schema.FieldType = f_type
			schema.DefaultValue = strings.Trim(row["Default"], "'")

			schemas[schema.FieldName] = schema

		}
	}

	return &schemas, nil
}

/*
 * Execute the query
 */
func (this *MysqlDb) SqlQuery(strSql string) (int, []map[string]string, error) {
	//Execute sql
	rows, err := this.Db.Query(strSql)
	if err != nil {
		return -1, nil, err
	}

	defer rows.Close()
	//Get the columns info
	cols, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	results := make([]map[string]string, 0, 20)
	i := 0

	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			return -1, nil, err
		}

		row := make(map[string]string)
		for j, v := range values {
			key := cols[j]
			row[key] = string(v)
		}
		results = append(results, row)
		i++
	}
	return i, results, nil
}

/*
* Count the number of rows affected
 */
func (this *MysqlDb) GetCount(strSql, countField string) int {
	count, data, err := this.SqlQuery(strSql)
	if nil != err || count < 0 {
		return -1
	}

	num, _ := strconv.Atoi(data[0][countField])
	return num
}

/*
 * Execute the query
 */
func (this *MysqlDb) SqlExec(strSql string) (int, error) {
	result, err := this.Db.Exec(strSql)
	if err != nil {
		return -1, err
	}

	rowsAffectNum, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	return int(rowsAffectNum), nil
}
