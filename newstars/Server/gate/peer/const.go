package peer

const (
	maxSessions = 100000
)

// 类型
const (
	TypeHall      = "Hall"
	TypeFish      = "Fish"
	TypeBlackjack = "Blackjack"

	TypeLandlord  = "Landlord"
	TypeThreecard = "Threecard"
	TypeOx100     = "Ox100"
	TypeRedBlack  = "RedBlack"
	TypeOx5       = "Ox5"
	TypeDT        = "DT"
	TypePay       = "PAY"
	TypeBenz      = "Benz"
	TypePoker13   = "Poker13"
	TypeActivity  = "Activity"
	TypeBaccarat  = "Baccarat"
	TypeWagency   = "Wagency"
	TypeDuofu     = "Duofu"
)

var GameKindMap = map[int32]string{
	//1:  TypeLandlord,
	//2:  TypeThreecard,
	//3:  TypeOx100,
	//4:  TypeRedBlack,
	//5:  TypeOx5,
	//6:  TypeDT,
	//7:  TypeBenz,
	1: TypeBlackjack,
	8: TypeFish,
	//9:  TypeBaccarat,
	//11: TypeDuofu,
}

// key
const (
	KeyTableID   = "TableID"
	KeySeatID    = "SeatID"
	KeyRoundName = "RoundName"
	KeyRoomID    = "RoomID"
	KeyUserID    = "UserID"
	KeyLeaveType = "LeaveType"
	KeyGameID    = "GameID"
	KeyUserToken = "UserToken"
)

// ServerConf conf
type ServerConf struct {
	ID       int
	Type     string
	Addr     string
	Status   int
	Reserved int
}

// Conf file
var Conf struct {
	GameNumber int32
	Host       string
	Heartbeat  int64
	Servers    []*ServerConf
}

var GGameNumber int32

// 用户所在游戏
const (
	StatusHall byte = iota
	StatusBlackjack
	StatusFish
)
