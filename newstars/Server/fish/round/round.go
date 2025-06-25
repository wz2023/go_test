package round

import (
	randc "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/rand"
	"newstars/Protocol/plr"
	"newstars/Server/fish/conf"
	"newstars/Server/fish/consts"
	"newstars/Server/fish/model"
	consts2 "newstars/framework/consts"
	"newstars/framework/core/session"
	"newstars/framework/game_center"
	"newstars/framework/glog"
	"newstars/framework/util"
	"newstars/framework/util/decimal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SessionClosedHandler 回调函数
type SessionClosedHandler func(tid, rid int32)

// PlayerLeaveHandler 用户离开
type PlayerLeaveHandler func(uid string)

type RSession struct {
	enity      *session.Session
	db         *sql.DB
	rid        int32
	tid        int32
	status     int32
	ps         []*Player
	chFunction chan func()
	cb         SessionClosedHandler
	cbpl       PlayerLeaveHandler
	minRatio   float64
	maxRatio   float64
	quit       chan int
	// rName             string
	bossFish          []*model.FishKind
	fishKind          map[int32][]*model.FishKind
	typeclock         map[int32]int64
	fishNums          int
	bossIndex         int32
	fishIndex         int32
	fishRecord        map[int32]*model.FishRecord
	bullet            map[string]*model.BulletInfo
	skillType         map[string]*model.FishSkill
	skillRecord       map[string]*model.SkillRecord
	fishPath          map[int32]*model.FishPath
	fishPathIndex     []int32
	fishTideStartTime int64
	fishTideStopTime  int64
	bossStartTime     int64
	control           *RevenueControl
	leaveTimer        map[int]*time.Timer
	gameCommission    map[int]decimal.Decimal // 游戏平台对应的佣金
	inven             *model.Inventory
	tideID            int
	bPushedTide       bool //记录波浪是否推送,避免波浪push多次
	bPushedBoss       bool
	minEnter          float64
	excludedFishKind  map[int32]int32
	lastFreezeEndTime int64
	gameids           []int32
}

// NewRSession create
func NewRSession(s *session.Session, db *sql.DB, tid, rid int32, minEnter float64, control *RevenueControl, gameids []int32) *RSession {
	kinds, err := model.QueryAllFishKinds(db)
	if err != nil {
		return nil
	}

	mapkind := make(map[int32][]*model.FishKind)

	typeclock := make(map[int32]int64)

	for _, v := range kinds {
		if v.KindType <= 0 {
			continue
		}

		if _, ok := mapkind[v.KindType]; !ok {
			mapkind[v.KindType] = make([]*model.FishKind, 0)
		}

		if _, ok := typeclock[v.KindType]; !ok {

			typeclock[v.KindType] = 0
		}
		mapkind[v.KindType] = append(mapkind[v.KindType], v)
	}

	bossFish, err := model.QueryBossFishKinds(db)
	if err != nil {
		return nil
	}

	skills, err := model.QueryFishSkill(db)
	if err != nil {
		return nil
	}

	mapskills := make(map[string]*model.FishSkill, 0)
	for _, v := range skills {
		key := fmt.Sprintf("%d_%d", v.GameRoomID, v.SkillType)
		mapskills[key] = v
	}

	paths, err := model.QueryFishPath(db)
	if err != nil {
		return nil
	}

	baseamount, commission, err := model.QueryRoomInfo(rid, db)
	if err != nil {
		glog.SErrorf("QueryRoomInfo failed.err:%v", err)
		return nil
	}

	pathmap := make(map[int32]*model.FishPath, 0)
	pathIndex := make([]int32, len(paths))
	for i, v := range paths {
		pathmap[v.ID] = v
		pathIndex[i] = v.ID
	}
	inventorys, err := model.QueryInventory(db)
	if err != nil {
		glog.SErrorf("QueryInventory failed:%v", inventorys)
	}

	kidmap := make(map[int32]int32)
	exKindids, err := model.QueryExcludeFishkinds(db, rid)
	if err != nil {
		glog.SErrorf("fish_room_config failed:%v", inventorys)
	} else {
		if exKindids != "" {
			exkindArr := strings.Split(exKindids, ",")
			for _, kv := range exkindArr {
				ikid, err := strconv.ParseInt(kv, 10, 0)
				if err == nil {
					kidmap[int32(ikid)] = int32(ikid)
				}
			}

		}
	}
	item := &model.Inventory{}
	for i := range inventorys {
		if inventorys[i].RoomID == rid {
			item = inventorys[i]
			break
		}
	}

	var maxRatio float64

	if baseamount == consts.SpecialRatio {
		maxRatio, _ = decimal.NewFromFloat(baseamount).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Float64()
	} else {
		maxRatio, _ = decimal.NewFromFloat(baseamount).Mul(decimal.NewFromFloat(consts.MaxMinRatio)).Float64()
	}

	allgid, err := model.QueryAllGameID(db) // 获取游戏ID列表

	gameCommission := make(map[int]decimal.Decimal)
	for _, gid := range allgid {
		cms := decimal.NewFromFloat(commission).Add(decimal.NewFromFloat(conf.Conf.ExtraCommission[gid]))
		gameCommission[gid] = cms
	}

	return &RSession{
		chFunction:       make(chan func(), 1024),
		enity:            s,
		db:               db,
		tid:              tid,
		rid:              rid,
		ps:               make([]*Player, consts.FishSeatNumbers),
		fishKind:         mapkind,
		bossFish:         bossFish,
		fishRecord:       make(map[int32]*model.FishRecord, 0),
		skillType:        mapskills,
		skillRecord:      make(map[string]*model.SkillRecord, 0),
		bullet:           make(map[string]*model.BulletInfo, 0),
		fishPath:         pathmap,
		fishPathIndex:    pathIndex,
		fishNums:         len(kinds),
		minRatio:         baseamount,
		maxRatio:         maxRatio,
		control:          control,
		typeclock:        typeclock,
		leaveTimer:       make(map[int]*time.Timer),
		gameCommission:   gameCommission,
		inven:            item,
		tideID:           1, //默认从飞机鱼潮开始
		minEnter:         minEnter,
		excludedFishKind: kidmap,
		gameids:          gameids,
	}
}

// RefreshGameids refresh gameids
func (p *RSession) RefreshGameids(gids []int32) {
	p.gameids = gids
}

// Go run
func (p *RSession) Go() {
	p.initFishes()
	go p.run()
	// time.AfterFunc(300*time.Millisecond, func() {
	p.Invoke(p.doRoundStart)
	// })

}

// Invoke do in goroutine
func (p *RSession) Invoke(fn func()) {
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

func (p *RSession) run() {
	settleTicker := time.NewTicker(60 * time.Second)
	// fishTicker := time.NewTicker(time.Second)
	pushTicker := time.NewTicker(time.Second)
	// tideTicker := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-settleTicker.C:
			p.doSettleTicker()
		// case <-fishTicker.C:
		case <-pushTicker.C:
			p.doTicker()
			p.pushAllFish()
			// p.pushRealtimeTideFish()
		// case <-tideTicker.C:
		case fn := <-p.chFunction:
			pinvoke(fn)
		case <-p.quit:
			return
		}
	}
}

func (p *RSession) checkNewPlayer(pl *Player) bool {
	if !pl.isnewPlayer {
		return false
	}

	if pl.amount > consts.NewLayerProtectAmount {
		pl.isnewPlayer = false
		return false
	}
	ontimes := p.getCurMillisecond() - pl.entertime + pl.onlinetimes
	if ontimes > consts.NewLayerProtectTime {
		pl.isnewPlayer = false
		return false
	}
	return true
}

// 检查是否充值过
func (p *RSession) checkPlayerPay(pl *Player) {
	vvip := 0
	err := p.db.QueryRow(`select ifnull(sum(vvipvalue),0) from vip_level_t where userid=?`, pl.uid).Scan(&vvip)
	if err != nil {
		glog.SErrorf("query vip_level_t failed.err:%v", err)
		return
	}
	if vvip > 0 {
		pl.isnewPlayer = false
		pl.isPayPlayer = true
	}
}

// 玩家个人概率
func (p *RSession) getPlayerExtraRate(pl *Player) float64 {
	//新手玩家
	if pl.isnewPlayer {
		if pl.amount < consts.NewplayerLessAmount {
			return 0.5
		}
		return 0.1
	}

	//普通玩家
	inv, _ := pl.inven.Float64()
	// glog.SInfof("个人库存 uid %v, 库存:%v 历史税收:%v", pl.uid, inv, pl.hisrevenue)
	if inv > 50+pl.cannonRatio*200 {
		if pl.hisrevenue.GreaterThan(pl.inven) {
			return 0.08
		}
	}

	//非充值玩家
	if !pl.isPayPlayer {
		if pl.amount+pl.bankAmount > 5 {
			return -0.30
		} else if pl.amount+pl.bankAmount > 4 {
			return -0.10
		}
	}

	return 0.0

}

func (p *RSession) savefishPlayerInfo(pl Player) {
	ontimes := p.getCurMillisecond() - pl.entertime + pl.onlinetimes
	inven, _ := pl.inven.Float64()
	hisrev := pl.totalRevenue.Add(pl.revenue)
	isnewplayer := 0
	if pl.isnewPlayer {
		isnewplayer = 1
	}
	_, err := p.db.Exec(`update fish_user_extend_t set onlinetimes=? ,inven=?,hisrevenue=hisrevenue+?,isnewplayer=? where userid=? `, ontimes, inven, hisrev, isnewplayer, pl.uid)
	if err != nil {
		glog.SErrorf(`update  fish_user_extend_t failed,err:%v`, err)
	}
}

func (p *RSession) doSettleTicker() {
	p.doSettle()
}

func (p *RSession) pushAllFish() {
	if p.status != consts.GameStatusReadyFishTide {
		if _, isok := p.skillRecord["0"]; !isok {
			fishes := make([]*plr.P3080015_FishInfo, 0)
			now := p.getCurMillisecond()
			for _, v := range p.fishRecord {
				var deadtime int64
				if path, ok := p.fishPath[v.Path]; ok {
					deadtime = path.Deadtime
				}
				if now >= v.StartTime && now-v.StartTime < deadtime {
					f := &plr.P3080015_FishInfo{
						FishID:      v.FishID,
						Path:        v.Path,
						ElapsedTime: now - v.StartTime,
					}
					fishes = append(fishes, f)
				}
			}
			if len(fishes) != 0 {
				push := &plr.P3080015{}
				push.Fishes = fishes
				push.RoomID = p.rid
				push.TableID = p.tid
				p.enity.Push("P3080015", push)
			}
		}
	}
}

