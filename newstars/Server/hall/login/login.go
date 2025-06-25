package login

import (
	"context"
	"database/sql"
	"encoding/json"
	"newstars/Protocol/plr"
	"newstars/Server/hall/errcode"
	"newstars/framework/game_center"
	"newstars/framework/glog"
	"newstars/framework/model/cachekey"
	"newstars/framework/model/data"
	"newstars/framework/redisx"
	"runtime"
)

type UserLogin struct {
	db         *sql.DB
	gid        int64
	uids       map[string]string
	chFunction chan func()
	quit       chan int
}

func NewUserLogin(db *sql.DB) *UserLogin {
	return &UserLogin{
		db:         db,
		uids:       make(map[string]string),
		gid:        500,
		chFunction: make(chan func(), 1024),
		quit:       make(chan int),
	}
}

// Go run
func (p *UserLogin) Go() {
	go p.run()
}

// Invoke do in goroutine
func (p *UserLogin) Invoke(fn func()) {
	p.chFunction <- fn
}

func pinvoke(fn func()) {

	defer func() {
		if err := recover(); err != nil {
			glog.SErrorf("invoke failed: %v stack:%v", err, stack())
		}
	}()

	fn()
}

func stack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)
	return s
}

func (p *UserLogin) run() {

	// ticker := time.NewTicker(300 * time.Millisecond)

	for {
		select {
		case fn := <-p.chFunction:
			pinvoke(fn)
		case <-p.quit:
			return
		}
	}
}

func (p *UserLogin) TokenLogin(msg *plr.C0000018) *plr.S0000018 {
	sMsg := &plr.S0000018{}

	if len(msg.GetToken()) == 0 {
		glog.SErrorf("TokenLogin PasswordError.%v", msg.GetToken())
		sMsg.RetCode = errcode.PasswordError
		return sMsg
	}

	resp, err := game_center.VerifyCustomToken(msg.GetToken(), []byte(data.BaseConfig.SelfGameJwt))
	if err != nil {
		glog.SErrorf("Failed to VerifyCustomToken. token:%v err:%v", msg.GetToken(), err)
		sMsg.RetCode = errcode.PasswordError
		return sMsg
	}

	uid := resp.Token

	p.uids[uid] = msg.GetAccountName()

	marshal, _ := json.Marshal(resp)
	redisx.Set(context.Background(), cachekey.GetUserKey(uid), marshal, -1)

	sMsg.RetCode = errcode.CodeOK
	sMsg.UserID = uid
	sMsg.HeatTimes = HeatTime
	sMsg.Token = msg.Token
	return sMsg
}

