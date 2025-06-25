package core

import (
	"database/sql"
	"fmt"
	"newstars/Protocol/plr"
	"newstars/Server/fish/consts"
	"newstars/Server/fish/model"
	"newstars/Server/fish/round"
	"newstars/Server/fish/table"
	"newstars/framework/core/component"
	"newstars/framework/core/server"
	"newstars/framework/core/session"
	"newstars/framework/game_center"
	"newstars/framework/glog"
	model2 "newstars/framework/model"
	"newstars/framework/util/decimal"
	"strconv"
	"time"
)

// FishServer 捕鱼
type FishServer struct {
	component.Base
	db      *sql.DB
	mgr     map[int32]*table.FishManager
	usrs    map[string]*table.FishSeat
	rs      map[int32]*round.RSession
	control *round.RevenueControl
}

// NewFishServer create instance
func NewFishServer(db *sql.DB) *FishServer {
	return &FishServer{
		db:      db,
		mgr:     make(map[int32]*table.FishManager),
		usrs:    make(map[string]*table.FishSeat),
		rs:      make(map[int32]*round.RSession),
		control: round.NewRevenueControl(db),
	}
}

// Init once
func (p *FishServer) AfterInit() {
	rooms, err := model.QueryRoomListByKind(consts.FishKindID, p.db)
	if err != nil {
		glog.SErrorf("QueryRoomListByKind failed kindid:%v", consts.FishKindID)
	}
	for _, v := range rooms {
		rid := v.Gameroomid
		if _, ok := p.mgr[rid]; !ok {
			p.mgr[rid] = table.NewFishManager(rid, p.db)
			err = p.mgr[rid].Init()
			if err != nil {
				glog.SErrorf("NewFishManager failed rid:%v error:%v", rid, err)
			}
		}
	}

	server.OnSessionClosed(p.doOutterSessionClosed)

	idmap, err := model.QueryPlatformTableMap(p.db, consts.FishKindID)
	if err != nil {
		glog.SErrorf("Init PlatformTableMap failed.err:%v", err)
		return
	}
	consts.PlatformTableMap = idmap
}

// C3080001 请求房间
func (p *FishServer) C3080001(s *session.Session, msg *plr.C3080001, mid uint) error {
	rsp := &plr.S3080001{}
	if msg.GetGameKindID() != consts.FishKindID {
		glog.SErrorf("Invalid Kindid msg.%v", msg)
		rsp.RetCode = consts.ErrorInvalidKindID
		return s.Response(rsp, mid)
	}
	rooms, err := model.QueryRoomListByKind(consts.FishKindID, p.db)
	if err != nil {
		glog.SErrorf("QueryRoomListByKind failed kindid:%v", consts.FishKindID)
		rsp.RetCode = consts.ErrorInvalidKindID
		return s.Response(rsp, mid)
	}
	for _, v := range rooms {
		rid := v.Gameroomid
		if v.Baseamount <= 0 {
			glog.SErrorf("game room baseamount config err. rid:%v baseamount:%v", rid, v.Baseamount)
			rsp.RetCode = consts.ErrorDB
			return s.Response(rsp, mid)
		}

		room := &plr.S3080001_RoomInfo{}
		room.BaseAmount = v.Baseamount
		room.MaxEnterAmount, _ = decimal.New(v.Maxenteramount, 0).Truncate(0).Float64()
		minEnter := decimal.New(int64(v.Minenteramount), 0).Truncate(3)
		if minEnter.Equal(decimal.Zero) {
			minEnter = decimal.NewFromFloat(consts.RoomMinEnterAmount)
		}
		room.MinEnterAmount, _ = minEnter.Float64()
		room.RoomID = rid
		room.RoomName = v.Gameroomname
		room.MinRatio = v.Baseamount
		if v.Baseamount == consts.SpecialRatio {
			room.MaxRatio, _ = decimal.NewFromFloat(v.Baseamount).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Float64()
		} else {
			room.MaxRatio, _ = decimal.NewFromFloat(v.Baseamount).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Float64()
		}
		rsp.Rooms = append(rsp.Rooms, room)
	}
	return s.Response(rsp, mid)
}

