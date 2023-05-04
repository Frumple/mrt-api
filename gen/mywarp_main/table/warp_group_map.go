//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/mysql"
)

var WarpGroupMap = newWarpGroupMapTable("mywarp_main", "warp_group_map", "")

type warpGroupMapTable struct {
	mysql.Table

	// Columns
	WarpID  mysql.ColumnInteger
	GroupID mysql.ColumnInteger

	AllColumns     mysql.ColumnList
	MutableColumns mysql.ColumnList
}

type WarpGroupMapTable struct {
	warpGroupMapTable

	NEW warpGroupMapTable
}

// AS creates new WarpGroupMapTable with assigned alias
func (a WarpGroupMapTable) AS(alias string) *WarpGroupMapTable {
	return newWarpGroupMapTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new WarpGroupMapTable with assigned schema name
func (a WarpGroupMapTable) FromSchema(schemaName string) *WarpGroupMapTable {
	return newWarpGroupMapTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new WarpGroupMapTable with assigned table prefix
func (a WarpGroupMapTable) WithPrefix(prefix string) *WarpGroupMapTable {
	return newWarpGroupMapTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new WarpGroupMapTable with assigned table suffix
func (a WarpGroupMapTable) WithSuffix(suffix string) *WarpGroupMapTable {
	return newWarpGroupMapTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newWarpGroupMapTable(schemaName, tableName, alias string) *WarpGroupMapTable {
	return &WarpGroupMapTable{
		warpGroupMapTable: newWarpGroupMapTableImpl(schemaName, tableName, alias),
		NEW:               newWarpGroupMapTableImpl("", "new", ""),
	}
}

func newWarpGroupMapTableImpl(schemaName, tableName, alias string) warpGroupMapTable {
	var (
		WarpIDColumn   = mysql.IntegerColumn("warp_id")
		GroupIDColumn  = mysql.IntegerColumn("group_id")
		allColumns     = mysql.ColumnList{WarpIDColumn, GroupIDColumn}
		mutableColumns = mysql.ColumnList{}
	)

	return warpGroupMapTable{
		Table: mysql.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		WarpID:  WarpIDColumn,
		GroupID: GroupIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
