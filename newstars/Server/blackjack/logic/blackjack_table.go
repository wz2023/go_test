package logic

// BlackjackTable 21点游戏桌实现
type BlackjackTable struct {
	BaseTable                      // 嵌入基础实现
	currentTurn string             // 当前轮到哪位玩家
	decksNum    int                // 牌副数
	decks       []int              // 牌池
	hands       map[string][]int32 // 玩家手牌
	dealerHands []int32            // 庄家牌
}