// C3080002 进入房间
func (p *FishServer) C3080002(s *session.Session, msg *plr.C3080002, mid uint) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime).Milliseconds() // 得到毫秒
		fmt.Printf("call C3080002 消耗的时间: %d ms\n", elapsed)
	}()
	rsp := &plr.S3080002{}
	rid := msg.GetRoomID()
	uid := msg.GetUserID()

	if _, ok := p.usrs[uid]; ok {
		glog.SErrorf("Enter room failed.user exists uid:%v", uid)
		rsp.RetCode = consts.ErrorInvalidParams
		return s.Response(rsp, mid)
	}

	if _, ok := p.mgr[rid]; !ok {
		glog.SErrorf("Enter room failed.rid invalid.%v", rid)
		rsp.RetCode = consts.ErrorInvalidParams
		return s.Response(rsp, mid)
	}

	//uinfo, err := models.QueryUserInfo(uid, p.db)
	uinfo, err := game_center.GetUserInfoByID(uid)
	if err != nil {
		glog.SErrorf("Enter room failed.QueryUserInfo uid:%v error：%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		return s.Response(rsp, mid)
	}
	var minEnter float64
	if minEnter, err = p.canAccess(uinfo, rid); err != nil {
		glog.SErrorf("Enter room failed.User wealth is not enough.uid:%v wealth:%v", uid, uinfo.Wealth)
		rsp.RetCode = consts.ErrorBalance
		return s.Response(rsp, mid)
	}

	var gameid int32
	if gid, ok := consts.PlatformTableMap[strconv.Itoa(int(uinfo.GameID))]; ok {
		gameid = gid
	}

	t := p.mgr[rid].TakeTable(uinfo.DisPlayName, gameid)
	if t == nil {
		glog.SErrorf("Enter room failed.TakeTable failed uid:%v", uid)
		rsp.RetCode = consts.ErrorTakeTable
		return s.Response(rsp, mid)
	}

	err = t.SitDown(uid, uinfo.DisPlayName)
	if err != nil {
		glog.SErrorf("Enter room failed.ErrorSitDown uid:%v", uid)
		rsp.RetCode = consts.ErrorSitDown
		return s.Response(rsp, mid)
	}

	seat := t.QuerySeat(uid)
	if seat == nil {
		glog.SErrorf("Enter room failed.QuerySeat failed uid:%v", uid)
		rsp.RetCode = consts.ErrorSitDown
		return s.Response(rsp, mid)
	}

	p.usrs[uid] = seat
	tid := seat.Tid
	_, ok := p.rs[tid]
	if !ok {
		gids := make([]int32, 0)
		for k, v := range consts.PlatformTableMap {
			if v == gameid {
				kgid, err := strconv.Atoi(k)
				if err == nil {
					gids = append(gids, int32(kgid))
				}
			}
		}
		p.rs[tid] = round.NewRSession(s, p.db, tid, rid, minEnter, p.control, gids)
		p.rs[tid].Go()
		p.rs[tid].OnCallBack(p.doInnerSessionClosed, p.playerLeave)
	}
	p.rs[tid].OnEnter(uid, seat.Sid, mid)
	return nil
}

