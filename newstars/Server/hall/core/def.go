package core

const (
	heatTimes = 30
)

const (
	psNone = iota
	psPlaying
)

type gameKindTable struct {
	kindid   int32
	kindname string
	status   int32
	icon     int32
}

type userTable struct {
	userid        int32
	usertoken     string
	platformid    int32
	usercode      string
	userlevel     int32
	userstatus    int32
	nickname      string
	faceid        int32
	isrobot       int32
	createtime    string
	lastlogontime string
	description   string
}

type accInfo struct {
	nick      string
	wealth    float64
	iparea    string
	faceid    int32
	sexuatily int32
}

type userStatus struct {
	userid     string
	playStatus int32
	kindid     int32
	roomid     int32
	tableid    int32
	seatid     int32
	nickname   string
	roundname  string
	wealth     float64
}

// Settlement 结算表
type Settlement struct {
	Roundcode    string
	Actualamount float64
	Odds         float64
	Bettype      int32
	Payoffvalue  float64
	Settletime   int64
	Results      string
}
