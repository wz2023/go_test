package core

import (
	"database/sql"
	"go.uber.org/zap"
	"newstars/Protocol/plr"
	"newstars/Server/hall/conf"
	"newstars/Server/hall/errcode"
	"newstars/Server/hall/login"
	"newstars/framework/core/component"
	server2 "newstars/framework/core/server"
	"newstars/framework/core/session"
	data2 "newstars/framework/game_center"
	"newstars/framework/glog"
	"newstars/framework/util/ip17mon"
)

// HallCore core
type HallCore struct {
	component.Base
	db        *sql.DB
	db2       *sql.DB
	users     map[string]*userStatus
	login     *login.UserLogin
	luckyodds map[string]float64
	rs        map[string]int64
}

// NewHallCore returns a new GateCore
func NewHallCore(p, db2 *sql.DB) *HallCore {
	return &HallCore{
		db:        p,
		db2:       db2,
		users:     make(map[string]*userStatus),
		login:     login.NewUserLogin(p),
		luckyodds: make(map[string]float64),
		rs:        make(map[string]int64),
	}
}

// Init init hallcore
func (w *HallCore) Init() {
	err := ip17mon.Init("conf/17monipdb.dat")
	if err != nil {
		glog.SErrorf("ip17mon Init conf/17monipdb.dat failed %v", err)
	}
	w.login.Go()
}

// AfterInit bind close event
func (w *HallCore) AfterInit() {
	server2.OnSessionClosed(func(s *session.Session) {
		glog.SErrorf("Connect gate server interrupt.addr:%v", s.RemoteAddr())
		//w.login.LogoutAll()
		w.login.Invoke(w.login.LogoutAll)
		w.users = make(map[string]*userStatus)
	})

	//server2.NewTimer(30*time.Minute, func() {
	//	//w.login.ClearExpireMobile()
	//	w.login.Invoke(w.login.ClearExpireMobile)
	//})

	conf.LoadExchangeConf(w.db)
}

func (w *HallCore) C0000003(s *session.Session, msg *plr.C0000003, mid uint) error {
	uid := msg.GetUserID()
	data := &plr.S0000003{}

	u, ok := w.users[uid]
	if ok {
		kindid, err := w.queryKindIDByRoom(u.roomid)
		if err != nil {
			glog.SErrorf("[获取用户信息C0000003] queryKindIDByRoom failed %v", err)
			data.RetCode = errcode.DBError
			return s.Response(data, mid)
		}
		data.KindID = kindid // 游戏ID
		data.RoomID = u.roomid
		data.RoundName = u.roundname
		data.SeatNo = u.seatid
		data.Status = u.playStatus
		data.TableID = u.tableid
	}

	uinfo, err := data2.GetUserInfoByID(uid)
	if err != nil {
		glog.SErrorf("[获取用户信息C0000003] GetUserInfoByID failed %v uid:%v", err, uid)
		data.RetCode = errcode.DBError
		return s.Response(data, mid)
	}

	data.Wealth = uinfo.Wealth
	data.UserName = uinfo.NickName
	data.IPArea = uinfo.DisPlayName
	data.FaceID = uinfo.FaceID
	data.Sexuality = uinfo.Sexuality
	data.AccType = uinfo.AccType
	data.UserID = uinfo.UserID
	data.FaceFrameID = uinfo.FaceFrameID
	glog.SInfo("[C0000003 获取用户信息] result", zap.Any("data", data))
	return s.Response(data, mid)
}