// // GuestLogin for guest
//
//	func (p *UserLogin) GuestLogin(msg *plr.C0000004) *plr.S0000004 {
//		var accName, accPassword string
//		var acctype, platformid int32
//
//		sMsg := &plr.S0000004{}
//
//		if msg.GetMachineID() == "" {
//			glog.SErrorf("GuestLogin Failed.MachineID is nil")
//			sMsg.RetCode = errcode.InvalidParamError
//			return sMsg
//		}
//
//		if msg.GetIPAddr() == "" {
//			glog.SErrorf("IP为空,machineID:%v", msg.GetMachineID())
//			sMsg.RetCode = errcode.IPEmptyIP
//			return sMsg
//		}
//
//		if ip.CheckBlackRegion(msg.IPAddr) {
//			glog.SErrorf("IP地址属于黑名单地区，IP：%s", msg.IPAddr)
//			sMsg.RetCode = errcode.InvalidParamError
//			return sMsg
//		}
//
//		// currhour := time.Now().Hour()
//		// if p.currHour != currhour {
//		// 	delIps := make([]string, 0)
//		// 	for k, v := range p.ipcount {
//		// 		if time.Now().Unix() > v.expiredtime {
//		// 			if v.count >= conf.Conf.IPLimitCount {
//		// 				if time.Now().Unix() > v.unlocktime {
//		// 					delIps = append(delIps, k)
//		// 				}
//		// 			} else {
//		// 				delIps = append(delIps, k)
//		// 			}
//		// 		}
//		// 	}
//		// 	for _, dip := range delIps {
//		// 		delete(p.ipcount, dip)
//		// 	}
//		// 	p.currHour = currhour
//		// }
//
//		//如果超过6万条，保留一半最新的数据
//		// var keeptimes int64
//		// if len(p.ipcount) > 60000 {
//		// 	iparr := make([]int64, 0)
//		// 	for _, v := range p.ipcount {
//		// 		iparr = append(iparr, v.expiredtime)
//		// 	}
//		// 	sort.Slice(iparr, func(i, j int) bool {
//		// 		return iparr[i] < iparr[j]
//		// 	})
//		// 	keeptimes = iparr[len(iparr)/2]
//		// }
//
//		// if keeptimes != 0 {
//		// 	newipcount := make(map[string]IPCount)
//		// 	for k, v := range p.ipcount {
//		// 		if v.expiredtime > keeptimes {
//		// 			newipcount[k] = v
//		// 		}
//		// 	}
//		// 	glog.Warning("GuestLogin 清理前ip个数:%v", len(p.ipcount))
//		// 	p.ipcount = newipcount
//		// 	glog.Warning("GuestLogin 清理后ip个数:%v", len(p.ipcount))
//		// }
//
//		// ip := msg.GetIPAddr()
//		// ipc, isok := p.ipcount[ip]
//		// if isok {
//		// 	if time.Now().Unix() > ipc.expiredtime {
//		// 		if time.Now().Unix() > ipc.unlocktime && ipc.unlocktime != 0 {
//		// 			ipconf := IPCount{
//		// 				count:       0,
//		// 				expiredtime: time.Now().Unix() + conf.Conf.IPCheckInterval,
//		// 			}
//		// 			p.ipcount[ip] = ipconf
//		// 		}
//		// 	}
//
//		// } else {
//		// 	ipconf := IPCount{
//		// 		count:       0,
//		// 		expiredtime: time.Now().Unix() + conf.Conf.IPCheckInterval,
//		// 	}
//		// 	p.ipcount[ip] = ipconf
//		// }
//
//		// if p.ipcount[ip].count >= conf.Conf.IPLimitCount {
//		// 	if p.ipcount[ip].unlocktime == 0 {
//		// 		ipconf := IPCount{
//		// 			count:       p.ipcount[ip].count,
//		// 			expiredtime: p.ipcount[ip].expiredtime,
//		// 			unlocktime:  time.Now().Unix() + conf.Conf.IPUnlockInterval,
//		// 		}
//		// 		p.ipcount[ip] = ipconf
//		// 	}
//		// 	glog.SErrorf("IP注册量过多:%s,限制量:%v 当前值:%+v", ip, conf.Conf.IPLimitCount, p.ipcount[ip])
//		// 	sMsg.RetCode = errcode.IPLimit
//		// 	return sMsg
//		// }
//
//		text := fmt.Sprintf("{%v:%v:%v:%v}", msg.GetMachineID(), msg.GetPlatFormID(), msg.GetIPAddr(), guestToken)
//		sign := util.Md5Hash(text)
//		if msg.GetSign() != sign {
//			glog.SErrorf("GuestLogin Failed.sign is error")
//			sMsg.RetCode = errcode.InvalidParamError
//			return sMsg
//		}
//
//		tx, err := p.db.Begin()
//		if err != nil {
//			glog.SErrorf("DB Error.Begin error:%v", err)
//			sMsg.RetCode = errcode.DBError
//			return sMsg
//		}
//
//		err = tx.QueryRow(`SELECT accountname,password,acctype,platformid FROM account_t WHERE machineid = ? AND acctype = 1 AND game_id = ?`,
//			msg.GetMachineID(), msg.GetGameID()).Scan(&accName, &accPassword, &acctype, &platformid)
//
//		switch {
//		case err == sql.ErrNoRows:
//			acc := p.buildGuest(msg.GetPlatFormID())
//			acc.machineid = msg.GetMachineID()
//			acc.ipaddr = msg.GetIPAddr()
//			loc, err := ip17mon.Find(acc.ipaddr)
//			if err == nil {
//				acc.iparea = loc.String()
//			} else {
//				glog.SWarnf("Can find user ip addr.error:%v", err)
//				// index := rand.Intn(len(models.Area))
//				// acc.iparea = models.Area[index]
//				acc.iparea = "未知"
//			}
//			acc.platformid = msg.GetPlatFormID()
//			acc.gameid = msg.GetGameID()
//			err = p.insertAcc(acc, tx)
//			if err != nil {
//				tx.Rollback()
//				sMsg.RetCode = errcode.DBError
//				return sMsg
//			}
//			sMsg.RetCode = errcode.CodeOK
//			sMsg.GuestName = acc.accountname
//			sMsg.GuestPassword = acc.password
//		case err != nil:
//			glog.SErrorf("GuestLogin Failed.Query error:%v", err)
//			tx.Rollback()
//			sMsg.RetCode = errcode.DBError
//			return sMsg
//		default:
//			if acctype == GuestType {
//				sMsg.GuestName = accName
//				sMsg.GuestPassword = accPassword
//				sMsg.RetCode = errcode.CodeOK
//			} else {
//				sMsg.RetCode = errcode.InvalidParamError
//			}
//
//		}
//		tx.Commit()
//
//		//ip注册用户计数
//		// ipc, isok = p.ipcount[ip]
//		// if isok {
//		// 	ipconf := IPCount{
//		// 		count:       ipc.count + 1,
//		// 		expiredtime: ipc.expiredtime,
//		// 		unlocktime:  ipc.unlocktime,
//		// 	}
//		// 	p.ipcount[ip] = ipconf
//		// }
//
//		return sMsg
//	}
//
// // AccountLogin for login
//
//	func (p *UserLogin) AccountLogin(msg *plr.C0000001) *plr.S0000001 {
//		sMsg := &plr.S0000001{}
//		var (
//			id          string
//			platformid  int32
//			accountname string
//			mobile      string
//			password    string
//			status      int32
//			acctype     int32
//		)
//		if msg.GetAccountName() == "" || msg.GetMachineID() == "" || msg.GetPassword() == "" {
//			glog.SErrorf("AccountLogin Failed.Invalid params:%+v", msg)
//			sMsg.RetCode = errcode.InvalidParamError
//			return sMsg
//		}
//
//		err := p.db.QueryRow(`SELECT userid,accountname,mobile,password,status,platformid,acctype FROM account_t WHERE machineid = ? and game_id = ?`, msg.GetMachineID(), msg.GetGameID()).Scan(&id, &accountname,
//			&mobile, &password, &status, &platformid, &acctype)
//
//		if err == sql.ErrNoRows {
//			glog.SErrorf("AccountLogin MachineIDNotExsit %v", msg.GetMachineID())
//			sMsg.RetCode = errcode.MachineIDNotExsit
//			return sMsg
//		}
//
//		if err != nil {
//			glog.SErrorf("AccountLogin error:%v", err)
//			sMsg.RetCode = errcode.DBError
//			return sMsg
//		}
//
//		if accountname != msg.GetAccountName() && acctype == GuestType {
//			glog.SErrorf("AccountLogin UserNameError.%v", msg.GetAccountName())
//			sMsg.RetCode = errcode.UserNameError
//			return sMsg
//		}
//
//		if mobile != msg.GetAccountName() && acctype == AccountType {
//			glog.SErrorf("AccountLogin UserNameError.%v", msg.GetAccountName())
//			sMsg.RetCode = errcode.UserNameError
//			return sMsg
//		}
//
//		if password != msg.GetPassword() {
//			glog.SErrorf("AccountLogin PasswordError.%v", msg.GetPassword())
//			sMsg.RetCode = errcode.PasswordError
//			return sMsg
//		}
//
//		if platformid != msg.GetPlatFormID() {
//			glog.SErrorf("AccountLogin PlatformError.%v", msg.GetPlatFormID())
//			sMsg.RetCode = errcode.PlatformError
//			return sMsg
//		}
//
//		if status == StatusForbid {
//			glog.SErrorf("AccountLogin AccountForbid.%v", status)
//			sMsg.RetCode = errcode.AccountForbid
//			return sMsg
//		}
//
//		if _, ok := p.uids[id]; ok {
//			glog.SErrorf("AccountLogin Failed.Mutiple %v", id)
//			sMsg.RetCode = errcode.MutilLoginError
//			return sMsg
//		}
//
//		_, err = p.db.Exec(`UPDATE account_t SET logintime = ? , status = ? WHERE userid = ?`, time.Now().Unix(), StatusLogin, id)
//		if err != nil {
//			glog.SErrorf("AccountLogin Failed.%+v update account error:%v", msg, err)
//			sMsg.RetCode = errcode.DBError
//			return sMsg
//		}
//
//		p.uids[id] = msg.GetAccountName()
//
//		sMsg.RetCode = errcode.CodeOK
//		sMsg.UserID = id
//		sMsg.HeatTimes = HeatTime
//		return sMsg
//	}
//
// // AccountLogin2 not bind machineid
//
//	func (p *UserLogin) AccountLogin2(msg *plr.C0000001) *plr.S0000001 {
//		return nil
//		//sMsg := &plr.S0000001{}
//		//var (
//		//	id int32
//		//	// platformid  int32
//		//	accountname string
//		//	mobile      string
//		//	password    string
//		//	status      int32
//		//	acctype     int32
//		//)
//		//
//		//if msg.GetAccountName() == "" || msg.GetPassword() == "" {
//		//	glog.SErrorf("AccountLogin Failed.Invalid params:%+v", msg)
//		//	sMsg.RetCode = errcode.InvalidParamError
//		//	return sMsg
//		//}
//		//
//		//err := p.db.QueryRow(`SELECT userid,accountname,mobile,password,status,acctype FROM account_t WHERE (accountname = ? OR mobile = ?) and game_id = ?`, msg.GetAccountName(),
//		//	msg.GetAccountName(), msg.GetGameID()).Scan(&id, &accountname, &mobile, &password, &status, &acctype)
//		//
//		//if err == sql.ErrNoRows {
//		//	glog.SErrorf("AccountLogin UserNameError %v", msg.GetAccountName())
//		//	sMsg.RetCode = errcode.UserNameError
//		//	return sMsg
//		//}
//		//
//		//if err != nil {
//		//	glog.SErrorf("AccountLogin error:%v", err)
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//
//		//if !control.CanLogin(id, p.db) {
//		//	glog.SErrorf("AccountLogin Failed.Server is %v cannot login.", control.ServerState)
//		//	sMsg.RetCode = errcode.ServerStopError
//		//	return sMsg
//		//}
//		//
//		//if accountname != msg.GetAccountName() && acctype == GuestType {
//		//	glog.SErrorf("AccountLogin UserNameError.%v", msg.GetAccountName())
//		//	sMsg.RetCode = errcode.UserNameError
//		//	return sMsg
//		//}
//		//
//		//if mobile != msg.GetAccountName() && acctype == AccountType {
//		//	glog.SErrorf("AccountLogin UserNameError.%v", msg.GetAccountName())
//		//	sMsg.RetCode = errcode.UserNameError
//		//	return sMsg
//		//}
//		//
//		//if acctype == AccountType && password != util.EncryptPassword(msg.GetPassword(), Salt) {
//		//	glog.SErrorf("AccountLogin PasswordError.%v", msg.GetPassword())
//		//	sMsg.RetCode = errcode.PasswordError
//		//	return sMsg
//		//}
//		//
//		//if acctype == GuestType && password != msg.GetPassword() {
//		//	glog.SErrorf("AccountLogin PasswordError.%v", msg.GetPassword())
//		//	sMsg.RetCode = errcode.PasswordError
//		//	return sMsg
//		//}
//		//
//		//// if platformid != msg.GetPlatFormID() {
//		//// 	glog.SErrorf("AccountLogin PlatformError.%v", msg.GetPlatFormID())
//		//// 	sMsg.RetCode = errcode.PlatformError
//		//// 	return sMsg
//		//// }
//		//
//		//if status == StatusForbid {
//		//	glog.SErrorf("AccountLogin AccountForbid.%v", status)
//		//	sMsg.RetCode = errcode.AccountForbid
//		//	return sMsg
//		//}
//		//
//		//if _, ok := p.uids[id]; ok {
//		//	glog.SWarnf("AccountLogin Failed.Mutiple %v", id)
//		//	// sMsg.RetCode = errcode.MutilLoginError
//		//	// return sMsg
//		//	// push := &plr.P1000008{}
//		//	// push.UserID = id
//		//	// s.Push("P1000008", push)
//		//}
//		//
//		//token := p.genPassword(accountname)
//		//
//		//_, err = p.db.Exec(`UPDATE account_t SET ipaddr = ? ,logintime = ? , status = ?, token = ? WHERE userid = ?`, msg.GetIPAddr(), time.Now().Unix(), StatusLogin, token, id)
//		//if err != nil {
//		//	glog.SErrorf("AccountLogin Failed.%+v update account error:%v", msg, err)
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//
//		//_, err = p.db.Exec(`INSERT INTO login_info_t (userid,optime,type,ip,iparea,terminaltype,osvesion) VALUES(?,?,?,?,?,?,?)`, id, time.Now().Unix(), LoginType, msg.GetIPAddr(), GetIPAddr(msg.GetIPAddr()), msg.GetTerminalType(), msg.GetClientVersion())
//		//if err != nil {
//		//	glog.SErrorf("AccountLogin Failed.%+v update account error:%v", msg, err)
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//
//		//p.UpdatePhoneInfo(id, msg.GetModel(), msg.GetVersion())
//		//
//		//p.uids[id] = msg.GetAccountName()
//		//
//		//sMsg.RetCode = errcode.CodeOK
//		//sMsg.UserID = id
//		//sMsg.HeatTimes = HeatTime
//		//sMsg.Token = token
//		//return sMsg
//	}
//
// // TokenLogin for token
// //func (p *UserLogin) TokenLogin(msg *plr.C0000018) *plr.S0000018 {
// //	sMsg := &plr.S0000018{}
// //	var (
// //		id int32
// //		// platformid  int32
// //		accountname string
// //		mobile      string
// //		token       string
// //		status      int32
// //		acctype     int32
// //	)
// //	if msg.GetAccountName() == "" || msg.GetToken() == "" {
// //		glog.SErrorf("TokenLogin Failed.Invalid params:%+v", msg)
// //		sMsg.RetCode = errcode.InvalidParamError
// //		return sMsg
// //	}
// //
// //	err := p.db.QueryRow(`SELECT userid,accountname,mobile,token,status,acctype FROM account_t WHERE (accountname = ? OR mobile = ?) AND game_id = ?`, msg.GetAccountName(),
// //		msg.GetAccountName(), msg.GetGameID()).Scan(&id, &accountname, &mobile, &token, &status, &acctype)
// //
// //	if err == sql.ErrNoRows {
// //		glog.SErrorf("TokenLogin UserNameError %v", msg.GetAccountName())
// //		sMsg.RetCode = errcode.UserNameError
// //		return sMsg
// //	}
// //
// //	if err != nil {
// //		glog.SErrorf("TokenLogin error:%v", err)
// //		sMsg.RetCode = errcode.DBError
// //		return sMsg
// //	}
// //
// //	if !control.CanLogin(id, p.db) {
// //		glog.SErrorf("AccountLogin Failed.Server is %v cannot login.", control.ServerState)
// //		sMsg.RetCode = errcode.ServerStopError
// //		return sMsg
// //	}
// //
// //	if accountname != msg.GetAccountName() && acctype == GuestType {
// //		glog.SErrorf("TokenLogin UserNameError.%v", msg.GetAccountName())
// //		sMsg.RetCode = errcode.UserNameError
// //		return sMsg
// //	}
// //
// //	if mobile != msg.GetAccountName() && acctype == AccountType {
// //		glog.SErrorf("TokenLogin UserNameError.%v", msg.GetAccountName())
// //		sMsg.RetCode = errcode.UserNameError
// //		return sMsg
// //	}
// //
// //	if token != msg.GetToken() {
// //		glog.SErrorf("TokenLogin PasswordError.%v", msg.GetToken())
// //		sMsg.RetCode = errcode.PasswordError
// //		return sMsg
// //	}
// //
// //	// if platformid != msg.GetPlatFormID() {
// //	// 	glog.SErrorf("TokenLogin PlatformError.%v", msg.GetPlatFormID())
// //	// 	sMsg.RetCode = errcode.PlatformError
// //	// 	return sMsg
// //	// }
// //
// //	if status == StatusForbid {
// //		glog.SErrorf("TokenLogin AccountForbid.%v", status)
// //		sMsg.RetCode = errcode.AccountForbid
// //		return sMsg
// //	}
// //
// //	if _, ok := p.uids[id]; ok {
// //		glog.SWarnf("TokenLogin Failed.Mutiple %v", id)
// //		// sMsg.RetCode = errcode.MutilLoginError
// //		// return sMsg
// //	}
// //
// //	newToken := p.genPassword(accountname)
// //
// //	_, err = p.db.Exec(`UPDATE account_t SET ipaddr = ? ,logintime = ? , status = ?, token = ? WHERE userid = ?`, msg.GetIPAddr(), time.Now().Unix(), StatusLogin, newToken, id)
// //	if err != nil {
// //		glog.SErrorf("TokenLogin Failed.%+v update account error:%v", msg, err)
// //		sMsg.RetCode = errcode.DBError
// //		return sMsg
// //	}
// //
// //	_, err = p.db.Exec(`INSERT INTO login_info_t (userid,optime,type,ip,iparea,terminaltype,osvesion) VALUES(?,?,?,?,?,?,?)`, id, time.Now().Unix(), LoginType, msg.GetIPAddr(), GetIPAddr(msg.GetIPAddr()), msg.GetTerminalType(), msg.GetClientVersion())
// //	if err != nil {
// //		glog.SErrorf("TokenLogin Failed.%+v insert login_info_t error:%v", msg, err)
// //		sMsg.RetCode = errcode.DBError
// //		return sMsg
// //	}
// //
// //	p.UpdatePhoneInfo(id, msg.GetModel(), msg.GetVersion())
// //
// //	p.uids[id] = msg.GetAccountName()
// //
// //	sMsg.RetCode = errcode.CodeOK
// //	sMsg.UserID = id
// //	sMsg.HeatTimes = HeatTime
// //	sMsg.Token = newToken
// //	return sMsg
// //}
//
// // Check 有效性
//
//	func (p *UserLogin) Check(mobile string, gameid int32) *plr.S0000019 {
//		rMsg := &plr.S0000019{}
//		var id int
//		p.db.QueryRow(`SELECT userid FROM account_t WHERE mobile = ? and game_id = ?`, mobile, gameid).Scan(&id)
//		if id == 0 {
//			rMsg.RetCode = errcode.MobileNotExst
//		}
//		return rMsg
//	}
//
// // ResetPassword 重置密码
//
//	func (p *UserLogin) ResetPassword(msg *plr.C0000020) *plr.S0000020 {
//		rMsg := &plr.S0000020{}
//		mobile := msg.GetMobile()
//		mobilecode := msg.GetMobileAuthCode()
//		v, ok := p.mobiles[mobile]
//
//		sysRet := p.checkSysMobile(mobile, mobilecode)
//		if !ok {
//			if sysRet != 0 {
//				rMsg.RetCode = sysRet
//				glog.SErrorf("MobileLogin failed %v", rMsg.RetCode)
//				return rMsg
//			}
//		} else {
//			if v.Expire < time.Now().Unix() && sysRet != 0 {
//				glog.SErrorf("ResetPassword auth code expire %v", mobile)
//				rMsg.RetCode = errcode.AuthCodeError
//				return rMsg
//			}
//
//			if v.MachineID != msg.GetMachineID() && sysRet != 0 {
//				glog.SErrorf("ResetPassword failed machineid %v:%v ", msg.GetMachineID(), v.MachineID)
//				rMsg.RetCode = errcode.InvalidParamError
//				return rMsg
//			}
//
//			if v.MobileAuth != msg.GetMobileAuthCode() && sysRet != 0 {
//				glog.SErrorf("ResetPassword auth code failed %v:%v", msg.GetMobileAuthCode(), v.MobileAuth)
//				rMsg.RetCode = errcode.AuthCodeError
//				return rMsg
//			}
//		}
//
//		password := util.EncryptPassword(msg.GetPassword(), Salt)
//		_, err := p.db.Exec(`UPDATE account_t SET password = ?,machineid=? WHERE mobile = ? and game_id = ?`, password, msg.GetMachineID(), mobile, msg.GetGameID())
//		if err != nil {
//			rMsg.RetCode = errcode.DBError
//		} else {
//			delete(p.mobiles, mobile)
//		}
//		return rMsg
//	}
//
// // AccountLogout logout
// //func (p *UserLogin) AccountLogout(id int32) error {
// //	delete(p.uids, id)
// //	if id > 0 {
// //		_, err := p.db.Exec(`UPDATE account_t SET status = ?,logouttime =? WHERE userid = ? AND status = ?`, StatusActive, time.Now().Unix(), id, StatusLogin)
// //		if err != nil {
// //			glog.SErrorf("AccountLogout Failed. update account error:%v", err)
// //		}
// //
// //		_, err = p.db.Exec(`INSERT INTO login_info_t (userid,optime,type) VALUES(?,?,?)`, id, time.Now().Unix(), LogOutType)
// //		if err != nil {
// //			glog.SErrorf("AccountLogout Failed. insert login_info_t error:%v", err)
// //		}
// //		return err
// //	}
// //	return nil
// //}

