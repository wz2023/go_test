package inter

type IPlayer interface {
	GetUserID() string
	GetName() string
	GetSeatID() int32
	GetBalance() int32
}