// C3080003 炮台种类
func (p *FishServer) C3080003(s *session.Session, msg *plr.C3080003, mid uint) error {
	rsp := &plr.S3080003{}
	uid := msg.GetUserID()

	exists, err := model.CheckUserExtendExists(uid, p.db)
	if err != nil {
		glog.SErrorf("query userextend failed uid:%v,err:%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		s.Response(rsp, mid)
		return err
	}
	if !exists {
		err = model.InitUserExtend(uid, p.db)
		if err != nil {
			glog.SErrorf("insert userextend failed uid:%v,err:%v", uid, err)
			rsp.RetCode = consts.ErrorDB
			s.Response(rsp, mid)
			return err
		}
	}
	exusr, err := model.QueryUserExtend(uid, p.db)
	if err != nil {
		glog.SErrorf("query userextend failed uid:%v,err:%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		s.Response(rsp, mid)
		return err
	}

	rows, err := p.db.Query(`select c.id,c.name,c.period_day,c.cost_wealth,c.isneedbuy,ifnull(u.userid,0) as userid,ifnull(u.endtime,0) as endtime,c.viplevel from fish_cannon_t  c
	left join fish_user_cannon_t u on u.cannonid=c.id and u.userid=? and endtime>?  order by c.id `, uid, time.Now().Unix())
	if err != nil {
		glog.SErrorf("Enter room failed.QuerySeat failed uid:%v,err:%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		return s.Response(rsp, mid)
	}
	defer rows.Close()

	//viplvl, _, _ := models.QueryVipLevel(uid, p.db)
	var viplvl int32 = 0
	cannons := make([]*plr.S3080003_CannonInfo, 0)
	for rows.Next() {
		c := &plr.S3080003_CannonInfo{}
		var userID string
		var isneedbuy = 1
		var canlevel int32
		var endtime int64
		err := rows.Scan(&c.ID, &c.Name, &c.PeriodDay, &c.CostWealth, &isneedbuy, &userID, &endtime, &canlevel)
		if err != nil {
			glog.SErrorf("scan cannon failed uid:%v,err:%v", uid, err)
			rsp.RetCode = consts.ErrorDB
			return s.Response(rsp, mid)
		}

		glog.SInfof("ucid :%v,cid:%v,needbuy:%v,uid:%v,userID:%v", exusr.CurrCannonID, c.ID, isneedbuy, uid, userID)
		if exusr.CurrCannonID == c.ID {
			c.UseStatus = consts.CannonInloaded
		} else {
			if userID == uid || isneedbuy == 0 {
				//vip炮台不需要购买 这里isneedbuy=0时要判断下
				if isneedbuy == 0 && viplvl < canlevel {
					c.UseStatus = consts.CannonNotBuy
				} else {
					c.UseStatus = consts.CannonNotNotload
				}
			} else {
				c.UseStatus = consts.CannonNotBuy
			}
		}

		if isneedbuy == 1 {
			lefttime := endtime - time.Now().Unix()
			if lefttime < 0 {
				lefttime = 0
			}
			c.Lefttime = lefttime
		}
		cannons = append(cannons, c)
	}

	rsp.Cannons = cannons
	s.Response(rsp, mid)
	return nil
}

// C3080004 图鉴
func (p *FishServer) C3080004(s *session.Session, msg *plr.C3080004, mid uint) error {
	kinds, err := model.QueryAllFishKinds(p.db)
	rsp := &plr.S3080004{}
	if err != nil {
		glog.SErrorf("db err:%v", err)
		rsp.RetCode = consts.ErrorDB
		s.Response(rsp, mid)
	}
	arr := make([]*plr.S3080004_FishKind, len(kinds))
	for i, v := range kinds {
		k := &plr.S3080004_FishKind{
			ID:       v.ID,
			KindName: v.KindName,
			Score:    v.Score,
		}
		arr[i] = k
	}
	rsp.Kinds = arr
	s.Response(rsp, mid)
	return nil
}

// C3080005 使用技能
func (p *FishServer) C3080005(s *session.Session, msg *plr.C3080005, mid uint) error {
	uid := msg.GetUserID()
	usr, ok := p.usrs[uid]
	rsp := &plr.S3080005{}
	rsp.SkillType = msg.GetSkillType()
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		rsp.RetCode = consts.ErrorInvalidUID
		s.Response(rsp, mid)
		return nil
	}

	if usr.Uid != uid {
		glog.SErrorf("user not match message:%v session:%v", uid, usr.Uid)
		rsp.RetCode = consts.ErrorInvalidUID
		s.Response(rsp, mid)
		return nil
	}

	tid := msg.GetTableID()
	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table not exists:%v", tid)
		rsp.RetCode = consts.ErrorInvalidTid
		s.Response(rsp, mid)
		return nil
	}
	//	userid,tableid,seatno,skilltype int32
	rs.OnSkill(msg.GetUserID(), msg.GetTableID(), msg.GetSeatNo(), msg.GetSkillType(), msg.GetOpType(), mid)
	return nil
}

// C3080006 购买炮台
func (p *FishServer) C3080006(s *session.Session, msg *plr.C3080006, mid uint) error {
	uid := msg.GetUserID()
	rsp := &plr.S3080006{}
	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		rsp.RetCode = consts.ErrorInvalidUID
		s.Response(rsp, mid)
		return nil
	}

	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table RSession not exists:%v", usr.Tid)
		rsp.RetCode = consts.ErrorInvalidTid
		s.Response(rsp, mid)
		return nil
	}

	rs.OnPurchaseCannon(uid, usr.Sid, msg.GetCannonID(), mid)
	return nil
}

