package core

// ServerConf conf
type ServerConf struct {
	ID     int
	Type   string
	Addr   string
	Status int
}

// Conf file
var Conf struct {
	Servers []*ServerConf
}

// const (
// 	hallServerType      = "hall"
// 	landlordsServerType = "landlords"
// )

// const (
// 	sessionKeyRoomID    = "room"
// 	sessionKeyTableID   = "table"
// 	sessionKeySeatID    = "seat"
// 	sessionKeyUserID    = "user"
// 	sessionKeyRoundName = "round"
// )

const (
	multipleLogin     = -200
	invalidUserid     = -201
	errFindLinkServer = -202
	errStatus         = -9999
)
