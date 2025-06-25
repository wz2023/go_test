package logic

import (
	"errors"
	"newstars/Server/blackjack/inter"
	"newstars/framework/game_center"
	"sync"
)

// 桌子管理器
type TableMgr struct {
	mu      sync.RWMutex
	tables  map[int32]inter.ITable // key:tableID value:*Table
	nextID  int32                  // 下一个可用桌子ID
	roomMap map[int32][]int32      // roomID -> [tid1, tid2] 每个房间中的桌子列表
}

// NewTableMgr 创建桌子管理系统
func NewTableMgr(startID int32) *TableMgr {
	return &TableMgr{
		tables:  make(map[int32]inter.ITable),
		nextID:  startID,
		roomMap: make(map[int32][]int32),
	}
}

// GetTableByID 获取桌子对象
func (mgr *TableMgr) GetTableByID(tid int32) (inter.ITable, bool) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	table, exists := mgr.tables[tid]
	return table, exists
}

// NewTable 创建新桌子
func (mgr *TableMgr) NewTable(roomID int32, users []string) inter.ITable {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// 分配新桌子ID
	tid := mgr.nextID
	mgr.nextID++ // 重启的话这个id会重复 TODO

	// 创建新桌子（这里使用基础实现，实际应用可替换为具体游戏实现）
	table := &BlackjackTable{
		BaseTable: BaseTable{
			id:     tid,
			roomID: roomID,
		},
		decksNum: 8, // todo 这里可以读配置
	}

	players := make([]inter.IPlayer, len(users))

	for k, v := range users {
		userInfo, _ := game_center.GetUserInfoByID(v)
		players[k] = NewPlayer(table, userInfo, int32(k))
	}

	table.players = players

	// 添加到全局表
	mgr.tables[tid] = table

	// 添加到房间
	if _, exists := mgr.roomMap[roomID]; !exists {
		mgr.roomMap[roomID] = make([]int32, 0)
	}
	mgr.roomMap[roomID] = append(mgr.roomMap[roomID], tid)

	// 启动游戏
	table.StartGame()

	return table
}

// GetTablesInRoom 获取房间中的所有桌子（扩展功能）
func (mgr *TableMgr) GetTablesInRoom(roomID int32) []inter.ITable {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	tids, exists := mgr.roomMap[roomID]
	if !exists {
		return nil
	}

	tables := make([]inter.ITable, 0, len(tids))
	for _, tid := range tids {
		if table, ok := mgr.tables[tid]; ok {
			tables = append(tables, table)
		}
	}

	return tables
}

// ReleaseTable 释放桌子资源（扩展功能）
func (mgr *TableMgr) ReleaseTable(tid int32) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	table, exists := mgr.tables[tid]
	if !exists {
		return errors.New("table not found")
	}

	// 结束游戏
	table.EndGame()

	// 从全局表中移除
	delete(mgr.tables, tid)

	// 从房间映射中移除
	if roomID, ok := table.(interface{ GetRoomID() int32 }); ok {
		roomID := roomID.GetRoomID()
		if tids, exists := mgr.roomMap[roomID]; exists {
			for i, id := range tids {
				if id == tid {
					mgr.roomMap[roomID] = append(tids[:i], tids[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}