// C3080007 装载炮台
func (p *FishServer) C3080007(s *session.Session, msg *plr.C3080007, mid uint) error {
	uid := msg.GetUserID()
	cid := msg.GetCannonID()
	rsp := &plr.S3080007{}

	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		rsp.RetCode = consts.ErrorInvalidUID
		s.Response(rsp, mid)
		return nil
	}

	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table RSession not exists:%v", usr.Tid)
		rsp.RetCode = consts.ErrorInvalidTid
		s.Response(rsp, mid)
		return nil
	}
	rs.OnChangeCannonID(uid, usr.Sid, cid, mid)

	return nil
}

// C3080008 重入房间
func (p *FishServer) C3080008(s *session.Session, msg *plr.C3080008, mid uint) error {
	uid := msg.GetUserID()
	tid := msg.GetTableID()
	sid := msg.GetSeatNo()
	rid := msg.GetRoomID()

	usr, ok := p.usrs[uid]
	rsp := &plr.S3080008{}
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		rsp.RetCode = consts.ErrorInvalidUID
		s.Response(rsp, mid)
		return nil
	}

	if usr.Rid != rid || usr.Tid != tid || usr.Sid != sid {
		glog.SErrorf("param invaild uid:%v,rid:%v[%v],tid:%v[%v],sid:%v[%v]", uid, rid, usr.Rid, tid, usr.Tid, sid, usr.Sid)
		rsp.RetCode = consts.ErrorInvalidParams
		s.Response(rsp, mid)
		return nil
	}

	if _, ok := p.mgr[rid]; !ok {
		glog.SErrorf("Enter room failed.rid invalid.%v", rid)
		rsp.RetCode = consts.ErrorInvalidParams
		return s.Response(rsp, mid)
	}

	//uinfo, err := models.QueryUserInfo(uid, p.db)
	//if err != nil {
	//	glog.SErrorf("Enter room failed.QueryUserInfo uid:%v error：%v", uid, err)
	//	rsp.RetCode = consts.ErrorDB
	//	return s.Response(rsp, mid)
	//}

	uinfo, err := game_center.GetUserInfoByID(uid)
	if err != nil {
		glog.SErrorf("Enter room failed.QueryUserInfo uid:%v error：%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		return s.Response(rsp, mid)
	}

	if _, err = p.canAccess(uinfo, rid); err != nil {
		glog.SErrorf("Enter room failed.User wealth is not enough.uid:%v wealth:%v", uid, uinfo.Wealth)
		rsp.RetCode = consts.ErrorBalance
		return s.Response(rsp, mid)
	}

	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table not exists:%v", tid)
		rsp.RetCode = consts.ErrorInvalidTid
		s.Response(rsp, mid)
		return nil
	}

	rs.OnReEnter(uid, sid, mid)

	return nil

}

