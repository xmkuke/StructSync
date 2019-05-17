// Syntax parser
package service

import (
	"fmt"
	"struct_sync/common"
	"struct_sync/model"
	"strings"
)

// MySchema table schema
type MySchema struct {
	SchemaRaw      string                        // Table DDL
	SchemaRawNoInc string                        // Table DDL， but not include Auto_Increment info
	Fields         map[string]string             // field
	FieldSchemas   map[string]*model.FieldSchema // field struct
	IndexAll       map[string]*DbIndex           // index
	ForeignAll     map[string]*DbIndex           // foreign key
	Extend         map[string]string             // extend info
}

func (mys *MySchema) String() string {
	s := "Fields:\n"
	fl := maxMapKeyLen(mys.Fields, 2)
	for name, v := range mys.Fields {
		s += fmt.Sprintf("  %"+fl+"s : %s\n", name, v)
	}

	s += "Index:\n"
	fl = maxMapKeyLen(mys.IndexAll, 2)
	for name, idx := range mys.IndexAll {
		s += fmt.Sprintf("  %"+fl+"s : %s\n", name, idx.SQL)
	}
	s += "ForeignKey:\n"
	fl = maxMapKeyLen(mys.ForeignAll, 2)
	for name, idx := range mys.ForeignAll {
		s += fmt.Sprintf("  %"+fl+"s : %s\n", name, idx.SQL)
	}

	s += "Extend:\n"
	fl = maxMapKeyLen(mys.Extend, 2)
	for name, v := range mys.Extend {
		s += fmt.Sprintf("  %"+fl+"s : %s\n", name, v)
	}

	return s
}

/**
* Get field list
 */
func (mys *MySchema) GetFieldNames() []string {
	var names []string
	for name := range mys.Fields {
		names = append(names, name)
	}
	return names
}

/**
* Get relation table
 */
func (mys *MySchema) RelationTables() []string {
	tbs := make(map[string]int)
	for _, idx := range mys.ForeignAll {
		for _, tb := range idx.RelationTables {
			tbs[tb] = 1
		}
	}
	var tables []string
	for tb := range tbs {
		tables = append(tables, tb)
	}
	return tables
}

/**
* Parser extend info
 */
func parseSchemaExtend(line string) map[string]string {
	line = strings.TrimLeft(line, ")")
	line = strings.TrimSpace(line)
	line = strings.Replace(line, "DEFAULT ", "", -1)
	var keys []string
	var buff = ""
	var tag = false
	for _, v := range line {
		ch := fmt.Sprintf("%c", v)
		if tag { // Tag start
			if ch == "'" { // Tab end
				buff += "'"
				tag = false
			} else {
				buff += ch
			}
		} else { // Tag end
			if ch == "'" { // Tag start
				buff += "'"
				tag = true
			} else if ch == " " { // ignore space
				keys = append(keys, buff)
				buff = "" // clean buff
			} else {
				buff += ch
			}
		}
	}

	if len(buff) > 0 {
		keys = append(keys, buff)
		buff = ""
	}

	// Parser key
	extendList := make(map[string]string)
	for _, v := range keys {
		if len(v) < 1 {
			continue
		}

		index := strings.Index(v, "=")
		name := strings.TrimSpace(v[:index])
		value := strings.TrimSpace(v[index+1:])
		if "AUTO_INCREMENT" != name { // ignore auto inc
			extendList[name] = value
		}
	}

	return extendList
}

/**
* Parser syntax
 */
func ParseSchema(schema string) *MySchema {
	if len(schema) < 1 {
		return nil
	}

	schema = strings.TrimSpace(schema)
	schema2 := common.RemoveAutoIncrement(schema)
	lines := strings.Split(schema, "\n")
	mys := &MySchema{
		SchemaRaw:      schema,
		SchemaRawNoInc: schema2,
		Fields:         make(map[string]string),
		IndexAll:       make(map[string]*DbIndex, 0),
		ForeignAll:     make(map[string]*DbIndex, 0),
		Extend:         make(map[string]string),
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if "" == line { // Empty line
			continue
		}

		if strings.HasPrefix(line, "--") { // Comment
			continue
		}

		if strings.HasPrefix(line, "CREATE TABLE") { // Field info start
			continue
		}

		if strings.HasPrefix(line, ")") { // start extend info
			mys.Extend = parseSchemaExtend(line)
			continue
		}

		line = strings.TrimRight(line, ",")
		if '`' == line[0] {
			index := strings.Index(line[1:], "`") // get field name
			field_name := line[1 : index+1]
			mys.Fields[field_name] = line
		} else {
			idx := parseIndexLine(line) // parser index
			if idx == nil {
				continue
			}
			switch idx.IndexType {
			case indexTypeForeignKey:
				mys.ForeignAll[idx.Name] = idx
			default:
				mys.IndexAll[idx.Name] = idx
			}
		}
	}

	return mys
}