func (p *UserLogin) AccountLogout(id string) error {
	delete(p.uids, id)
	if len(id) > 0 {
		err := redisx.Del(context.Background(), cachekey.GetUserKey(id)).Err()
		return err
	}
	return nil
}

// // AccountRegister for acc
//
//	func (p *UserLogin) AccountRegister(msg *plr.C0000006) *plr.S0000006 {
//		return nil
//		//sMsg := &plr.S0000006{}
//		//
//		//err := p.db.Ping()
//		//if err != nil {
//		//	glog.SErrorf("DB Error.error:%v", err)
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//
//		//acc := new(Account)
//		//acc.accountname = msg.GetAccountName()
//		//acc.acctype = AccountType
//		//acc.createtime = time.Now().Unix()
//		//acc.faceid = msg.GetFaceID()
//		//acc.ipaddr = msg.GetIPAddr()
//		//loc, err := ip17mon.Find(acc.ipaddr)
//		//if err == nil {
//		//	acc.iparea = loc.String()
//		//}
//		//acc.machineid = msg.GetMachineID()
//		//acc.mobile = msg.GetMobile()
//		//acc.nickname = acc.accountname
//		//acc.password = msg.GetPassword()
//		//acc.platformid = msg.GetPlatFormID()
//		//acc.sexuatily = msg.GetSexuality()
//		//acc.gameid = msg.GetGameID()
//		//acc.status = StatusActive
//		//
//		//bExsit, err := p.isExsitAcc(acc)
//		//if bExsit || err != nil {
//		//	glog.SErrorf("isExsitAcc:%v.error:%v ", bExsit, err)
//		//	sMsg.RetCode = errcode.MutilAccNameError
//		//	return sMsg
//		//}
//		//
//		//bExsit, err = p.isExsitMobile(acc)
//		//if bExsit || err != nil {
//		//	glog.SErrorf("isExsitMobile:%v.error:%v ", bExsit, err)
//		//	sMsg.RetCode = errcode.MutilMobileError
//		//	return sMsg
//		//}
//		//
//		//tx, err := p.db.Begin()
//		//if err != nil {
//		//	glog.SErrorf("DB Error.Begin error:%v", err)
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//err = p.insertAcc(acc, tx)
//		//if err != nil {
//		//	tx.Rollback()
//		//	sMsg.RetCode = errcode.DBError
//		//	return sMsg
//		//}
//		//
//		//tx.Commit()
//		//
//		//sMsg.RetCode = errcode.CodeOK
//		//sMsg.UserID = int32(acc.id)
//		//return sMsg
//	}
//
// // UpdateAccountInfo for acc info
// func (p *UserLogin) UpdateAccountInfo() {
//
// }
//
// LogoutAll logout all network error
func (p *UserLogin) LogoutAll() {

	ids := make([]string, 0)
	for i := range p.uids {
		ids = append(ids, i)
	}

	for _, v := range ids {
		p.AccountLogout(v)
	}
}