func (w *HallCore) queryKindIDByRoom(rid int32) (int32, error) {
	var kindid int32
	stmt, err := w.db.Prepare(`select gamekindid from gameroom_t where gameroomid = ?`)
	if err != nil {
		glog.SErrorf("prepare queryKindIDByRoom sql stmt fail %v", err)
		return kindid, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(rid).Scan(&kindid)
	return kindid, err
}

// // C0000001 proccess 登陆请求
//
//	func (w *HallCore) C0000001(s *session.Session, msg *plr.C0000001, mid uint) error {
//		w.login.Invoke(func() {
//			sMsg := w.login.AccountLogin2(msg)
//			s.Response(sMsg, mid)
//		})
//		return nil
//	}
//
// // C0000002 查询游戏种类
//
//	func (w *HallCore) C0000002(s *session.Session, msg *plr.C0000002, mid uint) error {
//		data := &plr.S0000002{}
//		kinds, err := models.QueryGameKind(w.db)
//		if err != nil {
//			data.RetCode = errcode.DBError
//			data.Error = err.Error()
//			return s.Response(data, mid)
//		}
//		for _, v := range kinds {
//			t := &plr.S0000002_KindType{}
//			t.KindID = v.KindID
//			t.KindName = v.KindName
//			t.Status = v.Status
//			t.IconType = v.IconType
//			data.KindTypes = append(data.KindTypes, t)
//		}
//		data.RetCode = errcode.CodeOK
//		return s.Response(data, mid)
//	}

// C0000003 查询玩家信息
//func (w *HallCore) C0000003(s *session.Session, msg *plr.C0000003, mid uint) error {
//	uid := msg.GetUserID()
//	data := &plr.S0000003{}
//
//	u, ok := w.users[uid]
//	if ok {
//		kindid, err := w.queryKindIDByRoom(u.roomid)
//		if err != nil {
//			glog.SErrorf("queryKindIDByRoom failed %v", err)
//			data.RetCode = errcode.DBError
//			return s.Response(data, mid)
//		}
//		data.KindID = kindid
//		data.RoomID = u.roomid
//		data.RoundName = u.roundname
//		data.SeatNo = u.seatid
//		data.Status = u.playStatus
//		data.TableID = u.tableid
//	}
//
//	acc, err := model.QueryUserInfo(uid, w.db)
//	if err != nil {
//		glog.SErrorf("QueryUserInfo failed %v uid:%v", err, uid)
//		data.RetCode = errcode.DBError
//		return s.Response(data, mid)
//	}
//
//	viplvl, _, err := model.QueryVipLevel(uid, w.db)
//	if err != nil {
//		glog.SErrorf("query vipLevel failed.uid:%v,err:%v", uid, err)
//	}
//
//	data.Wealth = acc.Wealth
//	data.UserName = acc.NickName
//	data.IPArea = acc.DisPlayName
//	data.FaceID = acc.FaceID
//	data.Sexuality = acc.Sexuality
//	data.AccType = acc.AccType
//	data.UserID = uid
//	data.VipLevel = viplvl
//	data.FaceFrameID = acc.FaceFrameID
//	return s.Response(data, mid)
//}

// N0000001 用户入座
func (w *HallCore) N0000001(s *session.Session, msg *plr.N0000001, mid uint) error {
	uid := msg.GetUserID()
	rid := msg.GetRoomID()
	tid := msg.GetTableID()
	sid := msg.GetSeatNo()

	us := &userStatus{}
	us.userid = uid
	us.roomid = rid
	us.tableid = tid
	us.seatid = sid
	w.users[uid] = us
	return nil
}

// N0000002 用户离座
func (w *HallCore) N0000002(s *session.Session, msg *plr.N0000002, mid uint) error {
	uid := msg.GetUserID()
	v, ok := w.users[uid]
	if ok {
		if v.playStatus == psNone {
			delete(w.users, uid)
		}
	}
	return nil
}

// N0000003 开始牌局
func (w *HallCore) N0000003(s *session.Session, msg *plr.N0000003, mid uint) error {
	tid := msg.GetTableID()
	roundname := msg.GetRoundName()

	for _, v := range w.users {
		if v.tableid == tid {
			v.roundname = roundname
			v.playStatus = psPlaying
		}
	}

	return nil
}

// N0000004 结束牌局
func (w *HallCore) N0000004(s *session.Session, msg *plr.N0000004, mid uint) error {
	roundname := msg.GetRoundName()
	glog.SInfo("[N0000004 结束牌局] result", zap.Any("roundname", roundname))

	us := make([]string, 0)
	for _, v := range w.users {
		if v.roundname == roundname {
			us = append(us, v.userid)
		}
	}

	for _, v := range us {
		delete(w.users, v)
	}

	return nil
}

// N0000006 开始
func (w *HallCore) N0000006(s *session.Session, msg *plr.N0000006, mid uint) error {
	uid := msg.GetUserID()
	rid := msg.GetRoomID()
	tid := msg.GetTableID()
	sid := msg.GetSeatNo()

	us := &userStatus{}
	us.userid = uid
	us.roomid = rid
	us.tableid = tid
	us.seatid = sid
	us.playStatus = psPlaying
	w.users[uid] = us
	return nil
}

// N0000007 结束
func (w *HallCore) N0000007(s *session.Session, msg *plr.N0000007, mid uint) error {
	delete(w.users, msg.GetUserID())
	return nil
}

//// C0000004 guest reg
//func (w *HallCore) C0000004(s *session.Session, msg *plr.C0000004, mid uint) error {
//	w.login.Invoke(func() {
//		sMsg := w.login.GuestLogin(msg)
//		s.Response(sMsg, mid)
//	})
//	return nil
//}
//
//// C0000006  AccountRegister
//func (w *HallCore) C0000006(s *session.Session, msg *plr.C0000006, mid uint) error {
//	sMsg := w.login.AccountRegister(msg)
//	return s.Response(sMsg, mid)
//}

// N0000005  logout
func (w *HallCore) N0000005(s *session.Session, msg *plr.N0000005, mid uint) error {
	glog.SInfo("[N0000005 logout] result", zap.Any("uid", msg.UserID))
	uid := msg.GetUserID()
	v, ok := w.users[uid]
	if ok {
		if v.playStatus == psNone {
			delete(w.users, uid)
		}
	}
	w.login.Invoke(func() {
		w.login.AccountLogout(uid)
	})
	return nil
}

//// N0000008 get phone info
//func (w *HallCore) N0000008(s *session.Session, msg *plr.N0000008, mid uint) error {
//	var machineCode, lastSysVer, clientVer string
//	if msg.GetMachineCode() == "" {
//		glog.SErrorf("machineCode can not be empty")
//		return nil
//	}
//	err := w.db.QueryRow(`select machinecode,last_system_version,client_version from phone_info_t where machinecode=?`, msg.GetMachineCode()).Scan(&machineCode, &lastSysVer, &clientVer)
//	if err != nil {
//		if err == sql.ErrNoRows {
//			_, err = w.db.Exec(`INSERT INTO phone_info_t(machinecode,manufacturer,phone_model,phone_code,resolution,
//				init_system_version,last_system_version,client_version,inside_ip) VALUES(?,?,?,?,?,?,?,?,?)
//				`, msg.GetMachineCode(), msg.GetManufacturer(), msg.GetPhoneModel(), msg.GetPhoneCode(), msg.GetResolution(), msg.GetSystemVersion(), msg.GetSystemVersion(), msg.GetClientVersion(), msg.GetInsideIp())
//			if err != nil {
//				glog.SErrorf("insert phone_info_t fail, msg:%v,err:%v", msg, err)
//			}
//		} else {
//			glog.SErrorf("select phone_info_t fail, err:%v", err)
//		}
//	} else {
//		if lastSysVer != msg.GetSystemVersion() || clientVer != msg.GetClientVersion() {
//			_, err = w.db.Exec(`update phone_info_t set last_system_version=?,client_version =? where machinecode=? `, msg.GetSystemVersion(), msg.GetClientVersion(), msg.GetMachineCode())
//			if err != nil {
//				glog.SErrorf("update phone_info_t fail, msg:%v,err:%v", msg, err)
//			}
//		}
//	}
//	return nil
//}
//
//// N0000010 邮件通知
//func (w *HallCore) N0000010(s *session.Session, msg *plr.N0000010, mid uint) error {
//	//uid := msg.GetUserID()
//	//if uid != 0 {
//	//	glog.SErrorf("N0000010 param invalid.  userid:%v", uid)
//	//	return nil
//	//}
//	//mails, err := mail.QueryUnSendMail(w.db)
//	//if err != nil {
//	//	glog.SErrorf("QueryUnSendMail failed. err:%v", err)
//	//	return nil
//	//}
//	//currentTime := time.Now().Unix()
//	//for sendTime, userids := range mails {
//	//	if currentTime < sendTime {
//	//		server.NewCountTimer(time.Duration(sendTime-currentTime)*time.Second, 1, func() {
//	//			mail.MailNotify(userids, s)
//	//		})
//	//	} else {
//	//		glog.SErrorf("N0000010 mail push failed.time expired:%v %v", sendTime, currentTime)
//	//	}
//	//}
//	//glog.SInfof("N0000010 mails:%v", mails)
//	return nil
//}
//
//// C0000009 更新头像
//func (w *HallCore) C0000009(s *session.Session, msg *plr.C0000009, mid uint) error {
//
//	sMsg := &plr.S0000009{}
//
//	_, err := w.db.Exec(`update account_t set faceid = ?,sexuatily = ?,faceframeid=?
//		where userid = ?`, msg.GetFaceID(), msg.GetSexuality(), msg.GetFaceFrameID(), msg.GetUserID())
//	if err != nil {
//		sMsg.RetCode = errcode.DBError
//		glog.SErrorf("update account_t failed.err:%v msg:%v", err, msg)
//	} else {
//		sMsg.RetCode = errcode.CodeOK
//	}
//
//	return s.Response(sMsg, mid)
//}
//
//// C000000A 登出
//func (w *HallCore) C000000A(s *session.Session, msg *plr.C000000A, mid uint) error {
//	uid := msg.GetUserID()
//	w.login.Invoke(func() {
//		sMsg := &plr.S000000A{}
//		err := w.login.AccountLogout(uid)
//		if err != nil {
//			glog.SErrorf("Uid:%v Logout failed.err:%v", uid, err)
//			sMsg.RetCode = errcode.InvalidParamError
//		}
//		s.Response(sMsg, mid)
//	})
//	return nil
//}
//
//// C0000014 获取代理
//func (w *HallCore) C0000014(s *session.Session, msg *plr.C0000014, mid uint) error {
//	rMsg := &plr.S0000014{}
//	uid := msg.GetUserID()
//
//	var platformid, gameid int32
//	err := w.db.QueryRow(`select platformid,game_id from account_t where userid=?`, uid).Scan(&platformid, &gameid)
//	if err != nil {
//		glog.SErrorf("query paltfromid failed.err:%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	agents, err := models.QueryAgentList(w.db, platformid, gameid)
//	if err != nil {
//		glog.SErrorf("QueryAgentList failed.err:%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	//如果查不到再查默认的代理
//	if len(agents) == 0 {
//		agents, err = models.QueryAgentList(w.db, -1, gameid)
//		if err != nil {
//			glog.SErrorf("QueryAgentList failed.err:%v", err)
//			rMsg.RetCode = errcode.DBError
//			return s.Response(rMsg, mid)
//		}
//	}
//
//	for k := range agents {
//		item := &plr.S0000014_AgentInfo{}
//		item.AgentName = agents[k].Name
//		item.QQ = agents[k].QQ
//		item.WebChat = agents[k].Weixin
//		rMsg.Agents = append(rMsg.Agents, item)
//	}
//	return s.Response(rMsg, mid)
//}
//
//// C0000012 修改昵称
//func (w *HallCore) C0000012(s *session.Session, msg *plr.C0000012, mid uint) error {
//	uid := msg.GetUserID()
//	nickName := msg.GetNickName()
//
//	rMsg := &plr.S0000012{}
//	// var count int
//	// err := w.db.QueryRow(`select count(*) from account_t where nickname = ?`, nickName).Scan(&count)
//	// if err != nil {
//	// 	glog.SErrorf("C0000012 for db error:%v", err)
//	// 	rMsg.RetCode = errcode.DBError
//	// 	return s.Response(rMsg, mid)
//	// }
//
//	// if count != 0 {
//	// 	glog.SErrorf("C0000012 NickNameDuplicate %v", nickName)
//	// 	rMsg.RetCode = errcode.NickNameDuplicate
//	// 	return s.Response(rMsg, mid)
//	// }
//
//	_, err := w.db.Exec(`update account_t set nickname = ? where userid = ?`, nickName, uid)
//	if err != nil {
//		glog.SErrorf("C0000012 for db error:%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	return s.Response(rMsg, mid)
//}
//
//// C0000011 保险箱存取
//func (w *HallCore) C0000011(s *session.Session, msg *plr.C0000011, mid uint) error {
//	return nil
//	//rMsg := &plr.S0000011{}
//	//
//	//uid := msg.GetUserID()
//	//amount := msg.GetAmount()
//	//opType := msg.GetOpType()
//	////password := msg.GetPassword()
//	//glog.SInfof("Client request C0000011 uid:%v amount:%v optype:%v", uid, amount, opType)
//	//
//	//if amount < 0.01 {
//	//	glog.SErrorf("C0000011 Invalid params amount %v", amount)
//	//	rMsg.RetCode = errcode.InvalidParamError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if opType != 1 && opType != 2 {
//	//	glog.SErrorf("C0000011 Invalid params opType %v", opType)
//	//	rMsg.RetCode = errcode.InvalidParamError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//u, ok := w.users[uid]
//	//if ok {
//	//	if u.playStatus == psPlaying {
//	//		glog.SErrorf("Invalid status is playing uid:%v", uid)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//}
//	//
//	//tx, err := w.db.Begin()
//	//
//	//if err != nil {
//	//	glog.SErrorf("C0000011 failed error:%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//var uWealth, uBank float64
//	////var oPassword sql.NullString
//	//err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, uid).Scan(&uWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//err = tx.QueryRow(`select bankamount from userbank_t where userid = ?`, uid).Scan(&uBank)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if opType == 1 {
//	//	if uWealth < amount {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000011 Invalid params amount %v less user wealth:%v", amount, uWealth)
//	//		rMsg.RetCode = errcode.InvalidParamError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//}
//	//
//	//if opType == 2 {
//	//	if uBank < amount {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000011 Invalid params amount %v less user uBank:%v", amount, uBank)
//	//		rMsg.RetCode = errcode.InvalidParamError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	//不验证密码
//	//	// if oPassword != w.login.Password(password) {
//	//	// 	tx.Rollback()
//	//	// 	glog.SErrorf("Withdraw money :password error ")
//	//	// 	rMsg.RetCode = errcode.PasswordError
//	//	// 	return s.Response(rMsg, mid)
//	//	// }
//	//}
//	//
//	//_, err = tx.Exec(`insert into record_bankinout_t (userid,amount,optype,optime) values(?,?,?,?)`, uid, amount, opType, time.Now().Unix())
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert bank inout failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if opType == 1 {
//	//	settlewid, err := models.InsertRecordAmount(uid, models.RecordToBank, "", tx)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	_, err = tx.Exec(`update userwealth_t set wealth = wealth - ? where userid = ?`, amount, uid)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("Update user wealth failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	_, err = tx.Exec(`update userbank_t set bankamount = bankamount + ? where userid = ?`, amount, uid)
//	//
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("Update user bank failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	err = models.UpdateRecordAmount(settlewid, uid, tx)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("update recordAmount failed for db.%v", err.Error())
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//} else {
//	//	settlewid, err := models.InsertRecordAmount(uid, models.RecordFromBank, "", tx)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	_, err = tx.Exec(`update userwealth_t set wealth = wealth + ? where userid = ?`, amount, uid)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("Update user wealth failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	_, err = tx.Exec(`update userbank_t set bankamount = bankamount - ? where userid = ?`, amount, uid)
//	//
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("Update user bank failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	err = models.UpdateRecordAmount(settlewid, uid, tx)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("update recordAmount failed for db.%v", err.Error())
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//}
//	//
//	//tx.Commit()
//	//
//	//uW, qErr := models.QueryUserWealth(uid, w.db)
//	//if qErr != nil {
//	//	glog.SErrorf("QueryUserWealth failed error:%v", qErr)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//rMsg.BankWealth = uW.BankWealth
//	//rMsg.CoinWealth = uW.CoinWealth
//	//
//	//return s.Response(rMsg, mid)
//}
//
//// C0000017 获取用户财富信息
//func (w *HallCore) C0000017(s *session.Session, msg *plr.C0000017, mid uint) error {
//	return nil
//	//uid := msg.GetUserID()
//	//rMsg := &plr.S0000017{}
//	//uWealth, err := models.QueryUserWealth(uid, w.db)
//	//if err != nil {
//	//	glog.SErrorf("QueryUserWealth failed error:%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//rMsg.BankWealth = uWealth.BankWealth
//	//rMsg.CoinWealth = uWealth.CoinWealth
//	//rMsg.BSetBankPassword = uWealth.BSetBankPassword
//	//return s.Response(rMsg, mid)
//}
//
//// C0000013 兑换
//func (w *HallCore) C0000013(s *session.Session, msg *plr.C0000013, mid uint) error {
//	return nil
//	//rMsg := &plr.S0000013{}
//	//
//	//uid := msg.GetUserID()
//	//amount := msg.GetAmount()
//	//device := msg.GetDevice()
//	//exType := msg.GetExchangeType()
//	//account := msg.GetAccount()
//	//accName := msg.GetAccName()
//	//machinecode := msg.GetMachineID()
//	//ip := msg.GetUserIP()
//	//bankcode := msg.GetBankCode()
//	//tel := msg.GetTelephone()
//	//
//	//glog.Infof("Client request exchange uid:%v amount:%v device:%v exType:%v account:%v machineid:%v ip:%v bankcode:%v", uid,
//	//	amount, device, exType, account, machinecode, ip, bankcode)
//	//
//	//ti, _ := time.Parse("2006-01-02 15:04:05", "2018-05-31 00:00:00")
//	//if time.Now().After(ti) {
//	//	if amount < 100 {
//	//		glog.SErrorf("C0000013 Invalid params amount %v should be more than 100", amount)
//	//		rMsg.RetCode = errcode.AmountlimitError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//}
//	//
//	//if device != 1 && device != 2 {
//	//	glog.SErrorf("C0000013 Invalid params device %v", device)
//	//	rMsg.RetCode = errcode.InvalidParamError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if exType != 0 && exType != 1 {
//	//	glog.SErrorf("C0000013 Invalid params exType %v", exType)
//	//	rMsg.RetCode = errcode.InvalidParamError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if ip == "" {
//	//	glog.SErrorf("C0000013 empty ip param. uid:%v", uid)
//	//	rMsg.RetCode = errcode.IPEmptyIP
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//swexchange := false
//	//excfg, err := models.QueryExchangeBYIP(ip, w.db)
//	//if err == nil {
//	//	if excfg != "" {
//	//		arr := strings.Split(excfg, ",")
//	//		for _, v := range arr {
//	//			if exType == 0 && v == "0" {
//	//				swexchange = true
//	//			} else if exType == 1 && v == "1" {
//	//				swexchange = true
//	//			}
//	//		}
//	//	}
//	//} else if err == sql.ErrNoRows {
//	//	swexchange = true
//	//} else {
//	//	glog.SErrorf("C0000013 ip exchange limit  uid:%v,ip:%v", uid, ip)
//	//	rMsg.RetCode = errcode.IPForbid
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if !swexchange {
//	//	glog.SErrorf("C0000013 QueryExchangeBYIP failed uid:%v,ip:%v", uid, ip)
//	//	rMsg.RetCode = errcode.IPForbid
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//tx, err := w.db.Begin()
//	//
//	//if err != nil {
//	//	glog.SErrorf("C0000013 failed error:%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//var uTel string
//	//err = tx.QueryRow(`select mobile from account_t where userid = ?`, uid).Scan(&uTel)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//if tel != uTel {
//	//	tx.Rollback()
//	//	glog.SErrorf("User telephone error.%v:%v", tel, uTel)
//	//	rMsg.RetCode = errcode.InvalidParamError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//var accWealth float64
//	//err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, uid).Scan(&accWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//var uWealth float64
//	//var uAccount, uAccName, uBankCode string
//	//err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, uid).Scan(&uWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//glog.SInfof("C0000013 before exchange user  %v wealth is %v", uid, uWealth)
//	//if uWealth < amount+10 {
//	//	tx.Rollback()
//	//	glog.SErrorf("C0000013 Invalid params amount %v should be less than %v+10", amount, uWealth)
//	//	rMsg.RetCode = errcode.WealthNotEnough
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	////支付宝
//	//if exType == 0 {
//	//	err = tx.QueryRow(`SELECT alipay,alipaydetail FROM  userextend_t where userid=?`, uid).Scan(&uAccName, &uAccount)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("DB failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	//空值判断
//	//	if account == "" || accName == "" {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000013 empty params account:%v  user account:%v", account, accName)
//	//		rMsg.RetCode = errcode.AlipayExsitError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	if uAccount != account || uAccName != accName {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000013 Invalid params account:%v[%v]  user account:%v[%v]", account, accName, uAccount, uAccName)
//	//		rMsg.RetCode = errcode.AlipayExsitError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	//必须绑定银行卡
//	//	// err = tx.QueryRow(`SELECT ifnull(bankaccount,""),ifnull(bankno,""),ifnull(bankdetail,"") FROM  userextend_t where userid=?`, uid).Scan(&uAccName, &uAccount, &uBankCode)
//	//	// if err != nil {
//	//	// 	tx.Rollback()
//	//	// 	glog.SErrorf("DB failed.%v", err)
//	//	// 	rMsg.RetCode = errcode.DBError
//	//	// 	return s.Response(rMsg, mid)
//	//	// }
//	//
//	//	// if uAccount == "" || uAccName == "" || uBankCode == "" {
//	//	// 	tx.Rollback()
//	//	// 	glog.SErrorf("C0000013 Invalid params  user account:%v[%v] bankcode:[%v]", uAccount, uAccName, uBankCode)
//	//	// 	rMsg.RetCode = errcode.BankCardExsitError
//	//	// 	return s.Response(rMsg, mid)
//	//	// }
//	//
//	//	if conf.ExchangeConf.AilPayInterval != 0 {
//	//		var chargecount int32
//	//		limittime := time.Now().Unix() - int64(conf.ExchangeConf.AilPayInterval*60)
//	//		err := w.db.QueryRow(`select count(1) as cnt from
//	//		exchange_t where apply_time>? and user_id=? and channel=0`,
//	//			limittime, msg.GetUserID()).Scan(&chargecount)
//	//		if err != nil {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.DBError
//	//			glog.SErrorf("query exchange_t failed.err:%v", err)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//		if chargecount > 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.PayBusyError
//	//			glog.SErrorf("operate too busy.")
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//	}
//	//
//	//	// var uamount float64
//	//	var count int32
//	//	if conf.ExchangeConf.AlipayLimitAmount != 0 || conf.ExchangeConf.AlipayLimitCount != 0 {
//	//		curday := time.Now().Format("20060102")
//	//		curtime, _ := time.ParseInLocation("20060102", curday, time.Local)
//	//		startSec := curtime.Unix()
//	//		endSec := curtime.Add(24 * 3600 * time.Second).Unix()
//	//
//	//		err := w.db.QueryRow(`select count(1) as cnt from
//	//		 exchange_t where apply_time between ? and ? and user_id=? and channel=0 and status<9`, startSec, endSec, msg.GetUserID()).Scan(&count)
//	//		if err != nil {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.DBError
//	//			glog.SErrorf("query exchange_t failed.err:%v", err)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//	}
//	//
//	//	if conf.ExchangeConf.AlipayLimitAmount != 0 {
//	//		leftAmount := conf.ExchangeConf.AlipayLimitAmount - amount
//	//		if leftAmount < 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.AlipayLimitedError
//	//			glog.SErrorf("alipay limitamount:%v usedamount:%v", conf.ExchangeConf.AlipayLimitAmount, amount)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//		// if leftAmount-amount < 0 {
//	//		// 	tx.Rollback()
//	//		// 	rMsg.RetCode = errcode.AlipayLimitingError
//	//		// 	glog.SErrorf("alipay left amount:%v,exhangamount:%v", leftAmount, amount)
//	//		// 	s.Response(rMsg, mid)
//	//		// 	return nil
//	//		// }
//	//	}
//	//
//	//	if conf.ExchangeConf.AlipayLimitCount != 0 {
//	//		leftCount := conf.ExchangeConf.AlipayLimitCount - count
//	//		if leftCount <= 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.AlipayCountError
//	//			glog.SErrorf("alipay limit left count:%v", leftCount)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//	}
//	//
//	//}
//	////银行卡
//	//if exType == 1 {
//	//	err = tx.QueryRow(`SELECT bankaccount,bankno,bankdetail FROM  userextend_t where userid=?`, uid).Scan(&uAccName, &uAccount, &uBankCode)
//	//	if err != nil {
//	//		tx.Rollback()
//	//		glog.SErrorf("DB failed.%v", err)
//	//		rMsg.RetCode = errcode.DBError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	//空值判断
//	//	if account == "" || accName == "" || bankcode == "" {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000013 empty params account:%v  user account:%v bankcode:%v", account, accName, bankcode)
//	//		rMsg.RetCode = errcode.AlipayExsitError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//	if uAccount != account || uAccName != accName || uBankCode != bankcode {
//	//		tx.Rollback()
//	//		glog.SErrorf("C0000013 Invalid params account:%v[%v]  user account:%v[%v] bankcode:%v[%v]", account, accName, uAccount, uAccName, bankcode, uBankCode)
//	//		rMsg.RetCode = errcode.AlipayExsitError
//	//		return s.Response(rMsg, mid)
//	//	}
//	//
//	//	if conf.ExchangeConf.BankCardInterval != 0 {
//	//		var chargecount int32
//	//		limittime := time.Now().Unix() - int64(conf.ExchangeConf.BankCardInterval*60)
//	//		err := w.db.QueryRow(`select count(1) as cnt from
//	//		exchange_t where apply_time>? and user_id=? and channel=1`,
//	//			limittime, msg.GetUserID()).Scan(&chargecount)
//	//		if err != nil {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.DBError
//	//			glog.SErrorf("query exchange_t failed.err:%v", err)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//		if chargecount > 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.PayBusyError
//	//			glog.SErrorf("operate too busy.")
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//	}
//	//
//	//	// var uamount float64
//	//	var count int32
//	//	if conf.ExchangeConf.BankCardLimitAmount != 0 || conf.ExchangeConf.BankCardLimitCount != 0 {
//	//		curday := time.Now().Format("20060102")
//	//		curtime, _ := time.ParseInLocation("20060102", curday, time.Local)
//	//		//每天5点到第二天5点
//	//		startSec := curtime.Add(5 * 3600 * time.Second).Unix()
//	//		endSec := curtime.Add(29 * 3600 * time.Second).Unix()
//	//		if time.Now().Hour() < 5 {
//	//			startSec = startSec - 24*3600
//	//			endSec = endSec - 24*3600
//	//		}
//	//
//	//		err := w.db.QueryRow(`select count(1) as cnt from
//	//		 exchange_t where apply_time between ? and ? and user_id=? and channel=1 and status<9`,
//	//			startSec, endSec, msg.GetUserID()).Scan(&count)
//	//		if err != nil {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.DBError
//	//			glog.SErrorf("query exchange_t failed.err:%v", err)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//	}
//	//
//	//	if conf.ExchangeConf.BankCardLimitAmount != 0 {
//	//		leftAmount := conf.ExchangeConf.BankCardLimitAmount - amount
//	//		if leftAmount < 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.BankLimitedError
//	//			glog.SErrorf("bankcard limitamount:%v usedamount:%v", conf.ExchangeConf.BankCardLimitAmount, amount)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//
//	//		// if leftAmount-amount < 0 {
//	//		// 	tx.Rollback()
//	//		// 	rMsg.RetCode = errcode.BankLimitingError
//	//		// 	glog.SErrorf("bankcard left amount:%v,exhangamount:%v", leftAmount, amount)
//	//		// 	s.Response(rMsg, mid)
//	//		// 	return nil
//	//		// }
//	//	}
//	//
//	//	if conf.ExchangeConf.BankCardLimitCount != 0 {
//	//		leftCount := conf.ExchangeConf.BankCardLimitCount - count
//	//		if leftCount <= 0 {
//	//			tx.Rollback()
//	//			rMsg.RetCode = errcode.BankCardCountError
//	//			glog.SErrorf("bankcard limit left count:%v", leftCount)
//	//			s.Response(rMsg, mid)
//	//			return nil
//	//		}
//	//	}
//	//}
//	//
//	//now := time.Now()
//	//orderid := strconv.Itoa(int(uid)) + now.Format("20060102150405")
//	//_, err = tx.Exec(`insert into exchange_t (device,channel,user_id, submit_money,ex_name,ex_account,apply_time,day_key,hour_key,status,machineid,ip,bankcode,orderid,telephone)
//	//values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, device, exType, uid, amount, accName, account, now.Unix(),
//	//	0, 0, 0, machinecode, ip, bankcode, orderid, tel)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert exchange failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//settlewid, err := models.InsertRecordAmount(uid, models.RecordRechargeApply, orderid, tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//_, err = tx.Exec(`update userwealth_t set wealth = wealth - ? where userid = ?`, amount, uid)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Update user wealth failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//
//	//err = models.UpdateRecordAmount(settlewid, uid, tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("update recordAmount failed for db.%v", err.Error())
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//var updateDWealth float64
//	//err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, uid).Scan(&updateDWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return s.Response(rMsg, mid)
//	//}
//	//rMsg.Wealth = updateDWealth
//	//glog.SInfof("C0000013 after exchange user  %v wealth is %v", uid, updateDWealth)
//	//err = tx.Commit()
//	//
//	////推送兑换申请邮件
//	//if err == nil {
//	//	mailInfo := &mail.CreateMailInfo{}
//	//	mailInfo.Type = mail.ExchangeApply
//	//	mailInfo.Amount = amount
//	//	mailInfo.MailToID = uid
//	//	mailInfo.Account = account
//	//	mailInfo.Name = accName
//	//	//0:支付宝 1:银行卡
//	//	switch exType {
//	//	case 0:
//	//		if strings.Contains(account, "@") {
//	//			mailInfo.AccType = mail.MailPay
//	//		} else {
//	//			mailInfo.AccType = mail.MobilePay
//	//		}
//	//	case 1:
//	//		mailInfo.AccType = mail.BankPay
//	//	}
//	//	if rMsg.RetCode == 0 {
//	//		mailtoID, err := mail.CreateHallMail(w.db, mailInfo)
//	//		push := &plr.P0000002{}
//	//		if err == nil {
//	//			push.UserID = mailtoID
//	//			s.Push("P0000002", push)
//	//		}
//	//	}
//	//}
//	//
//	//return s.Response(rMsg, mid)
//}
//
//// C0000005 用户绑定手机
//func (w *HallCore) C0000005(s *session.Session, msg *plr.C0000005, mid uint) error {
//	return nil
//	//w.login.Invoke(func() {
//	//	rMsg := w.login.BindUserMobile(msg)
//	//
//	//	mailInfo := &mail.CreateMailInfo{}
//	//	mailInfo.Type = mail.BindSuccess
//	//	mailInfo.MailToID = msg.UserID
//	//	if rMsg.RetCode == 0 {
//	//		mailtoID, err := mail.CreateHallMail(w.db, mailInfo)
//	//		push := &plr.P0000002{}
//	//		if err == nil {
//	//			push.UserID = mailtoID
//	//			s.Push("P0000002", push)
//	//		}
//	//	}
//	//	s.Response(rMsg, mid)
//	//})
//	//return nil
//}
//
//// C0000007 账户绑定支付包
//func (w *HallCore) C0000007(s *session.Session, msg *plr.C0000007, mid uint) error {
//	uid := msg.GetUserID()
//	alipayAcc := msg.GetAlipayAccount()
//	aliName := msg.GetAliPayNickName()
//
//	rMsg := &plr.S0000007{}
//
//	if alipayAcc == "" || aliName == "" {
//		glog.Errorln("Invalid params alipay")
//		rMsg.RetCode = errcode.InvalidParamError
//		return s.Response(rMsg, mid)
//	}
//
//	var count int32
//	err := w.db.QueryRow(`select count(1) from userextend_t where userid = ?`, uid).Scan(&count)
//	if err != nil {
//		glog.SErrorf("Update userextend_t info failed.%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	if count == 0 {
//		_, err = w.db.Exec(`INSERT INTO userextend_t (userid,alipay,alipaydetail,bankaccount,bankno,bankname,bankdetail) VALUE(?,"","","","","","")`, uid)
//		if err != nil {
//			glog.SErrorf("bindMobile failed %v", err)
//			rMsg.RetCode = errcode.DBError
//			return s.Response(rMsg, mid)
//		}
//	}
//
//	_, err = w.db.Exec(`update userextend_t set alipay = ?,alipaydetail = ?,bindalipaytime = ? where userid = ?`, aliName, alipayAcc, time.Now().Unix(), uid)
//	if err != nil {
//		glog.SErrorf("C0000007 update user alipay failed %v.", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	return s.Response(rMsg, mid)
//}
//
//// C0000008 账户绑定银行卡
//func (w *HallCore) C0000008(s *session.Session, msg *plr.C0000008, mid uint) error {
//	uid := msg.GetUserID()
//	bankAcc := msg.GetBankAccount()
//	bankNo := msg.GetBankCardNo()
//	bankcode := msg.GetBankCode()
//
//	rMsg := &plr.S0000008{}
//
//	if bankAcc == "" || bankNo == "" || bankcode == "" {
//		glog.SErrorf("Bind user bank card failed.Invalid params name:%v no:%v code:%v", bankAcc, bankNo, bankcode)
//		rMsg.RetCode = errcode.InvalidParamError
//		return s.Response(rMsg, mid)
//	}
//
//	var count int32
//	err := w.db.QueryRow(`select count(1) from userextend_t where userid = ?`, uid).Scan(&count)
//	if err != nil {
//		glog.SErrorf("Update bank info failed.%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	if count == 0 {
//		_, err = w.db.Exec(`INSERT INTO userextend_t (userid,alipay,alipaydetail,bankaccount,bankno,bankname,bankdetail) VALUE(?,"","","","","","")`, uid)
//		if err != nil {
//			glog.SErrorf("bindMobile failed %v", err)
//			rMsg.RetCode = errcode.DBError
//			return s.Response(rMsg, mid)
//		}
//	}
//
//	_, err = w.db.Exec(`update userextend_t set bankaccount = ?,bankno = ?,bindbanktime = ?,bankdetail=? where userid = ?`, bankAcc, bankNo, time.Now().Unix(), bankcode, uid)
//	if err != nil {
//		glog.SErrorf("Update bank info failed.%v", err)
//		rMsg.RetCode = errcode.DBError
//		return s.Response(rMsg, mid)
//	}
//
//	return s.Response(rMsg, mid)
//}
//
//// C0000015 手机登陆
//func (w *HallCore) C0000015(s *session.Session, msg *plr.C0000015, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.MobileLogin(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//// C0000016 获取验证码
//func (w *HallCore) C0000016(s *session.Session, msg *plr.C0000016, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.GenMobileCode(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}

// C0000018 Token登陆
func (w *HallCore) C0000018(s *session.Session, msg *plr.C0000018, mid uint) error {
	w.login.Invoke(func() {
		rMsg := w.login.TokenLogin(msg)
		s.Response(rMsg, mid)
	})
	return nil
}

//// C0000019 账户是否存在
//func (w *HallCore) C0000019(s *session.Session, msg *plr.C0000019, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.Check(msg.GetMobile(), msg.GetGameID())
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//// C0000020 重置密码
//func (w *HallCore) C0000020(s *session.Session, msg *plr.C0000020, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.ResetPassword(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//// C0000021 获取用户支付信息
//func (w *HallCore) C0000021(s *session.Session, msg *plr.C0000021, mid uint) error {
//	uid := msg.GetUserID()
//	rsp := &plr.S0000021{}
//	err := w.db.QueryRow(`select alipay,alipaydetail,bankaccount,bankno,bankdetail from userextend_t where userid = ?`, uid).Scan(&rsp.AliPayNickName,
//		&rsp.AlipayAccount, &rsp.BankAccount, &rsp.BankCardNo, &rsp.BankCode)
//	if err != nil {
//		glog.SErrorf("C0000021 %v", err)
//		rsp.RetCode = errcode.DBError
//	}
//	return s.Response(rsp, mid)
//}
//
//// C0000022 设置保险箱密码
//func (w *HallCore) C0000022(s *session.Session, msg *plr.C0000022, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.SetBankPassword(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//// C0000023 重置保险密码
//func (w *HallCore) C0000023(s *session.Session, msg *plr.C0000023, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.ResetBankPassword(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//// C0000026 查看邮件明细
//func (w *HallCore) C0000026(s *session.Session, msg *plr.C0000026, mid uint) error {
//	rMsg, _ := mail.ViewMailDetail(w.db, msg)
//	return s.Response(rMsg, mid)
//}
//
//// C0000025 根据用户查询邮件列表
//func (w *HallCore) C0000025(s *session.Session, msg *plr.C0000025, mid uint) error {
//	rMsg, _ := mail.QueryUserMail(w.db, *msg)
//	return s.Response(rMsg, mid)
//}
//
//// C0000024 苹果官方充值
//func (w *HallCore) C0000024(s *session.Session, msg *plr.C0000024, mid uint) error {
//	return nil
//	//go func() {
//	//	rsp := &plr.S0000024{}
//	//	receipt := msg.GetReceipt()
//	//	if ok, err := models.CheckAppleReceipt(receipt, w.db); !ok {
//	//		glog.SErrorf("CheckAppleReceipt failed :%v", err)
//	//		rsp.RetCode = errcode.AppleVerifyError
//	//		s.Response(rsp, mid)
//	//		return
//	//	}
//	//
//	//	client := appstore.New()
//	//	req := appstore.IAPRequest{
//	//		ReceiptData: receipt,
//	//	}
//	//	resp := &appstore.IAPResponse{}
//	//	err := client.Verify(req, resp)
//	//	if err != nil {
//	//		glog.SErrorf("AppleReceipt failed :%v", err)
//	//		rsp.RetCode = errcode.AppleVerifyError
//	//		s.Response(rsp, mid)
//	//		return
//	//	}
//	//
//	//	glog.SInfof("%+v", resp)
//	//
//	//	if amount, ok := conf.Conf.Products[msg.GetProductID()]; ok {
//	//		newAmount, err := models.RechargeUserCoin(msg.GetUserID(), amount, "", w.db)
//	//		if err != nil {
//	//			glog.SErrorf("Iap VerifyReceipt failed :%v", err)
//	//			rsp.RetCode = errcode.DBError
//	//			s.Response(rsp, mid)
//	//			return
//	//		}
//	//		rsp.CoinWealth = newAmount
//	//	}
//	//	models.InsertAppleReceipt(receipt, w.db)
//	//	s.Response(rsp, mid)
//	//}()
//	//return nil
//}
//
//// C0000027 兑换限额
//func (w *HallCore) C0000027(s *session.Session, msg *plr.C0000027, mid uint) error {
//	rsp := &plr.S0000027{}
//	var amount float64
//	var count int32
//	curday := time.Now().Format("20060102")
//	curtime, _ := time.ParseInLocation("20060102", curday, time.Local)
//	startSec := curtime.Unix()
//	endSec := curtime.Add(24 * 3600 * time.Second).Unix()
//
//	err := w.db.QueryRow(`select ifnull(sum(submit_money),0) as amount ,count(1) as cnt from
//	 exchange_t where apply_time between ? and ? and user_id=? and channel=0 and status<9`, startSec, endSec, msg.GetUserID()).Scan(&amount, &count)
//	if err != nil {
//		rsp.RetCode = errcode.DBError
//		glog.SErrorf("query exchange_t failed.err:%v", err)
//		s.Response(rsp, mid)
//		return nil
//	}
//
//	leftCount := conf.ExchangeConf.AlipayLimitCount - count
//	if leftCount < 0 {
//		leftCount = 0
//	}
//	leftAmount := conf.ExchangeConf.AlipayLimitAmount - amount
//	if leftAmount < 0 {
//		leftAmount = 0
//	}
//	exchangType := make([]int32, 0)
//	if conf.ExchangeConf.AliPayExchange {
//		exchangType = append(exchangType, 0)
//	}
//	if conf.ExchangeConf.BankCardExchange {
//		exchangType = append(exchangType, 1)
//	}
//
//	var bankamount float64
//	var bankcount int32
//	startSec = curtime.Add(5 * 3600 * time.Second).Unix()
//	endSec = curtime.Add(29 * 3600 * time.Second).Unix()
//	if time.Now().Hour() < 5 {
//		startSec = startSec - 24*3600
//		endSec = endSec - 24*3600
//	}
//	err = w.db.QueryRow(`select ifnull(sum(submit_money),0) as amount ,count(1) as cnt from
//	exchange_t where apply_time between ? and ? and user_id=? and channel=1 and status<9`, startSec, endSec, msg.GetUserID()).Scan(&bankamount, &bankcount)
//	if err != nil {
//		rsp.RetCode = errcode.DBError
//		glog.SErrorf("query exchange_t failed.err:%v", err)
//		s.Response(rsp, mid)
//		return nil
//	}
//
//	bankLeftCount := conf.ExchangeConf.BankCardLimitCount - bankcount
//	if bankLeftCount <= 0 {
//		bankLeftCount = 0
//	}
//	bankLeftAmount := conf.ExchangeConf.BankCardLimitAmount - bankamount
//	if bankLeftAmount < 0 {
//		bankLeftAmount = 0
//	}
//	rsp.BankCardLimitAmount = conf.ExchangeConf.BankCardLimitAmount
//	rsp.BankCardLeftAmount = bankLeftAmount
//	rsp.BankCardLimitCount = conf.ExchangeConf.BankCardLimitCount
//	rsp.BankCardLeftCount = bankLeftCount
//
//	rsp.LimitAmount = conf.ExchangeConf.AlipayLimitAmount
//	rsp.LeftAmount = leftAmount
//	rsp.LimitCount = conf.ExchangeConf.AlipayLimitCount
//	rsp.LeftCount = leftCount
//	rsp.ExchangeType = exchangType
//	return s.Response(rsp, mid)
//}
//
//// C0000028 找回账号
//func (w *HallCore) C0000028(s *session.Session, msg *plr.C0000028, mid uint) error {
//	w.login.Invoke(func() {
//		sMsg := w.login.SearchMachineUser(msg)
//		s.Response(sMsg, mid)
//	})
//	return nil
//}
//
//// C1070007 获取支付开关
//func (w *HallCore) C1070007(s *session.Session, msg *plr.C1070007, mid uint) error {
//	rsp := &plr.S1070007{}
//
//	err := w.db.QueryRow(`SELECT pay_switch FROM global_config_t`).Scan(&rsp.AuditStatus)
//	if err != nil {
//		glog.SErrorf("C1070007 failed.db error:%v", err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//
//	var chinese, foreign sql.NullString
//	var updatechinese, updateforeign sql.NullString
//	err = w.db.QueryRow(`SELECT website,gamesite_chinese,gamesite_foreign,updatesite,updatesite_foreign FROM address_t where game_id=?`, msg.GetGameID()).Scan(&rsp.Website, &chinese, &foreign, &updatechinese, &updateforeign)
//	if err != nil {
//		glog.SErrorf("C1070007 failed.db error:%v", err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//
//	rsp.Gamesite = ip.GetGamesite2(msg.GetIPAddr(), chinese.String, foreign.String)
//	rsp.Updatesite = ip.GetGamesite2(msg.GetIPAddr(), updatechinese.String, updateforeign.String)
//	return s.Response(rsp, mid)
//}
//
//// C0000029 获取玩家兑换记录
//func (w *HallCore) C0000029(s *session.Session, msg *plr.C0000029, mid uint) error {
//	uid := msg.GetUserID()
//	machineid := msg.GetMachineID()
//	rsp := &plr.S0000029{}
//	rows, err := w.db.Query(`SELECT apply_time,submit_money,channel,status FROM exchange_t where machineid = ? and user_id = ?`, machineid, uid)
//	if err != nil {
//		glog.SErrorf("exchange failed.err:%v", err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//	defer rows.Close()
//	for rows.Next() {
//		v := &plr.S0000029_Exchange{}
//		var (
//			applytime   int64
//			submitmoney float64
//			channel     int32
//			status      int32
//		)
//		err = rows.Scan(&applytime, &submitmoney, &channel, &status)
//		if err != nil {
//			glog.SErrorf("exchange failed.err:%v", err)
//			return err
//		}
//		tm := time.Unix(applytime, 0)
//		v.ApplyTime = tm.Format("2006-01-02 03:04:05")
//		v.ExchangeAmount = submitmoney
//		v.ExchangeType = getchannelstring(channel)
//		v.Status = getStatusStr(status)
//		rsp.ExchangeList = append(rsp.ExchangeList, v)
//	}
//	return s.Response(rsp, mid)
//}
//
//// C0000030 获取VIP等级
//func (w *HallCore) C0000030(s *session.Session, msg *plr.C0000030, mid uint) error {
//	return nil
//	//rsp := &plr.S0000030{}
//	//uid := msg.GetUserID()
//	//var frameid int32
//	//err := w.db.QueryRow(`select faceframeid from account_t where userid=?`, uid).Scan(&frameid)
//	//if err != nil {
//	//	glog.SErrorf("query frameid failed.uid:%v,err:%v", uid, err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//viplvl, vipvalue, err := models.QueryVipLevel(uid, w.db)
//	//if err != nil {
//	//	glog.SErrorf("query vipLevel failed.uid:%v,err:%v", uid, err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//rsp.VipLevel = viplvl
//	//rsp.VipValue = vipvalue
//	//rsp.FaceFrameID = frameid
//	//return s.Response(rsp, mid)
//}

// C0000031 获取服务器维护状态
//func (w *HallCore) C0000031(s *session.Session, msg *plr.C0000031, mid uint) error {
//	rsp := &plr.S0000031{}
//	var serverStopTime sql.NullInt64
//	err := w.db.QueryRow(`select start_time from stop_server_t`).Scan(&serverStopTime)
//	if err != nil && err != sql.ErrNoRows {
//		glog.SErrorf("query stop_server_t failed.err:%v", err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//	if serverStopTime.Valid {
//		inttime := serverStopTime.Int64
//		if inttime > time.Now().Unix() {
//			rsp.ServerStopTime = inttime
//		}
//	}
//	rsp.ServerState = int32(control.ServerState)
//	return s.Response(rsp, mid)
//}

//// C0000032 获取白名单状态
//func (w *HallCore) C0000032(s *session.Session, msg *plr.C0000032, mid uint) error {
//	rsp := &plr.S0000032{}
//	count := 0
//	err := w.db.QueryRow(`select count(*) from stop_white_list_t w join account_t a
//	on a.userid=w.user_id where machineid=?`, msg.GetMachineID()).Scan(&count)
//	if err != nil {
//		glog.SErrorf("query stop_white_list_t failed.machineid:%v,err:%v", msg.GetMachineID(), err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//	if count > 0 {
//		rsp.Status = 1
//	}
//	return s.Response(rsp, mid)
//}
//
//// C0000033 邮件领取金币
//func (w *HallCore) C0000033(s *session.Session, msg *plr.C0000033, mid uint) error {
//	return nil
//	//rsp := &plr.S0000033{}
//	//userid := msg.GetUserID()
//	//mailid := msg.GetMailId()
//	//
//	//tx, err := w.db.Begin()
//	//if err != nil {
//	//	glog.SErrorf("begin transaction failed,err:%v", err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//var uamount float64
//	//var uamountstatus int32
//	//err = tx.QueryRow(`select amountstatus, amount from hall_mail_t where mail_id=? and mail_to_id=?`, mailid, userid).Scan(&uamountstatus, &uamount)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("C0000033 failed.query hall_mail_t error mailid:%v,uid:%v,err:%v", mailid, userid, err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//if uamount <= 0 {
//	//	tx.Rollback()
//	//	glog.SErrorf("C0000033 failed.invalid mail amount:%v", uamount)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//if uamountstatus != mail.AmountUnReceive {
//	//	tx.Rollback()
//	//	glog.SErrorf("C0000033 failed.invalid amountstatus:%v", uamountstatus)
//	//	rsp.RetCode = errcode.InvalidParamError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//var userWealth float64
//	//err = tx.QueryRow(`SELECT wealth FROM userwealth_t WHERE userid=?`, userid).Scan(&userWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("query user wealth failed.playerUID%v err:%v", userid, err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//_, err = tx.Exec(`update hall_mail_t set amountstatus=? where mail_id=?`, mail.AmountReceived, msg.GetMailId())
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("update hallMail failed.err:%v msg:%v", err, msg.GetMailId())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//settlewid, err := models.InsertRecordAmount(userid, models.RecordMailCoin, "", tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//_, err = tx.Exec(`UPDATE userwealth_t SET wealth=wealth+? WHERE userid=?`, uamount, userid)
//	//if err != nil {
//	//	tx.Rollback()
//	//	rsp.RetCode = errcode.DBError
//	//	glog.SErrorf("update user wealth failed.playerUID:%v wealth:%v,err:%v", userid, uamount, err)
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//err = models.UpdateRecordAmount(settlewid, userid, tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("update recordAmount failed for db.%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//_, err = tx.Exec(`update hall_mail_t set amountstatus=? where mail_id=?`, mail.AmountReceived, msg.GetMailId())
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("update hallMail failed.err:%v msg:%v", err, msg.GetMailId())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//var currWealth float64
//	//err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, userid).Scan(&currWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("DB failed.%v", err)
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//rsp.Wealth = currWealth
//	//tx.Commit()
//	//
//	//return s.Response(rsp, mid)
//}
//
//// C0000035 盈利榜
//func (w *HallCore) C0000035(s *session.Session, msg *plr.C0000035, mid uint) error {
//	return nil
//	//rsp := &plr.S0000035{}
//	//uid := msg.GetUserID()
//	//nowday := time.Now().Format("20060102")
//	//var odds = 1.0
//	//if len(w.luckyodds) == 0 {
//	//	w.luckyodds[nowday] = 3 + rand.Float64()
//	//	odds = w.luckyodds[nowday]
//	//} else {
//	//	k, isok := w.luckyodds[nowday]
//	//	if isok {
//	//		odds = k
//	//	} else {
//	//		w.luckyodds = make(map[string]float64)
//	//		w.luckyodds[nowday] = 3 + rand.Float64()
//	//		odds = w.luckyodds[nowday]
//	//	}
//	//}
//	//uif, err := models.QueryUserInfo(uid, w.db)
//	//if err != nil {
//	//	glog.SErrorf("query userinfo failed.err:%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//gid := uif.GameID
//	//
//	//var topn int32 = 50
//	//lc := fmt.Sprintf("%%%v%%", gid)
//	//rs, err := w.db.Query(`select  userid,nickname,profit from
//	//((select  l.userid,nickname,case when l.userid<682500 then  truncate(profit*?,3) else  profit end as profit from lucky_list_t l join account_t a on l.userid=a.userid
//	//where (gameids like ?
//	//or (l.userid>=682500 and gameids='')
//	//) and profit>0
//	//order by profit desc
//	//limit ?)
//	//union
//	//select l.userid, nickname,profit from lucky_list_t l join account_t a on l.userid=a.userid
//	//where l.userid=?
//	//) as t
//	//order by profit desc,nickname
//	//`, odds, lc, topn, uid)
//	//if err != nil {
//	//	glog.SErrorf("query Platform map failed. err:%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//defer rs.Close()
//	//rlist := make([]*plr.S0000035_TopProfit, 0)
//	//
//	//var rn int32
//	//slf := &plr.S0000035_TopProfit{}
//	//for rs.Next() {
//	//	r := &plr.S0000035_TopProfit{}
//	//	var ruid int32
//	//	err := rs.Scan(&ruid, &r.NickName, &r.Amount)
//	//	if err != nil {
//	//		glog.SErrorf("scan luckylist failed. err:%v", err.Error())
//	//		continue
//	//	}
//	//	rn++
//	//	r.TopN = rn
//	//	if ruid == uid {
//	//		slf = r
//	//	}
//	//	if rn <= topn {
//	//		if r.Amount > 0 {
//	//			rlist = append(rlist, r)
//	//		}
//	//	}
//	//}
//	//
//	//if slf.TopN > topn || (slf.TopN == rn && slf.Amount < 0) {
//	//	slf.TopN = 0
//	//}
//	////玩家未下注时sql查询无自己的数据 这里补上
//	//if slf.NickName == "" {
//	//	slf.NickName = uif.NickName
//	//}
//	//rlist = append(rlist, slf)
//	//rsp.TopProfits = rlist
//	//return s.Response(rsp, mid)
//}
//
//// C0000036 幸运榜
//func (w *HallCore) C0000036(s *session.Session, msg *plr.C0000036, mid uint) error {
//	return nil
//	//rsp := &plr.S0000036{}
//	//uid := msg.GetUserID()
//	//
//	//uif, err := models.QueryUserInfo(uid, w.db)
//	//if err != nil {
//	//	glog.SErrorf("query userinfo failed.err:%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//gid := uif.GameID
//	//
//	//var topn int32 = 50
//	//lc := fmt.Sprintf("%%%v%%", gid)
//	//rs, err := w.db.Query(`select  userid,nickname,luckyprofit,gamekindid,luckytime from
//	//((select  l.userid,nickname,luckyprofit,gamekindid,luckytime from lucky_list_t l join account_t a on l.userid=a.userid
//	//where (gameids like ?
//	//or (l.userid>=682500 and gameids='')
//	//) and luckyprofit>0
//	//order by luckyprofit desc
//	//limit ?)
//	//union
//	//select l.userid, nickname,luckyprofit,gamekindid,luckytime from lucky_list_t l join account_t a on l.userid=a.userid
//	//where l.userid=?
//	//) as t
//	//order by luckyprofit desc,luckytime`, lc, topn, uid)
//	//if err != nil {
//	//	glog.SErrorf("query Platform map failed. err:%v", err.Error())
//	//	rsp.RetCode = errcode.DBError
//	//	return s.Response(rsp, mid)
//	//}
//	//
//	//defer rs.Close()
//	//rlist := make([]*plr.S0000036_TopLucky, 0)
//	//
//	//var rn int32
//	//slf := &plr.S0000036_TopLucky{}
//	//for rs.Next() {
//	//	r := &plr.S0000036_TopLucky{}
//	//	var ruid int32
//	//	err := rs.Scan(&ruid, &r.NickName, &r.Amount, &r.KindID, &r.LuckyTime)
//	//	if err != nil {
//	//		glog.SErrorf("scan luckylist failed. err:%v", err.Error())
//	//		continue
//	//	}
//	//	rn++
//	//	r.TopN = rn
//	//	if ruid == uid {
//	//		slf = r
//	//	}
//	//	if rn <= topn {
//	//		if r.Amount > 0 {
//	//			rlist = append(rlist, r)
//	//		}
//	//	}
//	//}
//	//
//	//if slf.TopN > topn || (slf.TopN == rn && slf.Amount < 0) {
//	//	slf.TopN = 0
//	//}
//	////玩家未下注时sql查询无自己的数据 这里补上
//	//if slf.NickName == "" {
//	//	slf.NickName = uif.NickName
//	//}
//	//rlist = append(rlist, slf)
//	//rsp.TopLuckys = rlist
//	//return s.Response(rsp, mid)
//}
//
//// C0000037 VIP等级配置
//func (w *HallCore) C0000037(s *session.Session, msg *plr.C0000037, mid uint) error {
//	rsp := &plr.S0000037{}
//	cf, err := models.QueryVipConfig(w.db)
//	if err != nil {
//		glog.SErrorf("query vipconfig failed.err:%v", err)
//		rsp.RetCode = errcode.DBError
//		return s.Response(rsp, mid)
//	}
//
//	if cf != nil {
//		for _, v := range cf {
//			cfg := &plr.S0000037_VIPConfig{
//				VipLvl:       v.VipLvl,
//				VipValue:     v.VipValue,
//				FaceFrameids: v.FaceFrameids,
//			}
//			rsp.Configs = append(rsp.Configs, cfg)
//		}
//	}
//	return s.Response(rsp, mid)
//}
//
//// C1010027 后台获取验证码
//func (w *HallCore) C1010027(s *session.Session, msg *explr.C1010027, mid uint) error {
//	w.login.Invoke(func() {
//		rMsg := w.login.GenMobileCodeByAdmin(msg)
//		s.Response(rMsg, mid)
//	})
//	return nil
//}
//
//func getchannelstring(channel int32) string {
//	switch channel {
//	case 0:
//		return "支付宝"
//	default:
//		return "银行卡"
//	}
//}
//
//func getStatusStr(status int32) string {
//	switch status {
//	case 0:
//		return "已提交"
//	case 1:
//		fallthrough
//	case 2:
//		fallthrough
//	case 3:
//		fallthrough
//	case 4:
//		fallthrough
//	case 5:
//		fallthrough
//	case 6:
//		fallthrough
//	case 7:
//		return "兑换中"
//	case 8:
//		return "兑换成功"
//	case 9:
//		return "兑换失败"
//	case 10:
//		return "退还"
//	case 11:
//		fallthrough
//	case 12:
//		fallthrough
//	case 13:
//		fallthrough
//	default:
//		return "订单异常"
//	}
//}