//// C1080001 订阅
//func (p *FishServer) C1080001(s *session.Session, msg *explr.C1080001, mid uint) error {
//	uid := msg.GetUID()
//	p.control.SetUID(uid)
//	rsp := &explr.S1080001{}
//	return s.Response(rsp, mid)
//}
//
//// C1080002 捕获概率查询
//func (p *FishServer) C1080002(s *session.Session, msg *explr.C1080002, mid uint) error {
//	rsp := &explr.S1080002{}
//	uid := msg.GetUID()
//	if len(uid) >= 0 {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("invalid uid :%v", uid)
//		return s.Response(rsp, mid)
//	}
//	rates := p.control.GetCaptureRates()
//	rspRates := make([]*explr.S1080002_CaptureRate, 0)
//	for i, v := range rates {
//		r := &explr.S1080002_CaptureRate{}
//		r.RoomID = i
//		r.CaptureRate = v
//		rspRates = append(rspRates, r)
//	}
//	rsp.Rates = rspRates
//	return s.Response(rsp, mid)
//}
//
//// C1080003 库存修改
//func (p *FishServer) C1080003(s *session.Session, msg *explr.C1080003, mid uint) error {
//	rsp := &explr.S1080003{}
//	uid := msg.GetUID()
//	if len(uid) >= 0 {
//		rsp.RetCode = consts.ErrorInvalidParams
//		return s.Response(rsp, mid)
//	}
//	rid := msg.GetRoomID()
//	amount := msg.GetPoolAmount()
//	p.control.SetAmount(rid, amount)
//	//更新到数据库
//	models.UpdateInventoryByRid(rid, amount, p.db)
//	return s.Response(rsp, mid)
//}
//
//func (p *FishServer) C1080004(s *session.Session, msg *explr.C1080004, mid uint) error {
//	rsp := &explr.S1080004{}
//	uid := msg.GetUID()
//	roomid := msg.GetRoomID()
//	rate := msg.GetCaptureRate()
//	if len(uid) >= 0 {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("invalid uid :%v", uid)
//		return s.Response(rsp, mid)
//	}
//
//	if rate < 0 {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("invalid CaptureRate :%v", rate)
//		return s.Response(rsp, mid)
//	}
//
//	err := model.UpdateRoomConfig(p.db, roomid, rate)
//	if err != nil {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("update capturerate failed.err:%v", err)
//		return s.Response(rsp, mid)
//	}
//	p.control.SetCaptureRate(roomid, rate)
//	return s.Response(rsp, mid)
//}
//
//// C1010026 游戏分平台配置更新通知
//func (p *FishServer) C1010026(s *session.Session, msg *explr.C1010026, mid uint) error {
//	rsp := &explr.S1010026{}
//	uid := msg.GetUID()
//	kindid := msg.GetKindID()
//	if len(uid) >= 0 {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("C1010026 invalid uid:%v", uid)
//		return s.Response(rsp, mid)
//	}
//
//	if kindid != consts.FishKindID {
//		rsp.RetCode = consts.ErrorInvalidParams
//		glog.SErrorf("C1010026 invalid gamekindid:%v", kindid)
//		return s.Response(rsp, mid)
//	}
//
//	idmap, err := models.QueryPlatformTableMap(p.db, consts.FishKindID)
//	if err != nil {
//		glog.SErrorf("update PlatformTableMap failed.err:%v", err)
//		rsp.RetCode = consts.ErrorDB
//		return s.Response(rsp, mid)
//	}
//	consts.PlatformTableMap = idmap
//
//	for rk, rv := range p.rs {
//		var tb *table.FishTable
//		for _, tv := range p.mgr {
//			qb := tv.GetTable(rk)
//			if qb != nil {
//				tb = qb
//			}
//		}
//		if tb != nil {
//			gameid := tb.Gameid()
//			gids := make([]int32, 0)
//			for k, v := range consts.PlatformTableMap {
//				if v == gameid {
//					kgid, err := strconv.Atoi(k)
//					if err == nil {
//						gids = append(gids, int32(kgid))
//					}
//				}
//			}
//			glog.SWarnf("更新平台配置 tid:%v，gids:%v", rk, gids)
//			rv.RefreshGameids(gids)
//		}
//	}
//	return s.Response(rsp, mid)
//}

// N3080001 玩家离开
func (p *FishServer) N3080001(s *session.Session, msg *plr.N3080001, mid uint) error {
	uid := msg.GetUserID()
	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		return nil
	}

	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table not exists:%v", usr.Tid)
		return nil
	}

	leaveType := msg.GetLeaveType()
	if leaveType == 1 {
		rs.OnLeave(uid)
	} else {
		rs.OnOffline(uid)
	}
	return nil
}