//
//// ClearExpireMobile clear expire
//func (p *UserLogin) ClearExpireMobile() {
//	removes := []string{}
//	for k, v := range p.mobiles {
//		if v.Expire < time.Now().Unix() {
//			removes = append(removes, k)
//		}
//	}
//	for _, v := range removes {
//		delete(p.mobiles, v)
//	}
//
//	removes = []string{}
//	for k, v := range p.sysMobiles {
//		if v.Expire < time.Now().Unix() {
//			removes = append(removes, k)
//		}
//	}
//	for _, v := range removes {
//		delete(p.sysMobiles, v)
//	}
//}
//
//// MobileLogin login for mobile
//func (p *UserLogin) MobileLogin(msg *plr.C0000015) *plr.S0000015 {
//	return nil
//	//rMsg := &plr.S0000015{}
//	//mobile := msg.GetMobileName()
//	//mobilecode := msg.GetMobileAuthCode()
//	//machineid := msg.GetMachineID()
//	//ip := msg.GetIPAddr()
//	//platfromid := msg.GetPlatFormID()
//	//gameid := msg.GetGameID()
//	//
//	//v, ok := p.mobiles[mobile]
//	//sysRet := p.checkSysMobile(mobile, mobilecode)
//	//if !ok {
//	//	if sysRet != 0 {
//	//		rMsg.RetCode = sysRet
//	//		glog.SErrorf("MobileLogin failed %v", rMsg.RetCode)
//	//		return rMsg
//	//	}
//	//} else {
//	//	if v.Expire < time.Now().Unix() && sysRet < 0 {
//	//		glog.SErrorf("MobileLogin auth code expire %v", mobile)
//	//		rMsg.RetCode = errcode.AuthCodeError
//	//		return rMsg
//	//	}
//	//
//	//	if v.MachineID != machineid && sysRet < 0 {
//	//		glog.SErrorf("MobileLogin failed machineid %v:%v ", machineid, v.MachineID)
//	//		rMsg.RetCode = errcode.InvalidParamError
//	//		return rMsg
//	//	}
//	//
//	//	if v.MobileAuth != mobilecode && sysRet < 0 {
//	//		glog.SErrorf("MobileLogin auth code failed %v:%v", mobilecode, v.MobileAuth)
//	//		rMsg.RetCode = errcode.AuthCodeError
//	//		return rMsg
//	//	}
//	//}
//	//
//	//var acc struct {
//	//	id         int32
//	//	platformid int32
//	//	mobile     string
//	//	status     int32
//	//	acctype    int32
//	//	accname    string
//	//}
//	//
//	//err := p.db.QueryRow(`SELECT userid,mobile,status,platformid,acctype,accountname FROM account_t
//	//	WHERE mobile = ? and game_id = ?`, mobile, gameid).Scan(&acc.id,
//	//	&acc.mobile, &acc.status, &acc.platformid, &acc.acctype, &acc.accname)
//	//
//	//if err == sql.ErrNoRows {
//	//	id, token, err := p.insertMobileAccount(mobile, machineid, ip, platfromid, gameid)
//	//	if err != nil {
//	//		rMsg.RetCode = errcode.DBError
//	//		return rMsg
//	//	}
//	//	rMsg.UserID = int32(id)
//	//	rMsg.HeatTimes = HeatTime
//	//	rMsg.Token = token
//	//	delete(p.mobiles, mobile)
//	//	delete(p.sysMobiles, mobile)
//	//	return rMsg
//	//}
//	//
//	//if err != nil {
//	//	glog.SErrorf("MobileLogin failed error:%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return rMsg
//	//}
//	//
//	//if !control.CanLogin(acc.id, p.db) {
//	//	glog.SErrorf("AccountLogin Failed.Server is %v cannot login.", control.ServerState)
//	//	rMsg.RetCode = errcode.ServerStopError
//	//	return rMsg
//	//}
//	//
//	//if acc.status == StatusForbid {
//	//	glog.SErrorf("MobileLogin failed user forbid:%v", StatusForbid)
//	//	rMsg.RetCode = errcode.AccountForbid
//	//	return rMsg
//	//}
//	//
//	//if acc.acctype == GuestType {
//	//	glog.SErrorf("MobileLogin failed is guest.mobile:%v", mobile)
//	//	rMsg.RetCode = errcode.UserNameError
//	//	return rMsg
//	//}
//	//
//	//newToken := p.genPassword(acc.accname)
//	//_, err = p.db.Exec(`UPDATE account_t SET ipaddr = ? ,logintime = ? , status = ?, token = ? WHERE userid = ?`, msg.GetIPAddr(), time.Now().Unix(), StatusLogin, newToken, acc.id)
//	//if err != nil {
//	//	glog.SErrorf("MobileLogin Failed. update account error:%v", err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return rMsg
//	//}
//	//
//	//_, err = p.db.Exec(`INSERT INTO login_info_t (userid,optime,type,ip,iparea,terminaltype,osvesion) VALUES(?,?,?,?,?,?,?)`, acc.id, time.Now().Unix(), LoginType, msg.GetIPAddr(), GetIPAddr(msg.GetIPAddr()), msg.GetTerminalType(), msg.GetClientVersion())
//	//if err != nil {
//	//	glog.SErrorf("MobileLogin Failed.%+v insert login_info_t error:%v", msg, err)
//	//	rMsg.RetCode = errcode.DBError
//	//	return rMsg
//	//}
//	//
//	//p.uids[acc.id] = mobile
//	//
//	//delete(p.mobiles, mobile)
//	//delete(p.sysMobiles, mobile)
//	//rMsg.UserID = acc.id
//	//rMsg.HeatTimes = HeatTime
//	//rMsg.Token = newToken
//	//return rMsg
//}
//
//// checkSysMobile checkSysMobile
//func (p *UserLogin) checkSysMobile(phone, mobilecode string) int32 {
//	v1, ok1 := p.sysMobiles[phone]
//	if !ok1 {
//		return errcode.UserNameError
//	}
//	if v1.Expire < time.Now().Unix() {
//		return errcode.AuthCodeError
//	}
//	if v1.MobileAuth != mobilecode {
//		return errcode.AuthCodeError
//	}
//	return 0
//}
//
//// BindUserMobile bind phone
//func (p *UserLogin) BindUserMobile(msg *plr.C0000005) *plr.S0000005 {
//	return nil
//	//auth := MobileAuth{}
//	//auth.Expire = time.Now().Unix() + 3600*12
//	//// auth.Mobile = "12345678901"
//	//// auth.MobileAuth = "012345"
//	//// msg.Mobile = "12345678901"
//	//// p.sysMobiles["12345678901"] = auth
//	//rMsg := &plr.S0000005{}
//	//uid := msg.GetUserID()
//	//phone := msg.GetMobile()
//	//mobilecode := msg.GetMobileAuthCode()
//	//machineid := msg.GetMachineID()
//	//platformid := msg.GetPlatFormID()
//	//password := msg.GetPassword()
//	//
//	//v, ok := p.mobiles[phone]
//	//
//	//sysRet := p.checkSysMobile(phone, mobilecode)
//	//if !ok {
//	//	if sysRet != 0 {
//	//		rMsg.RetCode = sysRet
//	//		glog.SErrorf("MobileLogin failed %v", rMsg.RetCode)
//	//		return rMsg
//	//	}
//	//} else {
//	//	if v.Expire < time.Now().Unix() && sysRet != 0 {
//	//		glog.SErrorf("MobileLogin auth code expire %v", phone)
//	//		rMsg.RetCode = errcode.AuthCodeError
//	//		return rMsg
//	//	}
//	//
//	//	if v.MachineID != machineid && sysRet != 0 {
//	//		glog.SErrorf("MobileLogin failed machineid %v:%v ", machineid, v.MachineID)
//	//		rMsg.RetCode = errcode.InvalidParamError
//	//		return rMsg
//	//	}
//	//
//	//	if v.MobileAuth != mobilecode && sysRet != 0 {
//	//		glog.SErrorf("MobileLogin auth code failed %v:%v", mobilecode, v.MobileAuth)
//	//		rMsg.RetCode = errcode.AuthCodeError
//	//		return rMsg
//	//	}
//	//}
//	//
//	//token, err := p.bindMobile(uid, platformid, phone, machineid, util.EncryptPassword(password, Salt))
//	//if err != nil {
//	//	if err.Error() == "Mobile is exsit" {
//	//		rMsg.RetCode = errcode.MutilMobileError
//	//	} else {
//	//		rMsg.RetCode = errcode.DBError
//	//	}
//	//	glog.SErrorf("BindUserMobile failed %v", err)
//	//	return rMsg
//	//}
//	//
//	//delete(p.mobiles, phone)
//	//delete(p.sysMobiles, phone)
//	//rMsg.Token = token
//	//return rMsg
//}
//
//func (p *UserLogin) bindMobile(uid, platfromid int32, phone, machineid, password string) (string, error) {
//	return "", nil
//	//tx, err := p.db.Begin()
//	//if err != nil {
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	return "", err
//	//}
//	//
//	//var (
//	//	oPlatformid int32
//	//	oMachineid  string
//	//	// nickName    string
//	//	accName string
//	//	// mobile  string
//	//)
//	//
//	//var id int
//	//var gameid int32
//	//err = tx.QueryRow(`SELECT game_id FROM account_t WHERE userid = ? `, uid).Scan(&gameid)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	return "", fmt.Errorf("DB error:%v", err)
//	//}
//	//err = tx.QueryRow(`SELECT userid FROM account_t WHERE mobile = ? AND game_id = ?`, phone, gameid).Scan(&id)
//	//if id != 0 {
//	//	tx.Rollback()
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	return "", fmt.Errorf("Mobile is exsit %v", phone)
//	//}
//	//
//	//err = tx.QueryRow(`SELECT platformid,machineid,accountname FROM account_t WHERE userid = ?`, uid).Scan(&oPlatformid,
//	//	&oMachineid, &accName)
//	//if err != nil {
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	tx.Rollback()
//	//	return "", err
//	//}
//	//
//	//// if mobile != "" {
//	//// 	tx.Rollback()
//	//// 	return "", fmt.Errorf("MutilMobileError")
//	//// }
//	//
//	//// if oPlatformid != platfromid {
//	//// 	tx.Rollback()
//	//// 	return "", fmt.Errorf("Invalid params for platformid:%v", platfromid)
//	//// }
//	//
//	//if oMachineid != machineid {
//	//	tx.Rollback()
//	//	return "", fmt.Errorf("Invalid params for machineid:%v", machineid)
//	//}
//	//
//	//token := p.genPassword(accName)
//	//
//	//// nickName = strings.Replace(nickName, "游客", "会员", 1)
//	//
//	//_, err = tx.Exec(`UPDATE account_t SET mobile = ?,bindtime=?,acctype =?,password = ?,token=? WHERE userid = ?`, phone, time.Now().Unix(), AccountType, password, token, uid)
//	//if err != nil {
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	tx.Rollback()
//	//	return "", err
//	//}
//	//
//	//settlewid, err := models.InsertRecordAmount(uid, models.RecordBindMobile, "", tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//	//	return "", err
//	//}
//	//
//	//_, err = tx.Exec(`UPDATE userwealth_t SET wealth = wealth + ? WHERE userid = ?`, conf.Conf.BindMobileWealth, uid)
//	//
//	//if err != nil {
//	//	glog.SErrorf("bindMobile failed %v", err)
//	//	tx.Rollback()
//	//	return "", err
//	//}
//	//
//	//err = models.UpdateRecordAmount(settlewid, uid, tx)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("update recordAmount failed for db.%v", err.Error())
//	//	return "", err
//	//}
//	//
//	//var count int32
//	//err = tx.QueryRow(`SELECT count(1) FROM userextend_t WHERE userid = ?`, uid).Scan(&count)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Update userextend_t info failed.%v", err)
//	//	return "", err
//	//}
//	//
//	//if count == 0 {
//	//	_, err = tx.Exec(`INSERT INTO userextend_t (userid,alipay,alipaydetail,bankaccount,bankno,bankname,bankdetail) VALUE(?,"","","","","","")`, uid)
//	//	if err != nil {
//	//		glog.SErrorf("insert userextend_t  failed %v", err)
//	//		tx.Rollback()
//	//		return "", err
//	//	}
//	//}
//	//
//	//return token, tx.Commit()
//}
//
//func (p *UserLogin) insertMobileAccount(phone, machineid, ip string, platfromid int32, gameid int32) (string, string, error) {
//	return "", "", nil
//	//now := time.Now().Unix()
//	//accName := fmt.Sprintf("%s%d%d", strconv.Itoa(int(platfromid)), p.gid, now)
//	//token := p.genPassword(accName)
//	//iparea := ""
//	//loc, err := ip17mon.Find(ip)
//	//if err == nil {
//	//	iparea = loc.String()
//	//}
//	//nickName := fmt.Sprintf("会员%s%v", strconv.Itoa(int(platfromid)), phone)
//	//acctype := AccountType
//	//faceid := 1
//	//sex := 1
//	//status := 1
//	//cTime := now
//	//lTime := now
//	//
//	//tx, err := p.db.Begin()
//	//if err != nil {
//	//	glog.SErrorf("Insert mobile account failed.error:%v", err)
//	//	return "", token, err
//	//}
//	//
//	//ret, err := tx.Exec(`INSERT INTO account_t (accountname,password,machineid,ipaddr,reg_ip,iparea,nickname,acctype,faceid,sexuatily
//	//	,platformid,mobile,status,createtime,logintime,token,game_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
//	//	accName, token, machineid, ip, ip, iparea, nickName, acctype, faceid, sex,
//	//	platfromid, phone, status, cTime, lTime, token, gameid)
//	//
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert mobile account failed.error:%v", err)
//	//	return "", token, err
//	//}
//	//
//	//id, err := ret.LastInsertId()
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert mobile account failed LastInsertId Failed.error:%v", err)
//	//	return "", token, err
//	//}
//	//
//	//_, err = tx.Exec(`INSERT INTO userwealth_t (userid,wealth) VALUES(?,?)`, id, conf.Conf.AccountBaseWealth)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert mobile account failed.error:%v", err)
//	//	return "", token, err
//	//}
//	//
//	//_, err = tx.Exec(`INSERT INTO userbank_t (userid,bankamount) VALUES(?,?)`, id, 0)
//	//if err != nil {
//	//	tx.Rollback()
//	//	glog.SErrorf("Insert mobile account failed.error:%v", err)
//	//	return "", token, err
//	//}
//	//
//	//tx.Commit()
//	//p.gid++
//	//return id, token, err
//}
//
//// GenMobileCode generate
//func (p *UserLogin) GenMobileCode(msg *plr.C0000016) *plr.S0000016 {
//	rMsg := &plr.S0000016{}
//
//	mobile := msg.GetMobileName()
//	machineid := msg.GetMachineID()
//	if mobile == "" {
//		glog.Errorln("GenMobileCode params failed mobile is nil")
//		rMsg.RetCode = errcode.InvalidParamError
//		return rMsg
//	}
//
//	if machineid == "" {
//		glog.Errorln("GenMobileCode params failed machineid is nil")
//		rMsg.RetCode = errcode.InvalidParamError
//		return rMsg
//	}
//
//	auth := MobileAuth{}
//	auth.Expire = time.Now().Unix() + AuthCodeFullExpire
//	auth.MachineID = machineid
//	auth.Mobile = mobile
//	auth.MobileAuth = p.sendSMS(mobile)
//	p.mobiles[mobile] = auth
//
//	rMsg.ExpireTime = AuthCodeExpireTime
//	return rMsg
//}
//
//// GenMobileCodeByAdmin 管理后台获取验证码
//func (p *UserLogin) GenMobileCodeByAdmin(msg *explr.C1010027) *explr.S1010027 {
//	rMsg := &explr.S1010027{}
//	mobile := msg.GetPhone()
//
//	auth := MobileAuth{}
//	auth.Expire = time.Now().Unix() + 36000
//	auth.Mobile = mobile
//	auth.MobileAuth = util.RandomCode(6)
//	p.sysMobiles[mobile] = auth
//
//	rMsg.Code = auth.MobileAuth
//	return rMsg
//}
//
//func (p *UserLogin) sendSMS(mobile string) string {
//	authcode := util.RandomCode(6)
//	glog.SInfof("Authcode :%v", authcode)
//	SendSMS(mobile, authcode)
//	return authcode
//}
//
//func (p *UserLogin) buildGuest(id int32) *Account {
//	now := time.Now().Unix()
//	acc := new(Account)
//	acc.accountname = fmt.Sprintf("%s%d%d", strconv.Itoa(int(id)), p.gid, now)
//	acc.password = p.genPassword(acc.accountname) // fmt.Sprintf("%x", h.Sum(nil))
//	acc.acctype = GuestType
//	acc.createtime = now
//	// acc.nickname = fmt.Sprintf("游客%s%v", strconv.Itoa(int(id)), util.RandomCode(6))
//	acc.nickname = fmt.Sprintf("游客%s%d%d", getPlatString(id), p.gid, (now-StartTime)/3600)
//	acc.status = StatusActive
//	p.gid++
//	if p.gid == 100000 {
//		p.gid = 500
//	}
//	return acc
//}
//
//func getPlatString(id int32) string {
//	switch id {
//	case 1:
//		return "a"
//	default:
//		return "x"
//	}
//}
//
//func (p *UserLogin) insertAcc(acc *Account, tx *sql.Tx) error {
//
//	guestWealth := conf.Conf.GuestBaseWealth
//	acccnt := 0
//	err := tx.QueryRow(`select count(1) from account_t where machineid=?	and game_Id=?`, acc.machineid, acc.gameid).Scan(&acccnt)
//	if err != nil {
//		glog.SErrorf("query account_t Failed.error:%v acc:%+v", err, acc)
//		return err
//	}
//	//同一个机器码只赠送一次金币
//	if acccnt > 0 {
//		guestWealth = 0
//	}
//
//	//如果是玩家推广员
//	var chanid int32
//	if models.IsRealPlayer(acc.platformid) {
//		chanid = acc.platformid
//		acc.platformid = 0
//	}
//
//	ret, err := tx.Exec(`INSERT INTO account_t (accountname,password,machineid,ipaddr,reg_ip,iparea,nickname,acctype,faceid,sexuatily
//		,platformid,mobile,status,createtime,logintime,game_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
//		acc.accountname, acc.password, acc.machineid, acc.ipaddr, acc.ipaddr, acc.iparea, acc.nickname, acc.acctype, acc.faceid, acc.sexuatily,
//		acc.platformid, acc.mobile, acc.status, acc.createtime, acc.logintime, acc.gameid)
//	if err != nil {
//		glog.SErrorf("insertAcc Failed.error:%v acc:%+v", err, acc)
//		return err
//	}
//
//	acc.id, err = ret.LastInsertId()
//	if err != nil {
//		glog.SErrorf("LastInsertId Failed.error:%v", err)
//		return err
//	}
//
//	ret, err = tx.Exec(`INSERT INTO userwealth_t (userid,wealth) VALUES(?,?)`, acc.id, guestWealth)
//	if err != nil {
//		glog.SErrorf("insert userwealth_t Failed.error:%v acc:%+v", err, acc)
//		return err
//	}
//
//	_, err = tx.Exec(`INSERT INTO userbank_t (userid,bankamount) VALUES(?,?)`, acc.id, 0)
//	if err != nil {
//		glog.SErrorf("insert userbank_t Failed.error:%v acc:%+v", err, acc)
//		return err
//	}
//
//	_, err = tx.Exec(`INSERT INTO userextend_t (userid,alipay,alipaydetail,bankaccount,bankno,bankname,bankdetail) VALUE(?,"","","","","","")`, acc.id)
//	if err != nil {
//		glog.SErrorf("insert userextend_t  failed %v", err)
//		return err
//	}
//
//	if guestWealth > 0 {
//		err = models.InsertRegisterRecordAmount(int32(acc.id), models.RecordRegister, tx)
//		if err != nil {
//			glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
//			return err
//		}
//	}
//
//	if chanid > 0 {
//		var puid int32 = -99
//		var puids string
//		glog.SInfof("uid:%v", acc.id)
//		rs, err := tx.Query(`select parentuid,parentuids from wagency_user_t where userid=?`, chanid)
//		if err != nil {
//			glog.SErrorf("query parent agency user failed.uid:%v,%v", chanid, err.Error())
//			return err
//		}
//		defer rs.Close()
//		for rs.Next() {
//			err = rs.Scan(&puid, &puids)
//			if err != nil {
//				glog.SErrorf("scan parent agency user failed.uid:%v,%v", chanid, err.Error())
//				return err
//			}
//		}
//
//		//如果还没有一级代理记录 先插入一级代理记录
//		if puid == -99 {
//			_, err = tx.Exec(`insert into wagency_user_t(userid,parentuid,parentuids,createtime) values(?,?,?,?)`, chanid, 0, "0,", time.Now().Unix())
//			if err != nil {
//				glog.SErrorf("insert parent agency user failed.uid:%v,%v", chanid, err.Error())
//				return err
//			}
//			err := tx.QueryRow(`select parentuid,parentuids from wagency_user_t where userid=?`, chanid).Scan(&puid, &puids)
//			if err != nil {
//				glog.SErrorf("query parent agency user failed.uid:%v,%v", chanid, err.Error())
//				return err
//			}
//		}
//		//插入玩家本身记录
//		puidsstr := puids
//		if puidsstr == "0," {
//			puidsstr = ""
//		}
//		puidsstr = fmt.Sprintf("%v%v,", puidsstr, chanid)
//		_, err = tx.Exec(`insert into wagency_user_t(userid,parentuid,parentuids,createtime) values(?,?,?,?)`, acc.id, chanid, puidsstr, time.Now().Unix())
//		if err != nil {
//			glog.SErrorf("insert agency user failed.uid:%v,%v", acc.id, err.Error())
//			return err
//		}
//
//		_, err = tx.Exec(`update wagency_user_t set direnums=direnums+1 where userid=?`, chanid)
//		if err != nil {
//			glog.SErrorf("update direnums failed.uid:%v,%v", chanid, err.Error())
//			return err
//		}
//
//		if puids != "0," {
//			strpuids := puids
//			if strings.HasSuffix(strpuids, ",") {
//				strpuids = strpuids[:len(strpuids)-1]
//			}
//			sql := fmt.Sprintf(`update wagency_user_t set othernums=othernums+1 where userid in (%s)`, strpuids)
//			_, err = tx.Exec(sql)
//			if err != nil {
//				glog.SErrorf("update othernums failed.uid:%v,%v", chanid, err.Error())
//				return err
//			}
//		}
//	}
//
//	// 绑定推广员
//	// go p.bindSale(acc.id, acc.ipaddr, acc.machineid)
//
//	return err
//}
//
//func (p *UserLogin) bindSale(accId int64, ipaddr, machineid string) (err error) {
//	if machineid == "" {
//		glog.SWarnf("userId:%d,machineid is empty", accId)
//		return
//	}
//
//	info := DownloadInfo{
//		ID:    int(accId),
//		OutIP: ipaddr,
//	}
//
//	dv := Device{}
//	var manufacturer sql.NullString
//	if err := p.db.QueryRow(`SELECT manufacturer,phone_model,phone_code,last_system_version,inside_ip FROM phone_info_t WHERE machinecode=?`, machineid).Scan(
//		&manufacturer, &dv.Device, &dv.Code, &dv.Version, &info.InIP); err != nil {
//		glog.SErrorf("获取phone_info_t失败，machineid %v，err:%v", machineid, err)
//		return err
//	}
//
//	dv.Ios = manufacturer.String == "Apple"
//	info.Device = dv
//
//	url := fmt.Sprintf("%s/bind", conf.Conf.BindSaleHost)
//	err = util.PostJSONNotRsp(url, info)
//
//	if err == nil {
//		glog.Infof("发送绑定信息成功，userId(%d)", accId)
//	} else {
//		glog.SErrorf("发送绑定信息失败:%v", err)
//	}
//
//	return
//}
//
//func (p *UserLogin) isExsitMobile(acc *Account) (bool, error) {
//	var id int
//	err := p.db.QueryRow(`SELECT userid FROM account_t WHERE acctype = ? AND mobile = ? AND game_id = ?`, acc.acctype, acc.mobile, acc.gameid).Scan(&id)
//	if err != sql.ErrNoRows {
//		return true, err
//	}
//	return false, nil
//}
//
//func (p *UserLogin) isExsitAcc(acc *Account) (bool, error) {
//	var id int
//	err := p.db.QueryRow(`SELECT userid FROM account_t WHERE accountname = ?`, acc.accountname).Scan(&id)
//	if err != sql.ErrNoRows {
//		return true, err
//	}
//	return false, nil
//}
//
//func (p *UserLogin) genPassword(account string) string {
//	token := fmt.Sprintf("%s%d", account, time.Now().Unix())
//	return util.Md5Hash(token)
//}
//
//// Password return
//func (p *UserLogin) Password(password string) string {
//	return util.EncryptPassword(password, Salt)
//}
//
//// SetBankPassword 设置银行密码
//func (p *UserLogin) SetBankPassword(msg *plr.C0000022) *plr.S0000022 {
//	uid := msg.GetUserID()
//	rmsg := &plr.S0000022{}
//	password := msg.GetPassword()
//	_, err := p.db.Exec(`UPDATE userbank_t SET password = ? WHERE userid = ?`, util.EncryptPassword(password, Salt), uid)
//	if err != nil {
//		rmsg.RetCode = errcode.DBError
//	}
//	return rmsg
//}
//
//// ResetBankPassword 重置银行密码
//func (p *UserLogin) ResetBankPassword(msg *plr.C0000023) *plr.S0000023 {
//	uid := msg.GetUserID()
//	mobile := msg.GetMobile()
//	password := msg.GetPassword()
//	mobilecode := msg.GetMobileAuthCode()
//	rMsg := &plr.S0000023{}
//
//	v, ok := p.mobiles[mobile]
//	sysRet := p.checkSysMobile(mobile, mobilecode)
//	if !ok {
//		if sysRet != 0 {
//			rMsg.RetCode = sysRet
//			glog.SErrorf("MobileLogin failed %v", rMsg.RetCode)
//			return rMsg
//		}
//	} else {
//		if v.Expire < time.Now().Unix() && sysRet != 0 {
//			glog.SErrorf("MobileLogin auth code expire %v", mobile)
//			rMsg.RetCode = errcode.AuthCodeError
//			return rMsg
//		}
//
//		if v.MobileAuth != mobilecode && sysRet != 0 {
//			glog.SErrorf("MobileLogin auth code failed %v:%v", mobilecode, v.MobileAuth)
//			rMsg.RetCode = errcode.AuthCodeError
//			return rMsg
//		}
//	}
//
//	_, err := p.db.Exec(`UPDATE userbank_t SET password = ? WHERE userid = ?`, util.EncryptPassword(password, Salt), uid)
//	if err != nil {
//		rMsg.RetCode = errcode.DBError
//	}
//	delete(p.mobiles, mobile)
//	delete(p.sysMobiles, mobile)
//	return rMsg
//}
//
//// GetIPAddr 获取ip区域
//func GetIPAddr(ip string) string {
//	loc, err := ip17mon.Find(ip)
//	if err == nil {
//		return loc.String()
//	}
//	return "中国"
//}
//
//func (p *UserLogin) SearchMachineUser(msg *plr.C0000028) *plr.S0000028 {
//	sMsg := &plr.S0000028{}
//	if msg.GetMachineID() == "" {
//		glog.SErrorf("GuestLogin Failed.MachineID is nil")
//		sMsg.RetCode = errcode.InvalidParamError
//		return sMsg
//	}
//
//	text := fmt.Sprintf("{%v:%v}", msg.GetMachineID(), guestToken)
//	sign := util.Md5Hash(text)
//	if msg.GetSign() != sign {
//		glog.SErrorf("sign is error")
//		sMsg.RetCode = errcode.InvalidParamError
//		return sMsg
//	}
//	rows, err := p.db.Query(`SELECT userid,mobile,nickname,acctype FROM account_t WHERE machineid=? and game_id = ?`, msg.GetMachineID(), msg.GetGameID())
//	if err != nil {
//		glog.SErrorf("query account err:%v", err)
//		sMsg.RetCode = errcode.DBError
//		return sMsg
//	}
//	defer rows.Close()
//
//	mobiles := make([]string, 0)
//	Uids := make([]string, 0)
//	guestName := ""
//	guestUID := ""
//	for rows.Next() {
//		var mobile, accouontname, userid string
//		var acctype int32
//		err = rows.Scan(&userid, &mobile, &accouontname, &acctype)
//		if err != nil {
//			glog.SErrorf("scan account err:%v", err)
//			sMsg.RetCode = errcode.DBError
//			return sMsg
//		}
//
//		if acctype == 1 {
//			guestName = accouontname
//			guestUID = userid
//
//		} else if acctype == 2 {
//			mobiles = append(mobiles, mobile)
//			Uids = append(Uids, userid)
//		}
//	}
//	sMsg.GestName = guestName
//	sMsg.Mobiles = mobiles
//	sMsg.Uids = Uids
//	sMsg.GestUid = guestUID
//	return sMsg
//}
//
//// UpdatePhoneInfo 更新手机信息
//func (p *UserLogin) UpdatePhoneInfo(uid string, model, version string) error {
//	if model == "" || version == "" {
//		return nil
//	}
//	var modelDB, versionDB string
//	err := p.db.QueryRow(`SELECT last_model,last_version FROM user_phone_t WHERE userid = ?`,
//		uid).Scan(&modelDB, &versionDB)
//
//	if err == sql.ErrNoRows {
//		_, err = p.db.Exec(`INSERT INTO user_phone_t (userid,first_model,first_version,last_model,last_version) VALUES(?,?,?,?,?)`,
//			uid, model, version, model, version)
//		if err != nil {
//			return err
//		}
//	} else if err != nil {
//		return err
//	} else { //err == nil
//		if modelDB != model || versionDB != version {
//			_, err = p.db.Exec(`UPDATE user_phone_t SET last_model=?, last_version=? where userid=?`, model, version, uid)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
