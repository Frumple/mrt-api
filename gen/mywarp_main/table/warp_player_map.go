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

var WarpPlayerMap = newWarpPlayerMapTable("mywarp_main", "warp_player_map", "")

type warpPlayerMapTable struct {
	mysql.Table

	// Columns
	WarpID   mysql.ColumnInteger
	PlayerID mysql.ColumnInteger

	AllColumns     mysql.ColumnList
	MutableColumns mysql.ColumnList
}

type WarpPlayerMapTable struct {
	warpPlayerMapTable

	NEW warpPlayerMapTable
}

// AS creates new WarpPlayerMapTable with assigned alias
func (a WarpPlayerMapTable) AS(alias string) *WarpPlayerMapTable {
	return newWarpPlayerMapTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new WarpPlayerMapTable with assigned schema name
func (a WarpPlayerMapTable) FromSchema(schemaName string) *WarpPlayerMapTable {
	return newWarpPlayerMapTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new WarpPlayerMapTable with assigned table prefix
func (a WarpPlayerMapTable) WithPrefix(prefix string) *WarpPlayerMapTable {
	return newWarpPlayerMapTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new WarpPlayerMapTable with assigned table suffix
func (a WarpPlayerMapTable) WithSuffix(suffix string) *WarpPlayerMapTable {
	return newWarpPlayerMapTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newWarpPlayerMapTable(schemaName, tableName, alias string) *WarpPlayerMapTable {
	return &WarpPlayerMapTable{
		warpPlayerMapTable: newWarpPlayerMapTableImpl(schemaName, tableName, alias),
		NEW:                newWarpPlayerMapTableImpl("", "new", ""),
	}
}

func newWarpPlayerMapTableImpl(schemaName, tableName, alias string) warpPlayerMapTable {
	var (
		WarpIDColumn   = mysql.IntegerColumn("warp_id")
		PlayerIDColumn = mysql.IntegerColumn("player_id")
		allColumns     = mysql.ColumnList{WarpIDColumn, PlayerIDColumn}
		mutableColumns = mysql.ColumnList{}
	)

	return warpPlayerMapTable{
		Table: mysql.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		WarpID:   WarpIDColumn,
		PlayerID: PlayerIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}