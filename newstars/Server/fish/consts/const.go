package consts

// 常量
const (
	ErrorInvalidKindID = -800
	ErrorInvalidParams = -801
	ErrorDB            = -802
	ErrorBalance       = -803
	ErrorTakeTable     = -804
	ErrorSitDown       = -805
	ErrorQuerySeat     = -806
	ErrorInvalidUID    = -807
	ErrorInvalidTid    = -808
	ErrorUserWealth    = -809

	FishKindID         = 100007
	FishSeatNumbers    = 4
	MaxMinRatio        = 10.00
	FishKindRateBase   = 100  //鱼种出现概率基数
	CaptureRate        = 0.98 //捕获概率
	MaxFishSeed        = 4
	DefaultConnonID    = 0
	DefaultConnonRatio = 0.001
	RoomMinEnterAmount = 0.001
	SpecialRatio       = 0.000001 //0.001-0.1炮倍之间差值不一样
	SpecialRatioLevel  = 12       //0.001炮倍有12级
	MaxLargeFish       = 3
	SpecilLimitAmount  = 200.00 //跑马灯推送金额

	SkillCDTime = 60 * 1000
)

// 桌位状态
const (
	SeatStatusNone = iota
	SeatStatusOk
)

// 游戏状态
const (
	GameStatusNone = iota
	GameStatusFreeFish
	GameStatusReadyBoss
	GameStatusReadyFishTide
	GameStatusStartFishTide
	GameStatusStartingFishTide
)

const (
	TimeForStart        = 3
	TimeForInter        = 1
	TimeForTideDelay    = 20
	TimeForFreedomDelay = 5
	TimeForBoss         = 4 * 1000
	TimeForRandom       = 5
	TimeForBossInterval = 180 * 1000 //毫秒
	TimeForTideInterval = 360 * 1000
	TimeForWave         = 5 * 1000
)

const (
	CannonNotBuy = iota
	CannonNotNotload
	CannonInloaded
)

const (
	NewLayerProtectTime   = 300 * 1000 //新手玩家在线时长保护期
	NewLayerProtectAmount = 3.0        //新手玩家最大金币
	NewplayerLessAmount   = 0.45       //新手玩家过少金币
	NewPalyerRoomid       = 17         //新手场概率不受库存概率影响
)

const (
	RecordBullet = iota
	RecordSkill
	RecordPurchase
)

var PlatformTableMap = make(map[string]int32)

const (
	FISH_GAME_ID      = 200009          // 捕鱼中台ID
	FISH_REAL_GAME_ID = 100007          // 游戏ID
	FISH_NAME         = "fisher tycoon" // 游戏名称
)
