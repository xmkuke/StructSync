package service

import (
	"encoding/json"
	"fmt"
	"struct_sync/logger"
	"regexp"
	"strings"
)

const (
	indexTypePrimary    = "PRIMARY"
	indexTypeIndex      = "INDEX"
	indexTypeForeignKey = "FOREIGN KEY"
)

type DbIndex struct {
	IndexType      string
	Name           string
	SQL            string
	RelationTables []string
}

var indexReg = regexp.MustCompile(`^([A-Z]+\s)?KEY\s`)

var fkeyReg = regexp.MustCompile("^CONSTRAINT `(.+)` FOREIGN KEY.+ REFERENCES `(.+)` ")

/**
* Parser index
 */
func parseIndexLine(line string) *DbIndex {
	line = strings.TrimSpace(line)
	//line = strings.ReplaceAll(line, " USING BTREE", "")
	idx := &DbIndex{
		SQL:            line,
		RelationTables: []string{},
	}

	if strings.HasPrefix(line, indexTypePrimary) { // Primary key
		idx.IndexType = indexTypePrimary
		idx.Name = "PRIMARY KEY"
		return idx
	}

	//  UNIQUE KEY `idx_a` (`a`) USING HASH COMMENT 'xx',
	//  FULLTEXT KEY `c` (`c`)
	//  PRIMARY KEY (`d`)
	//  KEY `idx_e` (`e`),
	if indexReg.MatchString(line) {
		arr := strings.Split(line, "`")
		idx.IndexType = indexTypeIndex
		idx.Name = arr[1]
		return idx
	}

	//CONSTRAINT `busi_table_ibfk_1` FOREIGN KEY (`repo_id`) REFERENCES `repo_table` (`repo_id`)
	foreignMatches := fkeyReg.FindStringSubmatch(line)
	if len(foreignMatches) > 0 {
		idx.IndexType = indexTypeForeignKey
		idx.Name = foreignMatches[1]
		idx.addRelationTable(foreignMatches[2])
		return idx
	}

	logger.Fatal("index#parseIndexLine: db_index parse failed,unsupport,line:", line)
	return nil
}

/**
* Append relation table
 */
func (idx *DbIndex) addRelationTable(table string) {
	table = strings.TrimSpace(table)
	if table != "" {
		idx.RelationTables = append(idx.RelationTables, table)
	}
}

func (idx *DbIndex) alterAddSQL(drop bool) string {
	alterSQL := []string{}
	if drop {
		dropSQL := idx.alterDropSQL()
		if dropSQL != "" {
			alterSQL = append(alterSQL, dropSQL)
		}
	}

	switch idx.IndexType {
	case indexTypePrimary:
		alterSQL = append(alterSQL, "ADD "+idx.SQL)
	case indexTypeIndex, indexTypeForeignKey:
		alterSQL = append(alterSQL, fmt.Sprintf("ADD %s", idx.SQL))
	default:
		logger.Fatal("index#alterAddSQL: unknow indexType", idx.IndexType)
	}
	return strings.Join(alterSQL, ",\n")
}

func (idx *DbIndex) String() string {
	bs, _ := json.MarshalIndent(idx, "  ", " ")
	return string(bs)
}

func (idx *DbIndex) alterDropSQL() string {
	switch idx.IndexType {
	case indexTypePrimary:
		return "DROP PRIMARY KEY"
	case indexTypeIndex:
		return fmt.Sprintf("DROP INDEX `%s`", idx.Name)
	case indexTypeForeignKey:
		return fmt.Sprintf("DROP FOREIGN KEY `%s`", idx.Name)
	default:
		logger.Fatal("index#alterDropSQL: unknow indexType", idx.IndexType)
	}
	return ""
}