// N3080002 发射子弹
func (p *FishServer) N3080002(s *session.Session, msg *plr.N3080002, mid uint) error {
	uid := msg.GetUserID()
	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		return nil
	}
	if usr.Sid != msg.GetSeatNo() {
		glog.SErrorf("invalid seat, uid:%v,msg seatno:%v,sessioin seatno:%v", uid, msg.GetSeatNo(), usr.Sid)
		return nil
	}

	tid := msg.GetTableID()
	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table not exists:%v", tid)
		return nil
	}
	//bulletid, userid, tableid, sid, vectorx, vectory, ratio, speed, fishid
	rs.OnShoot(msg.GetBulletID(), msg.GetUserID(), msg.GetTableID(), msg.GetSeatNo(),
		msg.GetVectorX(), msg.GetVectorY(), msg.GetRatio(), msg.GetSpeed(), msg.GetFishID())
	return nil
}

// N3080003 子弹碰到鱼
func (p *FishServer) N3080003(s *session.Session, msg *plr.N3080003, mid uint) error {
	uid := msg.GetUserID()
	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("user not exists:%v", uid)
		return nil
	}
	tid := msg.GetTableID()

	if usr.Tid != tid {
		glog.SErrorf("invalid session tid:%v,msg tid:%v", usr.Tid, tid)
		return nil
	}

	if usr.Sid != msg.GetSeatNo() {
		glog.SErrorf("invalid seat, uid:%v,msg seatno:%v,sessioin seatno:%v", uid, msg.GetSeatNo(), usr.Sid)
		return nil
	}

	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table not exists:%v", tid)
		return nil
	}
	//bulletid, userid, fishid, tableid, sid int32
	rs.OnShootFish(msg.GetBulletID(), msg.GetUserID(),
		msg.GetFishID(), msg.GetTableID(), msg.GetSeatNo())
	return nil
}

// N3080004 切换炮台倍率
func (p *FishServer) N3080004(s *session.Session, msg *plr.N3080004, mid uint) error {
	uid := msg.GetUserID()
	ratio := msg.GetRatio()
	if ratio <= 0 {
		glog.SErrorf("invalid param uid:%v, ratio:%v", uid, ratio)
		return nil
	}
	usr, ok := p.usrs[uid]
	if !ok {
		glog.SErrorf("invalid param uid:%v", uid)
		return nil
	}
	if ratio <= 0 {
		glog.SErrorf("invalid param uid:%v", uid)
		return nil
	}
	rs, ok := p.rs[usr.Tid]
	if !ok {
		glog.SErrorf("table RSession not exists:%v", usr.Tid)
		return nil
	}
	rs.OnChangeCannonRatio(uid, usr.Sid, ratio)
	return nil
}

func (p *FishServer) canAccess(u *model2.UserInfo, rid int32) (float64, error) {
	min, err := model.QueryRoomLimitAmount(rid, p.db)
	if err != nil {
		return min, fmt.Errorf("Cannot find room by rid:%v", rid)
	}
	if min == 0 {
		min = consts.RoomMinEnterAmount
	}

	if u.Wealth < min {
		return min, fmt.Errorf("User wealth:%v less then room limit:%v", u.Wealth, min)
	}

	if u.Status == 0 {
		return min, fmt.Errorf("User status:%v ", u.Status)
	}
	return min, nil
}

func (p *FishServer) doOutterSessionClosed(s *session.Session) {
	glog.SInfof("doOutterSessionClosed")
	for _, us := range p.usrs {
		if r, ok := p.rs[us.Tid]; ok {
			r.OnLeave(us.Uid)
		}
	}
}

func (p *FishServer) doInnerSessionClosed(tid, rid int32) {
	server.Invoke(func() {
		delete(p.rs, tid)
		if mgr, ok := p.mgr[rid]; ok {
			t := mgr.GetTable(tid)
			for _, v := range t.GetSeats() {
				if v.Status == consts.SeatStatusOk {
					delete(p.usrs, v.Uid)
				}
			}
			t.SitUpAll()
		}
	})

}

func (p *FishServer) playerLeave(uid string) {
	glog.SInfof("player leave uid :%v", uid)
	server.Invoke(func() {
		if seat, ok := p.usrs[uid]; ok {
			if mgr, ok := p.mgr[seat.Rid]; ok {
				t := mgr.GetTable(seat.Tid)
				if t != nil {
					t.SitUp(uid)
				}
			}
		}
		delete(p.usrs, uid)
	})
}