//鱼潮时需要加快同步
// func (p *RSession) pushRealtimeTideFish() {
// 	if p.status == consts.GameStatusStartingFishTide {
// 		if _, isok := p.skillRecord["0"]; !isok {
// 			fishes := make([]int32, 0)
// 			now := p.getCurMillisecond()
// 			for _, v := range p.fishRecord {
// 				if now >= v.StartTime && !v.IsPushed {
// 					fishes = append(fishes, v.FishID)
// 					v.IsPushed = true
// 				}
// 			}
// 			if len(fishes) != 0 {
// 				push := &plr.P3080017{}
// 				push.Fishes = fishes
// 				push.TableID = p.tid
// 				p.enity.Push("P3080017", push)
// 			}
// 		}
// 	}
// }

func (p *RSession) doClearFish() {
	now := p.getCurMillisecond()
	for i, v := range p.fishRecord {
		var deadtime int64 = 60000
		if path, ok := p.fishPath[v.Path]; ok {
			deadtime = path.Deadtime
		}
		if now-v.StartTime > deadtime {
			delete(p.fishRecord, i)
		}
	}
}

func (p *RSession) doRoundError() {
	if p.cb != nil {
		p.cb(p.tid, p.rid)
	}
}

func (p *RSession) doOffline(uid string) {
	glog.SInfof("user offline:%v", uid)
	index := -1
	var pl *Player
	for i, v := range p.ps {
		if v != nil {
			if v.uid == uid {
				index = i
				pl = v
			}
		}
	}

	if index != -1 {
		ppl := *pl
		if pl.amount < p.minEnter {
			//破产离线马上踢走
			p.doLeave(uid)
		} else {
			p.playerSettle(ppl)
			p.leaveTimer[index] = time.AfterFunc(30*time.Second, func() {
				p.Invoke(func() {
					p.doLeave(uid)
				})
			})
		}
	}
}

func (p *RSession) doLeave(uid string) {
	glog.SInfof("user leave uid:%v", uid)
	index := -1
	var pl *Player
	for i, v := range p.ps {
		if v != nil {
			if v.uid == uid {
				index = i
				pl = v
			}
		}
	}
	if index == -1 {
		return
	}

	push := &plr.P3080003{}
	push.TableID = p.tid
	push.SeatNo = p.ps[index].sid
	push.UserID = p.ps[index].uid
	p.enity.Push("P3080003", push)

	if p.cbpl != nil {
		p.cbpl(uid)
	}

	p.savefishPlayerInfo(*pl)
	p.saveNewPoolAmount()
	sid := p.ps[index].sid

	//把其他玩家加入到自己的同局玩家
	for _, v := range p.ps {
		if v != nil && len(v.uid) > 0 {
			if v.uid != uid {
				if v.isAmountChanged {
					pl.AddRoundPlayer(v.uid, v.sid, v.nick)
				}
			}
		}
	}
	if pl.isAmountChanged {
		//把自己加入到其他人人的同局玩家
		for _, v := range p.ps {
			if v != nil && len(v.uid) > 0 {
				if v.uid != uid {
					v.AddRoundPlayer(pl.uid, pl.sid, pl.nick)
				}
			}
		}
		pl.roundRecord.EndAmount = p.ps[sid].amount
		pl.roundRecord.PayoffValue, _ = decimal.NewFromFloat(pl.roundRecord.EndAmount).Sub(decimal.NewFromFloat(pl.roundRecord.StartAmount)).Float64()
		pl.roundRecord.SettleTime = time.Now().Unix()
		glog.SInfof("PayoffValue:%v", pl.roundRecord.PayoffValue)
		pl.FillFishInfo(p.db)
		//处理消耗
		rbs := make([]*RecordBullet, 0)
		for _, v := range pl.RecordBullet {
			rbs = append(rbs, v)
		}
		sort.Slice(rbs, func(i, j int) bool {
			if rbs[i].RecordType < rbs[j].RecordType {
				return true
			}
			if rbs[i].RecordType == rbs[j].RecordType {
				return rbs[i].BaseAmount < rbs[j].BaseAmount
			}
			return false
		})

		//处理同局玩家
		rps := make([]RecordPlayer, 0)
		for _, v := range pl.RoundPlayers {
			rps = append(rps, v)
		}
		sort.Slice(rps, func(i, j int) bool {
			return rps[i].Sid < rps[j].Sid
		})

		//处理奖励
		rrs := make([]*RecordReward, 0)
		for _, v := range pl.Rewards {
			rrs = append(rrs, v)
		}
		sort.Slice(rrs, func(i, j int) bool {
			return rrs[i].KindID < rrs[j].KindID
		})
		pl.roundRecord.RecordBullet = rbs
		pl.roundRecord.RoundPlayers = rps
		pl.roundRecord.Rewards = rrs
		// pl.roundRecord.DumpInfo()
		p.flushRoundRecordToDB(pl.roundRecord)

		// var buf = make([]byte, 0)
		// glog.SInfof("数据库查询局号:%v", pl.roundRecord.RoundName)
		// p.db.QueryRow("select history from  round_record_t where roundname=?",
		// 	pl.roundRecord.RoundName).Scan(&buf)
		// dbrecord, err := Decode(buf)
		// if err == nil {
		// 	dbrecord.DumpInfo()
		// } else {
		// 	glog.SErrorf("Decode failed.err:%v", err)
		// }

	} else {
		err := model.DeteteRecordAmount(pl.settlewid, pl.uid, p.db)
		if err != nil {
			glog.SErrorf("delete recordAmount failed for db. id:%v,uid:%v,err:%v", pl.settlewid, pl.uid, err)
		}
	}
	ppl := *pl
	p.playerSettle(ppl)
	if pl.isAmountChanged {
		_, err := p.db.Exec(`insert into settlement_fish_t (roundcode,gametableid,userid,betamount
		,bettype,actualamount,odds,payoffvalue,results,bettime,settletime,revenue) values(?,?,?,?,?,?,?,?,?,?,?,?)`,
			pl.rName, p.tid, pl.uid, pl.totalActualamount, 1, pl.totalActualamount, 1, pl.totalPayoff,
			"", time.Now().Unix(), time.Now().Unix(), pl.totalRevenue)
		if err != nil {
			glog.SErrorf("insert into settlement_fish_t  failed.Db error:%v", err)
		}

		//tx, err := p.db.Begin()
		//if err != nil {
		//	glog.SErrorf("AddLuckyProfit begin tx failed.Db error:%v", err)
		//}
		//ams, _ := pl.totalPayoff.Float64()
		//err = model.AddLuckyProfit(pl.uid, ams, consts.FishKindID, p.gameids, tx)
		//if err != nil {
		//	tx.Rollback()
		//	glog.SErrorf("AddLuckyProfit failed.uid:%v,error:%v", pl.uid, err)
		//} else {
		//	tx.Commit()
		//}

	}
	p.ps[index] = nil

	//清除锁定和加速技能
	recordkey := fmt.Sprintf("%d_%d_%d", p.rid, 1, sid)
	delete(p.skillRecord, recordkey)
	recordkey = fmt.Sprintf("%d_%d_%d", p.rid, 2, sid)
	delete(p.skillRecord, recordkey)
	//删除子弹
	keyprex := fmt.Sprintf("%d_%d_", p.tid, sid)
	deleteKeys := make([]string, 0)
	for k := range p.bullet {
		if strings.Contains(k, keyprex) {
			deleteKeys = append(deleteKeys, k)
		}
	}
	for _, v := range deleteKeys {
		delete(p.bullet, v)
	}

	if p.getPlayNumbers() == 0 {
		if p.cb != nil {
			p.cb(p.tid, p.rid)
		}
		p.quit <- 0
	}
}

func (p *RSession) flushRoundRecordToDB(record RoundRecord) {
	buf, err := Encode(record)
	if err != nil {
		glog.SErrorf("flushLogToDB error %v", err)
	}
	_, err = p.db.Exec(`insert into round_record_t (roundname,history,settletime,kind) values(?,?,?,?)`, record.RoundName, buf, time.Now().Unix(), consts.FishKindID)
	if err != nil {
		glog.SErrorf("flush db failed.%v", err)
	}
}

func (p *RSession) getPlayNumbers() int {
	size := 0
	for _, v := range p.ps {
		if v != nil && len(v.uid) > 0 {
			size++
		}
	}
	return size
}

// OnCallBack to main routine
func (p *RSession) OnCallBack(cb SessionClosedHandler, cbpl PlayerLeaveHandler) {
	p.cb = cb
	p.cbpl = cbpl
}

func (p *RSession) OnEnter(uid string, sid int32, mid uint) {
	p.Invoke(func() {
		p.doEnter(uid, sid, mid)
	})
}

func (p *RSession) OnReEnter(uid string, sid int32, mid uint) {
	p.Invoke(func() {
		p.doReEnter(uid, sid, mid)
	})
}

func (p *RSession) OnLeave(uid string) {
	p.Invoke(func() {
		p.doLeave(uid)
	})
}

func (p *RSession) OnOffline(uid string) {
	p.Invoke(func() {
		p.doOffline(uid)
	})
}

func (p *RSession) OnShoot(bulletid int32, userid string, tableid, sid int32, vectorx, vectory, ratio float64, speed, fishid int32) {
	p.Invoke(func() {
		p.doShoot(bulletid, userid, tableid, sid, vectorx, vectory, ratio, speed, fishid)
	})
}

func (p *RSession) OnShootFish(bulletid int32, userid string, fishid, tableid, sid int32) {
	p.Invoke(func() {
		p.doShootFish(bulletid, userid, fishid, tableid, sid)
	})
}

func (p *RSession) OnSkill(userid string, tableid, seatno, skilltype int32, opType int32, mid uint) {
	p.Invoke(func() {
		p.doSkill(userid, tableid, seatno, skilltype, opType, mid)
	})
}

func (p *RSession) OnChangeCannonRatio(uid string, sid int32, ratio float64) {
	p.Invoke(func() {
		p.doChangeCannonRatio(uid, sid, ratio)
	})
}

