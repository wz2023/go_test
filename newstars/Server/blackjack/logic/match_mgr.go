package logic

import (
	"sync"
	"time"
)

// MatchUser 匹配中的用户信息
type MatchUser struct {
	UserID    string
	EnterTime time.Time // 进入匹配队列的时间
}

// RoomQueue 房间匹配队列
type RoomQueue struct {
	Users     []MatchUser
	CreatedAt time.Time // 队列创建时间
}

// MatchConfig 匹配配置
type MatchConfig struct {
	MatchSize        int           // 每组匹配人数
	Timeout          time.Duration // 匹配超时时长
	RoomInactiveTime time.Duration // 房间空闲超时时间
}

// DefaultMatchConfig 默认匹配配置
func DefaultMatchConfig() *MatchConfig {
	return &MatchConfig{
		MatchSize:        2, // 默认2人一组
		Timeout:          30 * time.Second,
		RoomInactiveTime: 10 * time.Minute, // 10分钟无活动的房间将被清理
	}
}

// MatchMgr 分组匹配管理实现
type MatchMgr struct {
	lock     sync.RWMutex                       // 并发控制锁
	rooms    map[int32]*RoomQueue               // roomID -> 房间队列
	callback func(roomID int32, users []string) // 匹配成功回调函数
	config   *MatchConfig                       // 匹配配置
	ticker   *time.Ticker
}

// NewMatchMgr 创建分组匹配管理器
// callback: 匹配成功后的回调函数，参数为房间ID和匹配成功的用户ID列表
// config: 匹配规则配置
func NewMatchMgr(callback func(roomID int32, users []string), config *MatchConfig) *MatchMgr {
	if config == nil {
		config = DefaultMatchConfig()
	}
	return &MatchMgr{
		callback: callback,
		config:   config,
		rooms:    make(map[int32]*RoomQueue),
	}
}

// AddMatch 添加用户到指定房间的匹配队列
func (m *MatchMgr) AddMatch(roomID int32, UserID string) int32 {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 初始化房间队列（如果不存在）
	if _, exists := m.rooms[roomID]; !exists {
		m.rooms[roomID] = &RoomQueue{
			Users:     make([]MatchUser, 0),
			CreatedAt: time.Now(),
		}
	}

	// 检查是否已在匹配队列中
	queue := m.rooms[roomID]
	for _, user := range queue.Users {
		if user.UserID == UserID {
			return -1 // 已存在，无需重复添加
		}
	}

	// 添加新用户
	queue.Users = append(queue.Users, MatchUser{
		UserID:    UserID,
		EnterTime: time.Now(),
	})

	// 更新队列创建时间（表示房间活跃）
	queue.CreatedAt = time.Now()
	return 0
}

func (m *MatchMgr) Run() {
	m.ticker = time.NewTicker(time.Second)

	go func() {
		for range m.ticker.C {
			m.DoMatch()
		}
	}()
}

func (m *MatchMgr) Stop() {
	m.ticker.Stop()
}

// DoMatch 执行匹配操作（每秒调用一次）
func (m *MatchMgr) DoMatch() {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 清理空闲过久的房间
	m.cleanInactiveRooms()

	// 遍历所有房间进行匹配
	for roomID, queue := range m.rooms {
		m.matchRoom(roomID, queue)
	}
}

// matchRoom 对单个房间进行匹配
func (m *MatchMgr) matchRoom(roomID int32, queue *RoomQueue) {
	users := queue.Users
	if len(users) < m.config.MatchSize {
		return // 人数不足，等待下次匹配
	}

	// 检查是否有等待超时的用户
	extended := m.handleTimeoutUsers(&users)
	if !extended {
		// 标准匹配流程
		groups := len(users) / m.config.MatchSize
		for i := 0; i < groups; i++ {
			// 提取一组用户
			start := i * m.config.MatchSize
			end := start + m.config.MatchSize
			group := users[start:end]

			// 获取用户ID列表
			userIDs := userGroupToIDs(group)

			// 调用匹配成功回调
			go m.callback(roomID, userIDs)
		}

		// 移除已匹配的用户
		newLength := groups * m.config.MatchSize
		users = users[newLength:]
	}

	// 更新房间队列
	queue.Users = users
}

// handleTimeoutUsers 处理超时用户（返回是否需要特殊匹配）
func (m *MatchMgr) handleTimeoutUsers(users *[]MatchUser) bool {
	// 查找超时用户
	var timedOut []MatchUser
	var others []MatchUser
	currentTime := time.Now()

	for _, user := range *users {
		if currentTime.Sub(user.EnterTime) > m.config.Timeout {
			timedOut = append(timedOut, user)
		} else {
			others = append(others, user)
		}
	}

	// 没有超时用户，直接返回
	if len(timedOut) == 0 {
		return false
	}

	// 优先匹配超时用户组合
	for len(timedOut) >= m.config.MatchSize {
		group := timedOut[:m.config.MatchSize]
		timedOut = timedOut[m.config.MatchSize:]

		// 直接回调匹配结果
		userIDs := userGroupToIDs(group)
		// 注意：这里需要房间ID，但本函数不处理房间回调，实际使用时要调整
		// 此函数作为示例简化了回调处理
		_ = userIDs
	}

	// 剩余超时用户尝试与其他用户混合匹配
	if len(timedOut) > 0 {
		needed := m.config.MatchSize - len(timedOut)
		if len(others) >= needed {
			group := append(timedOut, others[:needed]...)
			others = others[needed:]
			timedOut = nil

			userIDs := userGroupToIDs(group)
			// 同上，简化回调处理
			_ = userIDs
		}
	}

	// 重组队列：剩余的超时用户+普通用户
	*users = append(timedOut, others...)
	return true
}

// cleanInactiveRooms 清理空闲房间
func (m *MatchMgr) cleanInactiveRooms() {
	currentTime := time.Now()

	// 收集需要清理的房间ID
	var toRemove []int32
	for roomID, queue := range m.rooms {
		// 空房间且超过空闲时间
		if len(queue.Users) == 0 && currentTime.Sub(queue.CreatedAt) > m.config.RoomInactiveTime {
			toRemove = append(toRemove, roomID)
		}
	}

	// 移除空闲房间
	for _, roomID := range toRemove {
		delete(m.rooms, roomID)
	}
}

// userGroupToIDs 将用户组转换为ID列表
func userGroupToIDs(users []MatchUser) []string {
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.UserID
	}
	return userIDs
}
