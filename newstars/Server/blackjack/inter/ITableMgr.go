package inter

type ITableMgr interface {
	GetTableByID(tid int32) (ITable, bool)
	NewTable(roomID int32, users []string) ITable
}