func (p *RSession) OnChangeCannonID(uid string, sid, cannonID int32, mid uint) {
	p.Invoke(func() {
		p.doChangeCannonID(uid, sid, cannonID, mid)
	})
}

func (p *RSession) OnPurchaseCannon(uid string, sid, cannonID int32, mid uint) {
	p.Invoke(func() {
		p.doPurchaseCannon(uid, sid, cannonID, mid)
	})
}

func (p *RSession) doEnter(uid string, sid int32, mid uint) {
	rsp := &plr.S3080002{}
	if p.leaveTimer[int(sid)] != nil {
		p.leaveTimer[int(sid)].Stop()
	}

	if err := p.sitdown(uid, sid); err != nil {
		glog.SErrorf("sitdown failed,uid:%v,sid:%v,err:%v", uid, sid, err)
		rsp.RetCode = consts.ErrorSitDown
		p.enity.Response(rsp, mid)
		return
	}

	rsp.RoomID = p.rid
	rsp.TableID = p.tid
	rsp.SeatNo = sid
	//桌内其他玩家信息
	seatInfos := make([]*plr.S3080002_SeatInfo, 0)
	for i := range p.ps {
		if p.ps[i] != nil {
			pl := p.ps[i]
			if len(pl.uid) > 0 {
				seat := &plr.S3080002_SeatInfo{}
				seat.FaceID = pl.faceid
				seat.SeatNo = pl.sid
				seat.TableID = p.tid
				seat.UserAmount = pl.amount
				seat.UserID = pl.uid
				seat.UserName = pl.name
				seat.CannonID = pl.cannonID
				seat.Ratio = pl.cannonRatio
				seat.VipLevel = pl.vipLevel
				seat.FaceFrameID = pl.faceFrameID
				seatInfos = append(seatInfos, seat)
			}
		}
	}

	rsp.Seats = seatInfos
	if len(seatInfos) == 0 {
		glog.SWarnf("empty seat uid:%v,sid:%v", uid, sid)
	}
	//鱼信息
	now := p.getCurMillisecond()
	fishes := make([]*plr.S3080002_FishInfo, 0)

	//技能信息
	var freezeCountDown int64
	skills := make([]*plr.S3080002_SkillInfo, 0)
	for _, v := range p.skillRecord {
		if v.SkillType == 0 {
			freezeCountDown = int64(v.FreezeTime) - now + v.StartTime
		}
		s := &plr.S3080002_SkillInfo{}
		s.UserID = v.UserID
		s.TableID = v.TableID
		s.SeatNo = v.SeatNo
		s.SkillType = v.SkillType
		skills = append(skills, s)

	}
	rsp.Skills = skills

	//鱼潮准备时间不需要放鱼
	if p.status != consts.GameStatusReadyFishTide {
		for _, v := range p.fishRecord {
			var deadtime int64
			if path, ok := p.fishPath[v.Path]; ok {
				deadtime = path.Deadtime
			}
			if now-v.StartTime+freezeCountDown < deadtime {
				f := &plr.S3080002_FishInfo{}
				f.FishID = v.FishID
				f.KindID = v.KindID
				f.Path = v.Path
				f.StartTime = v.StartTime - freezeCountDown
				f.CurrentTime = now
				fishes = append(fishes, f)
				if p.status != consts.GameStatusStartingFishTide {
					if f.StartTime > f.CurrentTime {
						glog.SWarnf("wrong time fishid:%v,starttime:%v,endtime:%v", v.FishID, f.StartTime, f.CurrentTime)
					}
				}
			}
			// else {
			// 	glog.SInfof("过滤鱼 fishid:%v,时间:%v,%v,当前时间:%v,开始时间:%v", v.FishID, now-v.StartTime+freezeCountDown, deadtime, now, v.StartTime)
			// }
		}
		rsp.Fishes = fishes
	}

	rsp.RoomID = p.rid
	rsp.TableID = p.tid
	rsp.TideID = int32(p.tideID)
	p.enity.Response(rsp, mid)

	tx, err := p.db.Begin()
	if err != nil {
		glog.SErrorf("doSettle failed.Db error:%v", err)
		return
	}

	settlewid, err := model.InsertRecordAmount(uid, consts.FishKindID, p.ps[sid].rName, tx)
	if err != nil {
		glog.SErrorf("insert recordAmount failed for db. err: %v", err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
	p.ps[sid].settlewid = settlewid

	for i := range p.ps {
		if p.ps[i] != nil {
			if p.ps[i].sid == sid {
				push := &plr.P3080002{}
				push.FaceID = p.ps[i].faceid
				push.SeatNo = p.ps[i].sid
				push.TableID = p.tid
				push.UserWealth = p.ps[i].amount
				push.UserID = p.ps[i].uid
				push.UserName = p.ps[i].name
				push.CannonID = p.ps[i].cannonID
				push.Ratio = p.ps[i].cannonRatio
				push.VipLevel = p.ps[i].vipLevel
				push.FaceFrameID = p.ps[i].faceFrameID
				p.enity.Push("P3080002", push)
				break
			}
		}
	}
}

func (p *RSession) doReEnter(uid string, sid int32, mid uint) {
	rsp := &plr.S3080008{}
	if p.leaveTimer[int(sid)] != nil {
		p.leaveTimer[int(sid)].Stop()
	}

	rsp.RoomID = p.rid
	rsp.TableID = p.tid
	rsp.SeatNo = sid
	//桌内其他玩家信息
	seatInfos := make([]*plr.S3080008_SeatInfo, 0)
	for i := range p.ps {
		if p.ps[i] != nil {
			pl := p.ps[i]
			if len(pl.uid) > 0 {
				seat := &plr.S3080008_SeatInfo{}
				seat.FaceID = pl.faceid
				seat.SeatNo = pl.sid
				seat.TableID = p.tid
				seat.UserAmount = pl.amount
				seat.UserID = pl.uid
				seat.UserName = pl.name
				seat.CannonID = pl.cannonID
				seat.Ratio = pl.cannonRatio
				seat.VipLevel = pl.vipLevel
				seat.FaceFrameID = pl.faceFrameID
				seatInfos = append(seatInfos, seat)
			}
		}
	}
	rsp.Seats = seatInfos
	if len(seatInfos) == 0 {
		glog.SWarnf("empty seat uid:%v,sid:%v", uid, sid)
	}
	//鱼信息
	now := p.getCurMillisecond()
	fishes := make([]*plr.S3080008_FishInfo, 0)

	//技能信息
	var freezeCountDown int64
	skills := make([]*plr.S3080008_SkillInfo, 0)
	for _, v := range p.skillRecord {
		if v.SkillType == 0 {
			freezeCountDown = int64(v.FreezeTime) - now + v.StartTime
		}
		s := &plr.S3080008_SkillInfo{}
		s.UserID = v.UserID
		s.TableID = v.TableID
		s.SeatNo = v.SeatNo
		// s.FreezeTime = v.FreezeTime
		// s.CountDonw = freezeCountDown
		s.SkillType = v.SkillType
		skills = append(skills, s)
	}
	rsp.Skills = skills
	//鱼潮准备时间不需要放鱼
	if p.status != consts.GameStatusReadyFishTide {
		for _, v := range p.fishRecord {
			var deadtime int64
			if path, ok := p.fishPath[v.Path]; ok {
				deadtime = path.Deadtime
			}
			if now-v.StartTime+freezeCountDown < deadtime {
				f := &plr.S3080008_FishInfo{}
				f.FishID = v.FishID
				f.KindID = v.KindID
				f.Path = v.Path
				f.StartTime = v.StartTime - freezeCountDown
				f.CurrentTime = now
				fishes = append(fishes, f)
				if p.status != consts.GameStatusStartingFishTide {
					if f.StartTime > f.CurrentTime {
						glog.SWarnf("wrong time fishid:%v,starttime:%v,endtime:%v", v.FishID, f.StartTime, f.CurrentTime)
					}
				}
			}
			// else {
			// 	glog.SInfof("过滤鱼 fishid:%v,时间:%v,%v,当前时间:%v,开始时间:%v", v.FishID, now-v.StartTime+freezeCountDown, deadtime, now, v.StartTime)
			// }
		}
		rsp.Fishes = fishes
	}

	rsp.RoomID = p.rid
	rsp.TableID = p.tid
	rsp.TideID = int32(p.tideID)
	p.enity.Response(rsp, mid)

	for i := range p.ps {
		if p.ps[i] != nil {
			if p.ps[i].sid == sid {
				push := &plr.P3080002{}
				push.FaceID = p.ps[i].faceid
				push.SeatNo = p.ps[i].sid
				push.TableID = p.tid
				push.UserWealth = p.ps[i].amount
				push.UserID = p.ps[i].uid
				push.UserName = p.ps[i].name
				push.CannonID = p.ps[i].cannonID
				push.Ratio = p.ps[i].cannonRatio
				push.VipLevel = p.ps[i].vipLevel
				push.FaceFrameID = p.ps[i].faceFrameID
				p.enity.Push("P3080002", push)
				break
			}
		}
	}
}

func (p *RSession) sitdown(uid string, sid int32) error {
	if sid < 0 || sid > consts.FishSeatNumbers-1 {
		return fmt.Errorf("Invalid sid:%v", sid)
	}
	//u, err := model.QueryUserInfo(uid, p.db)
	//if err != nil {
	//	glog.SErrorf("query userinfo failed err:%v,", err)
	//	return err
	//}

	u, err := game_center.GetUserInfoByID(uid)
	if err != nil {
		glog.SErrorf("query userinfo failed err:%v,", err)
		return err
	}

	exists, err := model.CheckUserExtendExists(uid, p.db)
	if err != nil {
		glog.SErrorf("check userextend failed err:%v,", err)
		return err
	}
	if !exists {
		err = model.InitUserExtend(uid, p.db)
		if err != nil {
			glog.SErrorf("init userextend failed err:%v,", err)
			return err
		}
	}

	exusr, err := model.QueryUserExtend(uid, p.db)
	if err != nil {
		glog.SErrorf("query userextend failed err:%v,", err)
		return err
	}

	//不在房间的倍率范围内，设为最小的倍率
	if exusr.CurrCannonRatio <= 0 || exusr.CurrCannonRatio < p.minRatio || exusr.CurrCannonRatio > p.maxRatio {
		err = model.UpdateUserCannonRate(p.db, p.minRatio, uid)
		if err != nil {
			return err
		}
		//再查一次
		exusr, err = model.QueryUserExtend(uid, p.db)
		if err != nil {
			return err
		}
	}

	var dbcid int32 = -1
	err = p.db.QueryRow(`select cannonid from fish_user_cannon_t where endtime>? and userid=? 
		and cannonid=? union select id from fish_cannon_t where id=? and isneedbuy=0`,
		time.Now().Unix(), uid, exusr.CurrCannonID, exusr.CurrCannonID).Scan(&dbcid)
	if err != nil {
		if err == sql.ErrNoRows {
			//如果过期或者没设置过，则设置为默认炮台
			err = model.UpdateUserCannonId(p.db, consts.DefaultConnonID, uid)
			if err != nil {
				glog.SErrorf("updateUser cannonId failed err:%v,", err)
				return err
			}
			exusr, err = model.QueryUserExtend(uid, p.db)
			if err != nil {
				glog.SErrorf("query UserExtend failed err:%v,", err)
				return err
			}
		} else {
			glog.SErrorf("query fish_user_cannon_t failed err:%v,", err)
			return err
		}
	}

	var hisrev float64
	if exusr.HisRevenue == 0 {
		err := p.db.QueryRow(`select ifnull(sum(revenue),0) from settlement_fish_t where userid=?`, uid).Scan(&hisrev)
		if err != nil {
			glog.SErrorf("query settlement_fish_t failed uid:%v,err:%v,", uid, err)
			return err
		}
		_, err = p.db.Exec(`update fish_user_extend_t set hisrevenue=? where  userid=?`, hisrev, uid)
		if err != nil {
			glog.SErrorf("update hisrevenue failed uid:%v,err:%v,", uid, err)
			return err
		}
		exusr.HisRevenue = hisrev
	}

	p.ps[sid] = NewPlayer(uid, sid, u.FaceID, u.FaceFrameID, exusr.CurrCannonID, u.Wealth, p.minRatio, u.DisPlayName, u.NickName, int(u.GameID))
	p.ps[sid].currency = u.Currency
	p.ps[sid].hisrevenue = decimal.NewFromFloat(exusr.HisRevenue)
	p.ps[sid].inven = decimal.NewFromFloat(exusr.Inven)
	p.ps[sid].isnewPlayer = (exusr.IsNewplayer == 1)
	p.ps[sid].onlinetimes = exusr.Onlinetimes
	p.ps[sid].entertime = p.getCurMillisecond()
	p.ps[sid].rName = fmt.Sprintf("rn%10d:%d:%d:%s", time.Now().Unix(), p.rid, p.tid, uid)
	p.ps[sid].sessionID = fmt.Sprintf("%d:%d:%s", p.rid, p.tid, uid)
	p.ps[sid].roundRecord = NewRoundRecord(p.ps[sid].rName, p.minRatio, p.rid, uid, sid, u.NickName)
	p.ps[sid].roundRecord.StartAmount = u.Wealth
	//p.ps[sid].vipLevel, _, _ = model.QueryVipLevel(uid, p.db)
	// if p.ps[sid].isnewPlayer {
	p.checkPlayerPay(p.ps[sid])
	// }
	//var dbBankamount float64
	//if err == nil {
	//	err = p.db.QueryRow(`select bankamount from userbank_t where userid =?`, uid).Scan(&dbBankamount)
	//	if err != nil {
	//		glog.SErrorf("query bankamount failed.Db error:%v", err)
	//		return err
	//	}
	//}
	//p.ps[sid].roundRecord.BankAmount = dbBankamount
	//p.ps[sid].bankAmount = dbBankamount
	return nil
}

func (p *RSession) doRoundStart() {
	var seed int64
	binary.Read(randc.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
	// p.rName = fmt.Sprintf("rn%10d_%d_%d", time.Now().Unix(), p.rid, p.tid)
	glog.SInfof("doRoundStart:%v start", time.Now().Unix())
	// p.doRandomFish()
	p.status = consts.GameStatusFreeFish
	p.fishTideStartTime = p.getCurMillisecond() + consts.TimeForTideInterval
	p.bossStartTime = p.getCurMillisecond() + consts.TimeForBossInterval
}

func (p *RSession) initFishes() {
	p.doRandomFish()
	for i := 0; i < 8; i++ {
		for _, v := range p.fishRecord {
			v.StartTime = v.StartTime - int64(rand.Intn(2))*1000 - 1000
		}
		for i, v := range p.typeclock {
			p.typeclock[i] = v - 6000
		}
		p.doRandomFish()
	}
	now := p.getCurMillisecond()
	delids := make([]int32, 0)
	for _, v := range p.fishRecord {
		var deadtime int64
		if path, ok := p.fishPath[v.Path]; ok {
			deadtime = path.Deadtime
		}
		if now-v.StartTime > deadtime {
			delids = append(delids, v.FishID)
		}
	}
	for _, v := range delids {
		delete(p.fishRecord, v)
	}
}

func (p *RSession) isCapture(bullet model.BulletInfo, fish model.FishRecord, uextraRate float64, isnewplayer bool) bool {

	bulletCoin := bullet.Wealth
	fishscore := fish.Score
	if fishscore <= 0 {
		glog.SErrorf("invalid fish score:%v", fishscore)
		return false
	}

	var deadtime int64
	if path, isOk := p.fishPath[fish.Path]; isOk {
		deadtime = path.Deadtime
	}

	now := p.getCurMillisecond()
	var freezeCountDown int64
	if sk, isok := p.skillRecord["0"]; isok {
		freezeCountDown = int64(sk.FreezeTime) - now + sk.StartTime
	}
	fishStart := fish.StartTime - freezeCountDown
	if now >= fishStart && now-fishStart < deadtime {
		//检查库存
		if fish.KindID == 28 {
			var totalScore float64
			for _, v := range p.fishRecord {
				if v.Score <= 7 {
					totalScore = totalScore + v.Score
				}
			}
			payWealth := totalScore * bulletCoin
			if payWealth > p.control.GetToTal(p.rid) {
				return false
			}
		} else {
			payWealth := fishscore * bulletCoin
			if payWealth > p.control.GetToTal(p.rid) {
				return false
			}
		}
		rate := rand.Float64() * fishscore
		extrainv := 1.0
		if !isnewplayer && p.rid != consts.NewPalyerRoomid {
			extrainv = (1 + p.control.getExtraCaptureRate(p.rid))
		}
		baseinv := p.control.GetRoomCaptureRate(p.rid)
		uinv := (1 + uextraRate)
		totalinv := baseinv * extrainv * uinv
		glog.SInfof("基础概率:%v,个人概率:%v,库存概率:%v，总概率:%v 是否捕获:%v", baseinv, uinv, extrainv, totalinv, rate < totalinv)
		return rate < totalinv
	}
	glog.SWarnf("expired fishid %v,starttime:%v,deadtime:%v,bulletid:%v,wealth:%v", fish.FishID, fishStart, deadtime, bullet.BulletID, bullet.Wealth)
	return false
}

func (p *RSession) getTypeNums(kindtype int32) int {
	var count = 0
	for _, v := range p.fishRecord {
		if v.KindType == kindtype {
			count++
		}
	}
	return count
}

// 过滤出鱼
func (p *RSession) filterFishKinds(kinds []*model.FishKind, filternums int32) []*model.FishKind {
	km := make(map[int32]int32)
	now := p.getCurMillisecond()
	for _, v := range p.fishRecord {
		if _, ok := km[v.KindID]; ok {
			km[v.KindID] = km[v.KindID] + 1
		} else {
			km[v.KindID] = 1
		}
		//如果出鱼间隔还不到，也过滤掉
		if now-v.StartTime < int64(v.Interval) {
			if v.KindType != 1 {
				km[v.KindID] = filternums
			}
		}
	}

	newkinds := make([]*model.FishKind, 0)
	for _, v := range kinds {
		if km[v.ID] < filternums {
			if _, isok := p.excludedFishKind[v.ID]; !isok {
				newkinds = append(newkinds, v)
			}
		}
	}
	return newkinds
}

// 过滤鱼路径
func (p *RSession) filterFishPath(intervalSec int32) map[int32]int32 {
	now := p.getCurMillisecond()
	sec := int64(intervalSec * 1000)
	pmap := make(map[int32]int32)
	for _, v := range p.fishRecord {
		if now-v.StartTime < sec {
			pmap[v.Path] = v.Path
		}
	}
	return pmap
}

func (p *RSession) doTicker() {
	p.doClearFish()
	p.getCurState()
	switch p.status {
	case consts.GameStatusFreeFish:
		p.doRandomFish()
	case consts.GameStatusReadyBoss:
		p.doBossFishReady()
		p.doBossFish()
	case consts.GameStatusReadyFishTide:
		p.doFishTideReady()
	case consts.GameStatusStartFishTide:
		p.doFishTideStart()
	}
}
func (p *RSession) doRandomFish() {

	if _, isok := p.skillRecord["0"]; !isok {

		for t, tc := range p.typeclock {

			if t == 4 {
				continue
			}
			push := &plr.P3080001{}
			if tc < p.getCurMillisecond() {
				seednums := rand.Intn(consts.MaxFishSeed + 1)
				numok := true
				if t == 2 {
					seednums = 2
				} else if t == 3 {
					nums := p.getTypeNums(t)
					if nums >= consts.MaxLargeFish {
						numok = false
					}
					seednums = 1
				} else if t == 6 {
					nums := p.getTypeNums(t)
					if nums >= 1 {
						numok = false
					}
					seednums = 1
				}
				if numok {
					arr := make([]*plr.P3080001_FishInfo, 0)
					fkinds := p.fishKind[t]
					switch t {
					case 1:
						//小鱼
						fkinds = p.filterFishKinds(fkinds, 5)
					case 6:
						//黄色鱼
						fkinds = p.filterFishKinds(fkinds, 1)
					default:
						fkinds = p.filterFishKinds(fkinds, 2)
					}

					if len(fkinds) == 0 {
						continue
					}

					excludes := p.filterFishPath(8)
					for i := 0; i < seednums; i++ {
						fishIndex := rand.Intn(len(fkinds))
						fkind := fkinds[fishIndex]
						p.fishIndex = (p.fishIndex + 1) % 100000
						path := p.doRandomPath(fkind.Paths, excludes)
						if path == -1 {
							continue
						}
						excludes[path] = path
						push := &plr.P3080001_FishInfo{
							FishID: p.fishIndex,
							KindID: fkind.ID,
							Path:   int32(path),
						}
						frec := &model.FishRecord{}
						frec.FishID = p.fishIndex
						frec.KindID = fkind.ID
						frec.KindType = fkind.KindType
						frec.Score = fkind.Score
						frec.KindDesc = fkind.KindDesc
						frec.Path = path
						frec.Interval = fkind.Interval * 1000
						frec.StartTime = p.getCurMillisecond()
						p.fishRecord[p.fishIndex] = frec
						arr = append(arr, push)
					}
					push.Fishes = arr
					push.RoomID = p.rid
					push.TableID = p.tid
					if len(arr) > 0 {
						var interval int64 = 60 * 1000
						if ti, ok := conf.Conf.TypeInterval[t]; ok {
							interval = ti * 1000
						}
						p.typeclock[t] = p.getCurMillisecond() + interval
					}
				}
			}
			if len(push.Fishes) > 0 {
				//redisx.SInfo("push P3080001", zap.Any("fishes", push.Fishes))
				p.enity.Push("P3080001", push)
			}
		}
	}
}

func (p *RSession) doFishTideReady() {
	//避免推多次
	if !p.bPushedTide {
		glog.SInfof("鱼潮准备开始")
		p.tideID = (p.tideID + 1) % len(conf.Conf.FishTide)
		push := &plr.P3080016{
			TableID:  p.tid,
			TideID:   int32(p.tideID),
			WaveTime: consts.TimeForWave,
		}
		p.enity.Push("P3080016", push)
		p.bPushedTide = true
	}
}

func (p *RSession) doFishTideStart() {
	glog.SInfof("鱼潮开始")
	p.fishRecord = make(map[int32]*model.FishRecord)

	now := p.getCurMillisecond()
	var maxtime int32 = -1
	confTide := conf.Conf.FishTide[p.tideID]
	mapTide := make(map[int32]conf.TideData, 0)
	for _, v := range confTide.TideData {
		mapTide[v.TimeAxis] = v
		if maxtime < v.TimeAxis {
			maxtime = v.TimeAxis
		}
	}

	push := &plr.P3080009{}
	arr := make([]*plr.P3080009_FishInfo, 0)
	for _, tv := range confTide.TideData {
		kids := strings.Split(tv.KindIDs, ",")
		paths := strings.Split(tv.Paths, ",")
		if len(kids) == len(paths) {
			kmap := p.fishKindToMap()
			for i, v := range kids {
				p.fishIndex = (p.fishIndex + 1) % 100000
				kid, err1 := strconv.ParseInt(v, 10, 0)
				pid, err2 := strconv.ParseInt(paths[i], 10, 0)
				if err1 == nil && err2 == nil {
					if kind, ok := kmap[int32(kid)]; ok {
						push01 := &plr.P3080009_FishInfo{
							FishID:    p.fishIndex,
							KindID:    kind.ID,
							Path:      int32(pid),
							SpawnTime: int64(tv.TimeAxis),
						}
						frec := &model.FishRecord{}
						frec.FishID = p.fishIndex
						frec.KindID = kind.ID
						frec.KindType = kind.KindType
						frec.Score = kind.Score
						frec.KindDesc = kind.KindDesc
						frec.Path = int32(pid)
						frec.Interval = kind.Interval
						frec.SpawnTime = push01.SpawnTime

						frec.StartTime = now + int64(tv.TimeAxis)
						p.fishRecord[p.fishIndex] = frec
						arr = append(arr, push01)
					}
				}
			}
		}
	}
	push.Fishes = arr
	push.TableID = p.tid
	p.enity.Push("P3080009", push)

	p.fishTideStopTime = p.getCurMillisecond() + consts.TimeForWave + int64(maxtime) + int64(confTide.Delay)
	p.fishTideStartTime = p.fishTideStopTime + consts.TimeForTideInterval
	p.bossStartTime = p.fishTideStopTime + consts.TimeForBossInterval
	p.status = consts.GameStatusStartingFishTide
}

//返回去除指定元素后的数组
// func (p *RSession) excludeData(arr []int32, exclude int32) []int32 {
// 	ret := make([]int32, 0)
// 	for _, v := range arr {
// 		if v != exclude {
// 			ret = append(ret, v)
// 		}
// 	}
// 	return ret
// }

func (p *RSession) doBossFishReady() {
	if !p.bPushedBoss {
		glog.SInfof("准备出boss鱼")
		push02 := &plr.P3080011{
			TableID: p.tid,
		}
		p.enity.Push("P3080011", push02)
		p.bPushedBoss = true
	}
}

func (p *RSession) doBossFish() {
	if _, isok := p.skillRecord["0"]; !isok {
		if p.bossStartTime+consts.TimeForBoss < p.getCurMillisecond() && p.status == consts.GameStatusReadyBoss {
			p.bossStartTime = p.getCurMillisecond() + consts.TimeForTideInterval
			boss := p.bossFish[p.bossIndex]
			//boss的路径是固定的，直接用fishid作为路径
			path := boss.ID
			p.fishIndex = (p.fishIndex + 1) % 100000
			p.status = consts.GameStatusReadyBoss
			p.bossIndex = (p.bossIndex + 1) % int32(len(p.bossFish))
			p.status = consts.GameStatusFreeFish
			p.pushBossFish(p.fishIndex, boss, path)
		}
	}
}

func (p *RSession) pushBossFish(fishIndex int32, fkind *model.FishKind, path int32) {
	frec := &model.FishRecord{}
	frec.FishID = fishIndex
	frec.KindID = fkind.ID
	frec.KindType = fkind.KindType
	frec.Score = fkind.Score
	frec.KindDesc = fkind.KindDesc
	frec.Path = path
	frec.Interval = fkind.Interval
	frec.StartTime = p.getCurMillisecond()
	frec.BBoss = true
	p.fishRecord[fishIndex] = frec
	push := &plr.P3080001{}
	arr := make([]*plr.P3080001_FishInfo, 0)
	push01 := &plr.P3080001_FishInfo{
		FishID: fishIndex,
		KindID: fkind.ID,
		Path:   path,
	}
	arr = append(arr, push01)
	push.Fishes = arr
	push.RoomID = p.rid
	push.TableID = p.tid
	p.enity.Push("P3080001", push)
}

func (p *RSession) fishKindToMap() map[int32]*model.FishKind {
	k := make(map[int32]*model.FishKind, 0)
	for _, v := range p.fishKind {
		for _, v1 := range v {
			k[v1.ID] = v1
		}
	}
	return k
}

func (p *RSession) doRandomPath(paths string, exludes map[int32]int32) int32 {
	pathIds := make([]int32, 0)
	if paths == "-1" {
		pathIds = p.fishPathIndex
	} else {
		pathArr := strings.Split(paths, ",")
		for _, v := range pathArr {
			pid, err := strconv.ParseInt(v, 10, 0)
			if err == nil {
				pathIds = append(pathIds, int32(pid))
			}
		}
	}

	//除去指定的路径
	if len(exludes) > 0 {
		mappaths := make(map[int32]int32)
		for _, m := range pathIds {
			mappaths[m] = m
		}
		for _, v := range exludes {
			delete(mappaths, v)
		}
		ids := make([]int32, 0)
		for v := range mappaths {
			ids = append(ids, v)
		}
		pathIds = ids
	}

	if len(pathIds) == 1 {
		return pathIds[0]
	}

	zoneLen := len(pathIds)
	if zoneLen == 0 {
		return -1
	}
	return pathIds[rand.Intn(zoneLen)]
}

func (p *RSession) doShoot(bulletid int32, userid string, tableid, sid int32, vectorx, vectory, ratio float64, speed, fishid int32) {
	now := time.Now().Unix()
	if sid < 0 || sid >= consts.FishSeatNumbers {
		glog.SErrorf("shoot failed.invalid sid:%v", sid)
		return
	}

	if ratio <= 0 || ratio < p.minRatio || ratio > p.maxRatio {
		glog.SErrorf("shoot failed.invalid ratio:%v", ratio)
		return
	}

	ps := p.ps[sid]
	if ps == nil {
		glog.SErrorf("shoot failed.user not exists:%v", sid)
		return
	}

	if ps.amount < ratio {
		glog.SErrorf("user wealth no enough.uid:%v,wealth:%v,ratio:%v", ps.uid, ps.amount, ratio)
		return
	}
	if userid == p.ps[sid].uid {
		glog.SInfof("[doShoot] userid:%v bulletid:%v costAmount:%v", userid, bulletid, ratio)
		p.setPlBullet(ratio, sid)
	} else {
		glog.SErrorf("shoot failed. session uid:%v,msg uid:%v", p.ps[sid].uid, userid)
		return
	}
	p.ps[sid].isAmountChanged = true

	bullet := &model.BulletInfo{
		BulletID:  bulletid,
		TableID:   tableid,
		SeatNo:    sid,
		VectorX:   vectorx,
		VectorY:   vectory,
		Speed:     speed,
		UserID:    userid,
		StartTime: now,
		Wealth:    ratio,
	}
	key := fmt.Sprintf("%d_%d_%d", tableid, sid, bulletid)
	p.bullet[key] = bullet

	p.ps[sid].AddRecordBullet(fmt.Sprintf("%0.3f", bullet.Wealth), bullet.Wealth, consts.RecordBullet)

	push := &plr.P3080004{
		BulletID:   bulletid,
		TableID:    tableid,
		SeatNo:     sid,
		VectorX:    vectorx,
		VectorY:    vectory,
		Speed:      speed,
		UserID:     userid,
		UserAmount: ps.amount,
		FishID:     fishid,
	}
	p.enity.Push("P3080004", push)
}

func (p *RSession) doShootFish(bulletid int32, userid string, fishID, tableid, sid int32) {
	key := fmt.Sprintf("%d_%d_%d", tableid, sid, bulletid)
	pbullet, ok := p.bullet[key]
	if !ok {
		glog.SErrorf("invalid bullet id %v", key)
		return
	}
	bullet := *pbullet
	if bullet.Wealth <= 0 {
		glog.SErrorf("invalid bullet wealth %v", bullet.Wealth)
		return
	}

	pl := p.ps[sid]
	p.checkNewPlayer(pl)
	uextrarate := p.getPlayerExtraRate(pl)
	pfish, ok := p.fishRecord[fishID]
	if !ok {
		glog.SWarnf("invalid fishid %v,uid:%v,bulletid:%v,wealth:%v", fishID, userid, bullet.BulletID, bullet.Wealth)
	} else {
		fish := *pfish
		//redisx.SInfo("子弹打中鱼",
		//	zap.Any("fish", fish),
		//	zap.Any("bullet", bullet),
		//	zap.Any("uextrarate", uextrarate),
		//	zap.Any("isnewPlayer", pl.isnewPlayer))
		if p.isCapture(bullet, fish, uextrarate, pl.isnewPlayer) {
			//redisx.SInfo("被捕获")
			if fish.KindID == 28 {
				push := &plr.P3080010{}
				arr := make([]*plr.P3080010_CaptureFish, 0)
				//电鳗鱼
				delIds := make([]int32, 0)
				now := p.getCurMillisecond()
				smallcounter := 0
				addWealth := decimal.Zero
				rwarr := make([]*SettleReward, 0)

				//5分以下的鱼必然死，5-7分随机
				for _, v := range p.fishRecord {
					pathdata := p.fishPath[v.Path]
					if v.Score < 5 {
						if pathdata.Deadtime-(now-v.StartTime) > 1000 {
							if smallcounter < 10 {
								delIds = append(delIds, v.FishID)
								fishAmount := decimal.NewFromFloat(v.Score).Mul(decimal.NewFromFloat(bullet.Wealth))
								addWealth = addWealth.Add(fishAmount)
								p := &plr.P3080010_CaptureFish{}
								p.FishID = v.FishID
								p.Score, _ = fishAmount.Float64()
								arr = append(arr, p)
								rw := &SettleReward{
									bulletID:         bulletid,
									bulletCostWealth: bullet.Wealth,
									kindID:           v.KindID,
									kindScore:        v.Score,
								}
								rwarr = append(rwarr, rw)
								smallcounter++
							}
						}
					} else if v.Score <= 7 {
						if pathdata.Deadtime-(now-v.StartTime) > 1000 {
							rnd := rand.Intn(100)
							if rnd < 50 {
								delIds = append(delIds, v.FishID)
								fishAmount := decimal.NewFromFloat(v.Score).Mul(decimal.NewFromFloat(bullet.Wealth))
								addWealth = addWealth.Add(fishAmount)
								p := &plr.P3080010_CaptureFish{}
								p.FishID = v.FishID
								p.Score, _ = fishAmount.Float64()
								arr = append(arr, p)
								rw := &SettleReward{
									bulletID:         bulletid,
									bulletCostWealth: bullet.Wealth,
									kindID:           v.KindID,
									kindScore:        v.Score,
								}
								rwarr = append(rwarr, rw)
							}
						}
					}
				}
				addWealth = addWealth.Truncate(3)
				glog.SInfof("电鳗鱼金币:%v", addWealth)
				for _, fid := range delIds {
					delete(p.fishRecord, fid)
				}
				delete(p.fishRecord, fishID)

				// roundrecord := p.ps[sid].roundRecord
				for _, rw := range rwarr {
					p.setPlRewards(rw.bulletCostWealth, rw.kindScore, sid)
					p.ps[sid].AddRecordReward(bullet.Wealth, rw.kindID, p.minRatio)
				}
				if len(arr) > 0 {
					push.BulletID = bulletid
					push.FishID = fishID
					push.TableID = tableid
					push.SeatNo = sid
					push.Ratio = bullet.Wealth
					push.AddWealth, _ = addWealth.Float64()
					push.Fishes = arr
					push.UserAmount = p.ps[sid].amount
					p.enity.Push("P3080010", push)

				}
			} else {
				addWealth, _ := decimal.NewFromFloat(bullet.Wealth).Mul(decimal.NewFromFloat(fish.Score)).Truncate(3).Float64()
				if userid == p.ps[sid].uid {
					rw := &SettleReward{
						bulletID:         bulletid,
						bulletCostWealth: bullet.Wealth,
						kindID:           fish.KindID,
						kindScore:        fish.Score,
					}
					rwAmount := p.setPlRewards(rw.bulletCostWealth, rw.kindScore, sid)
					p.ps[sid].AddRecordReward(bullet.Wealth, rw.kindID, p.minRatio)
					delete(p.fishRecord, fishID)
					push2 := &plr.P3080006{
						BulletID:   bulletid,
						FishID:     fishID,
						TableID:    tableid,
						SeatNo:     sid,
						Ratio:      bullet.Wealth,
						AddWealth:  addWealth,
						UserAmount: p.ps[sid].amount,
					}
					p.enity.Push("P3080006", push2)
					if fish.KindID == 26 || fish.KindID == 27 {
						if rwAmount >= consts.SpecilLimitAmount {
							var kindname string
							for _, fv := range p.bossFish {
								if fv.ID == fish.KindID {
									kindname = fv.KindName
									break
								}
							}
							if kindname != "" {
								//sPush := &plr.P0000001{}
								//sPush.Amount = rwAmount
								//sPush.KindID = consts.FishKindID
								//sPush.UserName = p.ps[sid].nick
								//sPush.CardType = kindname
								//sPush.GameIDs = p.gameids
								//sPush.VipLevel = p.ps[sid].vipLevel
								//p.enity.Push("P0000001", sPush)
							}
						}
					}
				} else {
					glog.SErrorf("invalid userid session uid:%v,msg uid:%v", p.ps[sid].uid, userid)
				}
			}
		} else {
			//glog.SInfof("子弹miss")
		}
	}
	delete(p.bullet, key)
}

func (p *RSession) setPlRewards(bulletRatio, fishscore float64, sid int32) float64 {
	rewardValue := decimal.NewFromFloat(bulletRatio).Mul(decimal.NewFromFloat(fishscore)).Truncate(3)
	p.ps[sid].amount, _ = rewardValue.Add(decimal.NewFromFloat(p.ps[sid].amount)).Truncate(3).Float64()
	p.ps[sid].payoff = p.ps[sid].payoff.Add(rewardValue).Truncate(3)
	if p.ps[sid].isnewPlayer {
		p.control.UpdateNewAmount(p.rid, rewardValue.Neg(), decimal.Zero)
	} else {
		p.control.UpdateAmount(p.rid, rewardValue.Neg(), decimal.Zero)
	}

	p.ps[sid].inven = p.ps[sid].inven.Sub(rewardValue)

	rwamount, _ := rewardValue.Truncate(3).Float64()
	return rwamount
}

func (p *RSession) setPlSkill(skillType int32, costwealth float64, sid int32) {
	// p.ps[sid].amount, _ = decimal.NewFromFloat(p.ps[sid].amount).Sub(decimal.NewFromFloat(costwealth)).Truncate(3).Float64()
	// skillAmount := decimal.NewFromFloat(costwealth)
	// p.ps[sid].skillAmount[skillType] = p.ps[sid].skillAmount[skillType].Add(skillAmount).Truncate(3)
	p.ps[sid].skillCount[skillType] = p.ps[sid].skillCount[skillType] + 1
}

func (p *RSession) setPlBullet(costAmount float64, sid int32) {
	bulletAmount := decimal.NewFromFloat(costAmount)
	p.ps[sid].amount, _ = decimal.NewFromFloat(p.ps[sid].amount).Sub(bulletAmount).Truncate(3).Float64()
	poolAmount := bulletAmount
	reven := decimal.Zero
	poolreven := decimal.Zero
	if p.ps[sid].isPayPlayer {
		poolAmount = bulletAmount.Mul(p.gameCommission[p.ps[sid].gameid])
		reven = bulletAmount.Sub(poolAmount.Truncate(6))
		poolreven = bulletAmount.Mul(conf.Conf.DecRevenueRate).Truncate(6)
		poolAmount = poolAmount.Sub(poolreven)
	}
	p.ps[sid].actualamount = p.ps[sid].actualamount.Add(bulletAmount).Truncate(3)
	p.ps[sid].payoff = p.ps[sid].payoff.Sub(bulletAmount).Truncate(3)
	p.ps[sid].revenue = p.ps[sid].revenue.Add(reven).Truncate(6)
	if p.ps[sid].isnewPlayer {
		p.control.UpdateNewAmount(p.rid, poolAmount, poolreven)
	} else {
		p.control.UpdateAmount(p.rid, poolAmount, poolreven)
	}
	p.ps[sid].inven = p.ps[sid].inven.Add(poolAmount)
}

func (p *RSession) doSettle() {
	for _, v := range p.ps {
		if v != nil {
			ppl := *v
			p.playerSettle(ppl)
		}
	}
}

func (p *RSession) playerSettle(v Player) {
	if len(v.uid) > 0 {
		glog.SInfof("[playerSettle] uid:%v payoff:%v", v.uid, v.payoff.String())

		//技能结算
		skillSettle := false
		bulletSettle := true
		for _, v := range v.skillCount {
			if v > 0 {
				skillSettle = true
			}
		}

		if v.payoff.Equal(decimal.Zero) && v.revenue.Equal(decimal.Zero) &&
			v.actualamount.Equal(decimal.Zero) {
			bulletSettle = false
		}
		if skillSettle == false && bulletSettle == false {
			return
		}

		payofValue := v.payoff
		tx, err := p.db.Begin()
		if err != nil {
			glog.SErrorf("doSettle failed.Db error:%v", err)
			return
		}

		if bulletSettle {
			// _, err = tx.Exec(`insert into settlement_fish_t (roundcode,gametableid,userid,betamount
			// 	,bettype,actualamount,odds,payoffvalue,results,bettime,settletime,revenue) values(?,?,?,?,?,?,?,?,?,?,?,?)`,
			// 	v.rName, p.tid, v.uid, v.actualamount, 1, v.actualamount, 1, payofValue,
			// 	"", time.Now().Unix(), time.Now().Unix(), v.revenue)
			// if err != nil {
			// 	glog.SErrorf("insert into settlement_fish_t  failed.Db error:%v", err)
			// 	tx.Rollback()
			// 	return
			// }

			p.ps[v.sid].totalActualamount = p.ps[v.sid].totalActualamount.Add(v.actualamount)
			p.ps[v.sid].totalPayoff = p.ps[v.sid].totalPayoff.Add(payofValue)
			p.ps[v.sid].totalRevenue = p.ps[v.sid].totalRevenue.Add(v.revenue)

			// settlewid, err := model.InsertRecordAmount(v.uid, consts.FishKindID, v.rName, tx)
			// if err != nil {
			// 	glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
			// 	tx.Rollback()
			// 	return
			// }

			/*
				_, err = tx.Exec(`update userwealth_t set wealth = wealth + ?,profit = profit + ? where userid=?`, payofValue, payofValue, v.uid)
				if err != nil {
					glog.SErrorf("update  wealth failed.Db error:%v", err)
					tx.Rollback()
					return
				}
			*/

			changeBalance, _ := payofValue.Float64()
			changeBalanceType := consts2.CHANGE_BALANCE_WIN
			if changeBalance < 0 {
				changeBalanceType = consts2.CHANGE_BALANCE_BET
			}
			_, err = v.ChangeBalance(changeBalanceType, int(changeBalance), util.GetTraceId(), false)
			if err != nil {
				glog.SErrorf("[playerSettle] ChangeBalance failed! uid:%v err:%v", v.uid, err)
				tx.Rollback()
				return
			}

			// err = model.UpdateRecordAmount(v.settlewid, v.uid, tx)
			// if err != nil {
			// 	glog.SErrorf("update recordAmount failed for db.%v", err.Error())
			// 	tx.Rollback()
			// 	return
			// }
			// tx.Commit()
			// p.savePoolAmount()

		}

		if skillSettle {
			err := p.UpdateUseSkill(tx, &v)
			if err != nil {
				glog.SErrorf("[playerSettle] ChangeBalance failed! uid:%v err:%v", v.uid, err)
				tx.Rollback()
				return
			}
		}

		err = model.UpdateRecordAmount(v.settlewid, v.uid, tx)
		if err != nil {
			glog.SErrorf("update recordAmount failed for db.%v", err.Error())
			tx.Rollback()
			return
		}

		if rev := v.revenue; rev.GreaterThan(decimal.Zero) {
			//pf, _ := rev.Float64()
			//err = model.AddSVipValue(v.uid, pf, tx)
			//if err != nil {
			//	glog.SErrorf("AddSVipValue failed for db.%v", err)
			//	tx.Rollback()
			//}

			// 	vlevel, newVlevel, err := model.ComputeVipLevel(v.uid, tx)
			// 	if err != nil {
			// 		tx.Rollback()
			// 		glog.SErrorf("ComputeVipLevel failed.error:%v", err)
			// 	}

			// 	if vlevel != newVlevel {
			// 		vPush := &plr.P0000003{}
			// 		vPush.UserID = v.uid
			// 		vPush.VipLevel = newVlevel
			// 		p.enity.Push("P0000003", vPush)
			// 	}
		}

		tx.Commit()
		p.savePoolAmount()
		p.ps[v.sid].CleaSettleData()
	}
}

func (p *RSession) UpdateUseSkill(tx *sql.Tx, player *Player) error {
	var (
		uid         = player.uid
		skillCount  = player.skillCount
		skillAmount = player.skillAmount
	)
	//
	if len(skillCount) != len(skillAmount) {
		glog.SErrorf("invalid skillcount and skillamount len,countlen:%v,amount:%v", len(skillCount), len(skillAmount))
		return fmt.Errorf("invalid skillcount and skillamount len")
	}
	setsql := ""
	totalAmount := decimal.Zero
	for i, v := range skillCount {
		if v > 0 {
			if setsql == "" {
				totalAmount = totalAmount.Sub(skillAmount[i])
				setsql = fmt.Sprintf(" skill_%v_count=skill_%v_count+%v,skill_%v_amount=skill_%v_amount+%v", i, i, v, i, i, skillAmount[i])
			} else {
				totalAmount = totalAmount.Sub(skillAmount[i])
				setsql = fmt.Sprintf(" %s,skill_%v_count=skill_%v_count+%v,skill_%v_amount=skill_%v_amount+%v", setsql, i, i, v, i, i, skillAmount[i])
			}
		}

	}

	if setsql != "" {
		sql := fmt.Sprintf(`update fish_user_extend_t set %s where userid=?`, setsql)
		_, err := tx.Exec(sql, uid)
		if err != nil {
			glog.SErrorf("update fish_user_extend_t failed err:%v", err)
			return err
		}
	}
	//_, err := tx.Exec(`update userwealth_t set wealth = wealth + ?,profit = profit + ? where userid=?`, totalAmount, totalAmount, uid)
	//if err != nil {
	//	glog.SErrorf("update  wealth failed.Db error:%v", err)
	//	return err
	//}

	changeBalance, _ := totalAmount.Float64()
	_, err := player.ChangeBalance(consts2.CHANGE_BALANCE_BET, int(changeBalance), util.GetTraceId(), false)
	if err != nil {
		glog.SErrorf("[UpdateUseSkill] ChangeBalance failed! uid:%v err:%v", player.uid, err)
		return err
	}

	return nil
}

func (p *RSession) savePoolAmount() {
	amount := p.control.GetToTal(p.rid)
	p.inven.PoolAmount = amount
	p.inven.UpdateTime = time.Now().Unix()
	p.inven.Revenue = p.control.GetRevenue(p.rid)
	model.UpdateInventory(p.inven, p.db)
	//push := &explr.P1010002{}
	//push.KindID = consts.FishKindID
	//push.PoolAmount = amount
	//push.RoomID = p.rid
	//push.TableID = p.tid
	//push.UID = p.control.UID()
	//p.enity.Push("P1010002", push)
}

func (p *RSession) saveNewPoolAmount() {
	amount := p.control.GetNewToTal(p.rid)
	_, err := p.db.Exec(`update fish_room_config set poolamount=? where gameroomid=?`, amount, p.rid)
	if err != nil {
		glog.SErrorf("update fish_room_config poolamount failed.err:%v", err)
		return
	}

	// p.inven.Revenue = 0
	// model.UpdateInventory(p.inven, p.db)
	// push := &plr.P1010002{}
	// push.KindID = consts.FishKindID
	// push.PoolAmount = amount
	// push.RoomID = p.rid
	// push.TableID = p.tid
	// push.UID = p.control.UID()
	// p.enity.Push("P1010002", push)
}

func (p *RSession) doChangeCannonRatio(uid string, sid int32, ratio float64) {
	if ratio <= 0 {
		glog.SErrorf("invalid ratio:%v", ratio)
		return
	}
	//改成每次进入房间默认最小倍率,因此不需要写入数据库
	// err := model.UpdateUserCannonRate(p.db, ratio, uid)
	// if err != nil {
	// 	glog.SErrorf("update curr_cannon_ratio failed.uid:%v, err:%v", uid, err)
	// 	return
	// }

	if sid < int32(len(p.ps)) {
		if p.ps[sid].uid == uid {
			p.ps[sid].cannonRatio = ratio
		}
	}

	push := &plr.P3080008{
		TableID: p.tid,
		SeatNo:  sid,
		Ratio:   ratio,
	}
	p.enity.Push("P3080008", push)
}

func (p *RSession) doPurchaseCannon(uid string, sid, cid int32, mid uint) {
	now := time.Now().Unix()
	rsp := &plr.S3080006{}
	var cannonID int32

	if sid < 0 || sid >= int32(len(p.ps)) {
		glog.SErrorf("invalid sid:%v", sid)
		rsp.RetCode = consts.ErrorInvalidParams
		p.enity.Response(rsp, mid)
		return
	}

	if p.ps[sid].uid != uid {
		glog.SErrorf("invalid uid session uid :%v,msg uid:%v", p.ps[sid].uid, uid)
		rsp.RetCode = consts.ErrorInvalidParams
		p.enity.Response(rsp, mid)
		return
	}

	err := p.db.QueryRow(`select cannonid from fish_user_cannon_t where userid=? and   endtime>? and cannonid=?
		union 
		select id as cannonid from  fish_cannon_t where isneedbuy=0 and id=?`, uid, now, cid, cid).Scan(&cannonID)
	if err != sql.ErrNoRows {
		glog.SErrorf("user has own the cannon cid:%v ,err:%v", cid, err)
		rsp.RetCode = consts.ErrorInvalidParams
		p.enity.Response(rsp, mid)
		return
	}

	var periodDay int32
	var costWealth float64
	var cannonName string
	err = p.db.QueryRow(`select period_day,cost_wealth,name from  fish_cannon_t where id=?`, cid).Scan(&periodDay, &costWealth, &cannonName)
	if err != nil {
		glog.SErrorf("query cannon failed cid:%v ,err:%v", cid, err)
		rsp.RetCode = consts.ErrorDB
		p.enity.Response(rsp, mid)
		return
	}

	coinWealth := p.ps[sid].amount
	if coinWealth < costWealth {
		glog.SErrorf("user wealth not enough  uid:%v , wealth:%v costwealth:%v", uid, coinWealth, costWealth)
		rsp.RetCode = consts.ErrorUserWealth
		p.enity.Response(rsp, mid)
		return
	}

	tx, err := p.db.Begin()
	if err != nil {
		glog.SErrorf("user has own the cannon cid:%v ,err:%v", cid, err)
		rsp.RetCode = consts.ErrorDB
		p.enity.Response(rsp, mid)
		return
	}

	validTime := time.Duration(periodDay) * 24 * time.Hour
	_, err = tx.Exec(`insert into fish_user_cannon_t(userid,cannonid,starttime,endtime ) 
	values(?,?,?,?)`, uid, cid, time.Now().Unix(), time.Now().Add(validTime).Unix())
	if err != nil {
		tx.Rollback()
		glog.SErrorf("insert fish_user_cannon_t failed cid:%v ,err:%v", cid, err)
		rsp.RetCode = consts.ErrorDB
		p.enity.Response(rsp, mid)
		return
	}

	//只有三个需要购买的炮台,cid为1-3
	if costWealth > 0 && cid > 0 && cid <= 3 {
		sql := fmt.Sprintf(`update fish_user_extend_t set  cannon_%v_count=cannon_%v_count+1,
			 cannon_%v_amount=cannon_%v_amount+?  where userid=?`, cid, cid, cid, cid)
		_, err = tx.Exec(sql, costWealth, uid)
		if err != nil {
			tx.Rollback()
			glog.SErrorf("update  fish_user_extend_t failed uid:%v ,err:%v", uid, err)
			rsp.RetCode = consts.ErrorDB
			p.enity.Response(rsp, mid)
			return
		}
	}

	//_, err = tx.Exec(`update userwealth_t set wealth = wealth - ?,profit = profit - ? where userid=?`, costWealth, costWealth, uid)
	//if err != nil {
	//	tx.Rollback()
	//	glog.SErrorf("update  wealth failed.Db error:%v", err)
	//	rsp.RetCode = consts.ErrorDB
	//	p.enity.Response(rsp, mid)
	//	return
	//}

	_, err = p.ps[sid].ChangeBalance(consts2.CHANGE_BALANCE_BET, int(costWealth), util.GetTraceId(), false)
	if err != nil {
		tx.Rollback()
		glog.SErrorf("[doPurchaseCannon] ChangeBalance failed! error:%v", err)
		rsp.RetCode = consts.ErrorDB
		p.enity.Response(rsp, mid)
		return
	}

	tx.Commit()

	p.ps[sid].amount, _ = decimal.NewFromFloat(p.ps[sid].amount).Sub(decimal.NewFromFloat(costWealth)).Float64()
	p.ps[sid].AddRecordBullet(cannonName, costWealth, consts.RecordPurchase)
	rsp.CostWealth = costWealth
	rsp.CannonID = cid
	rsp.Lefttime = int64(validTime / time.Second)
	p.enity.Response(rsp, mid)
	p.ps[sid].isAmountChanged = true
	push := &plr.P3080013{}
	push.TableID = p.tid
	push.SeatNo = sid
	push.UserWealth = p.ps[sid].amount
	p.enity.Push("P3080013", push)
}

func (p *RSession) doChangeCannonID(uid string, sid, cid int32, mid uint) {
	rsp := &plr.S3080007{}
	var cannonID int32
	currTime := time.Now().Unix()
	if sid < int32(len(p.ps)) {
		if p.ps[sid].uid == uid {
			p.ps[sid].cannonID = cid
		} else {
			glog.SErrorf("invalid sid:%v", sid)
			rsp.RetCode = consts.ErrorInvalidParams
			p.enity.Response(rsp, mid)
			return
		}
	} else {
		glog.SErrorf("invalid sid:%v", sid)
		rsp.RetCode = consts.ErrorInvalidParams
		p.enity.Response(rsp, mid)
		return
	}

	err := p.db.QueryRow(`select cannonid from fish_user_cannon_t where userid=? and   endtime>? and cannonid=?
		union 
		select id as cannonid from  fish_cannon_t where isneedbuy=0 and id=?`, uid, currTime, cid, cid).Scan(&cannonID)
	if err != nil {
		glog.SErrorf("usre has not buy the cannon uid:%v, cid:%v ,err:%v", uid, cid, err)
		rsp.RetCode = consts.ErrorInvalidParams
		p.enity.Response(rsp, mid)
		return
	}

	_, err = p.db.Exec(`update fish_user_extend_t set curr_cannon_id=? where userid=?`, cid, uid)
	if err != nil {
		glog.SErrorf("update user curr cannon failed ,uid:%v,err:%v", uid, err)
		rsp.RetCode = consts.ErrorDB
		p.enity.Response(rsp, mid)
		return
	}
	rsp.CannonID = cid
	p.enity.Response(rsp, mid)

	push := &plr.P3080014{}
	push.TableID = p.tid
	push.SeatNo = sid
	push.CannonID = cid
	p.enity.Push("P3080014", push)
}

func (p *RSession) doSkill(userid string, tableid, seatno, skillType int32, opType int32, mid uint) {
	rsp := &plr.S3080005{}
	rsp.SkillType = skillType
	tid := tableid
	sid := seatno
	uid := userid

	if skillType < 1 || skillType > 2 {
		rsp.RetCode = consts.ErrorInvalidParams
		glog.SErrorf("invalid skillType:%v", skillType)
		p.enity.Response(rsp, mid)
		return
	}

	recordkey := fmt.Sprintf("%d_%d_%d", p.rid, skillType, sid)
	skillkey := fmt.Sprintf("%d_%d", p.rid, skillType)
	if skillType == 0 {
		recordkey = "0"
		_, isok := p.skillRecord[recordkey]
		if isok {
			rsp.RetCode = consts.ErrorInvalidParams
			glog.SErrorf("now is skilling key:%v", recordkey)
			p.enity.Response(rsp, mid)
			return
		}

		leftcdtime := consts.SkillCDTime - (p.getCurMillisecond() - p.lastFreezeEndTime)
		if leftcdtime < 0 {
			leftcdtime = 0
		}
		//冰冻技能冷却中
		if leftcdtime > 0 {
			rsp.RetCode = consts.ErrorInvalidParams
			glog.SErrorf("cdtime is greater than zero,uid:%v,leftcdtime:%v", uid, leftcdtime)
			p.enity.Response(rsp, mid)
			return
		}

		if p.status == consts.GameStatusReadyFishTide {
			rsp.RetCode = consts.ErrorInvalidParams
			glog.SErrorf("game status not allow skill,status:%v", p.status)
			p.enity.Response(rsp, mid)
			return
		}
	}

	skill, ok := p.skillType[skillkey]
	if !ok {
		rsp.RetCode = consts.ErrorInvalidParams
		glog.SErrorf("invalid skill rid:%v,skillType:%v", p.rid, skillType)
		p.enity.Response(rsp, mid)
		return
	}

	ps := p.ps[sid]
	if ps == nil {
		rsp.RetCode = consts.ErrorQuerySeat
		glog.SErrorf("invalid seatno tid:%v sid:%v", tid, sid)
		p.enity.Response(rsp, mid)
		return
	}

	// if ps.amount < skill.CostWealth {
	// 	rsp.RetCode = consts.ErrorUserWealth
	// 	glog.SErrorf("user wealth not enough uid:%v, welath:%v,costwealth:%v", uid, ps.amount, skill.CostWealth)
	// 	p.enity.Response(rsp, mid)
	// 	return
	// }

	// ps.isAmountChanged = true
	// rsp.CDTime = cdTime
	// rsp.CostWealth = skill.CostWealth
	rsp.SkillType = skillType
	rsp.OpType = opType
	p.enity.Response(rsp, mid)

	//使用技能
	if opType == 0 || skillType == 0 {
		srec := &model.SkillRecord{}
		srec.SkillType = skill.SkillType
		srec.FreezeTime = skill.FreezeTime
		srec.UserID = userid
		srec.TableID = tableid
		srec.SeatNo = seatno
		srec.StartTime = p.getCurMillisecond()
		p.skillRecord[recordkey] = srec

		if skillType == 0 {
			for _, f := range p.fishRecord {
				f.StartTime = f.StartTime + int64(skill.FreezeTime)
			}
			if p.status >= consts.GameStatusReadyFishTide {
				p.fishTideStartTime = p.fishTideStartTime + int64(skill.FreezeTime)
				p.fishTideStopTime = p.fishTideStopTime + int64(skill.FreezeTime)
				p.bossStartTime = p.bossStartTime + int64(skill.FreezeTime)
			}
		}

		setSkill := &SettleSkill{
			skillID:   skill.ID,
			skillType: skillType,
			// costWealth: skill.CostWealth,
		}
		p.setPlSkill(setSkill.skillType, setSkill.costWealth, sid)
		skillName := skill.SkillName
		p.ps[sid].AddRecordBullet(skillName, 0, consts.RecordSkill)
	}

	//自动不用推P消息
	if skillType != 1 {
		push := &plr.P3080007{
			UserID:    userid,
			TableID:   tableid,
			SeatNo:    seatno,
			SkillType: skillType,
			OpType:    opType,
		}
		p.enity.Push("P3080007", push)
	}

	if skill.FreezeTime > 0 {
		p.clearSkill(skill.FreezeTime, recordkey)
	} else {
		//取消技能
		if opType == 1 {
			delete(p.skillRecord, recordkey)
		} else {
			//锁定和自动不能同时使用
			if skillType == 1 {
				recordkey := fmt.Sprintf("%d_%d_%d", p.rid, 2, sid)
				delete(p.skillRecord, recordkey)
			} else if skillType == 2 {
				recordkey := fmt.Sprintf("%d_%d_%d", p.rid, 1, sid)
				delete(p.skillRecord, recordkey)
			}
		}
	}
}

func (p *RSession) clearSkill(freezetime int32, recordkey string) {
	time.AfterFunc(time.Duration(freezetime)*time.Millisecond, func() {
		p.Invoke(func() {
			delete(p.skillRecord, recordkey)
			if recordkey == "0" {
				push := &plr.P3080005{}
				push.TableID = p.tid
				p.enity.Push("P3080005", push)
				p.lastFreezeEndTime = p.getCurMillisecond()
			}
		})
	})
}

func (p *RSession) getCurMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *RSession) getCurState() {
	switch p.status {
	case consts.GameStatusFreeFish:
		if p.fishTideStartTime < p.getCurMillisecond() {
			if _, isok := p.skillRecord["0"]; !isok {
				glog.SInfof("current game curr status:%v,changed status:鱼潮准备", p.status)
				p.status = consts.GameStatusReadyFishTide
				p.bPushedTide = false
				return
			}
		} else if p.bossStartTime < p.getCurMillisecond() {
			// if _, isok := p.skillRecord["0"]; !isok {
			glog.SInfof("current game status:%v,变更后状态：Boss准备", p.status)
			p.status = consts.GameStatusReadyBoss
			p.bPushedBoss = false
			// }
		}
	case consts.GameStatusReadyFishTide:
		if p.fishTideStartTime+int64(consts.TimeForWave) < p.getCurMillisecond() {
			glog.SInfof("current game status:%v,变更后状态：鱼潮开始", p.status)
			p.status = consts.GameStatusStartFishTide
		}
	case consts.GameStatusStartingFishTide:
		if p.fishTideStopTime < p.getCurMillisecond() {
			glog.SInfof("current game status:%v，变更后状态:普通鱼", p.status)
			p.status = consts.GameStatusFreeFish
		}
	}
}
