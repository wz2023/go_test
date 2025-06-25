package round

import (
	"database/sql"
	"fmt"
	"newstars/Server/fish/consts"
	"newstars/Server/fish/model"
	"newstars/framework/game_center"
	"newstars/framework/glog"
	model2 "newstars/framework/model"
	"newstars/framework/model/data"
	"newstars/framework/util/decimal"
	"time"
)

// Player info
type Player struct {
	uid         string
	gameid      int
	sid         int32
	faceid      int32
	name, nick  string
	amount      float64
	cannonID    int32
	cannonRatio float64
	currency    string
	// bullets     map[int32]float64
	// skills      []*SettleSkill
	// fishRewards []*SettleReward
	actualamount      decimal.Decimal // 下注金额
	skillCount        map[int32]int32
	skillAmount       map[int32]decimal.Decimal
	payoff            decimal.Decimal // 结算值,回报
	revenue           decimal.Decimal // 营收
	rName             string
	sessionID         string
	isAmountChanged   bool  //金币记录是否有变更过
	settlewid         int64 //金币记录id
	settleid          int64 //结算id
	roundRecord       RoundRecord
	RoundPlayers      map[string]RecordPlayer
	Rewards           map[int32]*RecordReward
	RecordBullet      map[string]*RecordBullet
	vipLevel          int32
	faceFrameID       int32
	totalActualamount decimal.Decimal // 总下注金额
	totalPayoff       decimal.Decimal // 总结算值,每1分钟汇总一次
	totalRevenue      decimal.Decimal // 总营收
	isnewPlayer       bool            //是否新玩家
	isPayPlayer       bool            //是否充值玩家
	bankAmount        float64         //弃用
	inven             decimal.Decimal // 个人库存
	onlinetimes       int64
	entertime         int64
	hisrevenue        decimal.Decimal //历史税收，玩家历史营收
}

type SettleReward struct {
	bulletID         int32
	bulletCostWealth float64
	kindID           int32
	kindScore        float64
}

type SettleSkill struct {
	skillID    int32
	skillType  int32
	costWealth float64
}

// NewPlayer create
func NewPlayer(uid string, sid, faceid, faceframeid, cannonID int32, amount, cannonRatio float64, name, nick string, gameid int) *Player {

	skillCount := make(map[int32]int32)
	skillCount[0] = 0
	skillCount[1] = 0
	skillCount[2] = 0

	skillAmount := make(map[int32]decimal.Decimal)
	skillAmount[0] = decimal.Zero
	skillAmount[1] = decimal.Zero
	skillAmount[2] = decimal.Zero

	return &Player{
		uid:         uid,
		gameid:      gameid,
		sid:         sid,
		faceid:      faceid,
		faceFrameID: faceframeid,
		amount:      amount,
		name:        name,
		nick:        nick,
		// bullets:     make(map[int32]float64, 0),
		// skills:      make([]*SettleSkill, 0),
		// fishRewards: make([]*SettleReward, 0),
		cannonRatio:       cannonRatio,
		cannonID:          cannonID,
		revenue:           decimal.Zero,
		actualamount:      decimal.Zero,
		skillCount:        skillCount,
		skillAmount:       skillAmount,
		RoundPlayers:      make(map[string]RecordPlayer),
		Rewards:           make(map[int32]*RecordReward),
		RecordBullet:      make(map[string]*RecordBullet),
		totalActualamount: decimal.Zero,
		totalPayoff:       decimal.Zero,
		totalRevenue:      decimal.Zero,
		hisrevenue:        decimal.Zero,
	}
}

// CleaSettleData clear settle data
func (p *Player) CleaSettleData() {
	// p.bullets = make(map[int32]float64, 0)
	// p.skills = make([]*SettleSkill, 0)
	// p.fishRewards = make([]*SettleReward, 0)
	p.revenue = decimal.Zero
	p.payoff = decimal.Zero
	p.actualamount = decimal.Zero
	skillCount := make(map[int32]int32)
	skillCount[0] = 0
	skillCount[1] = 0
	skillCount[2] = 0
	p.skillCount = skillCount
	skillAmount := make(map[int32]decimal.Decimal)
	skillAmount[0] = decimal.Zero
	skillAmount[1] = decimal.Zero
	skillAmount[2] = decimal.Zero
	p.skillAmount = skillAmount
	// p.RoundPlayers = make(map[int32]RecordPlayer)
	// p.Rewards = make(map[int32]*RecordReward)
	// p.RecordBullet = make(map[string]*RecordBullet)
}

func (p *Player) SubSettleData(v Player) {
	// p.bullets = make(map[int32]float64, 0)
	// p.skills = make([]*SettleSkill, 0)
	// p.fishRewards = make([]*SettleReward, 0)
	p.revenue = p.revenue.Sub(v.revenue)
	p.payoff = p.payoff.Sub(v.payoff)
	p.actualamount = p.actualamount.Sub(v.actualamount)

	skillCount := make(map[int32]int32)
	skillCount[0] = 0
	skillCount[1] = 0
	skillCount[2] = 0
	p.skillCount = skillCount
	skillAmount := make(map[int32]decimal.Decimal)
	skillAmount[0] = decimal.Zero
	skillAmount[1] = decimal.Zero
	skillAmount[2] = decimal.Zero
	p.skillAmount = skillAmount
	// p.RoundPlayers = make(map[int32]RecordPlayer)
	// p.Rewards = make(map[int32]*RecordReward)
	// p.RecordBullet = make(map[string]*RecordBullet)
}

func (p *Player) FillFishInfo(db *sql.DB) {
	kinds, err := model.QueryAllFishKinds(db)
	if err != nil {
		return
	}

	mapkinds := make(map[int32]*model.FishKind)
	for _, v := range kinds {
		mapkinds[v.ID] = v

	}

	for _, v := range p.Rewards {
		if k, isok := mapkinds[v.KindID]; isok {
			v.Score = k.Score
		}
	}
}

func (p *Player) getLevelIndex(ratio, minRatio float64) int64 {
	var level int64 = -1
	if minRatio == consts.SpecialRatio {
		if ratio == 0.001 {
			level = 1
		} else if ratio == 0.005 {
			level = 2
		} else if ratio >= 0.01 {
			level = decimal.NewFromFloat(ratio).Div(decimal.NewFromFloat(0.01)).Round(0).IntPart() + 2
		}
	} else {
		level = decimal.NewFromFloat(ratio).Div(decimal.NewFromFloat(minRatio)).Round(0).IntPart()
	}
	return level
}

func (p *Player) AddRecordReward(ratio float64, kindID int32, minRatio float64) {
	if _, isok := p.Rewards[kindID]; !isok {
		reward := &RecordReward{
			KindID: kindID,
		}
		if minRatio == consts.SpecialRatio {
			reward.Levels = make([]int32, consts.SpecialRatioLevel)
		} else {
			reward.Levels = make([]int32, 10)
		}
		p.Rewards[kindID] = reward
	}

	rw := p.Rewards[kindID]
	// level := decimal.NewFromFloat(ratio).Div(decimal.NewFromFloat(minRatio)).Round(0).IntPart()
	level := p.getLevelIndex(ratio, minRatio)
	if level > 0 && level <= int64(len(rw.Levels)) {
		rw.Levels[level-1]++
	} else {
		glog.SErrorf("invalid ratio:%v,minratio:%v,level:%v,level len:%v", ratio, minRatio, level, len(rw.Levels))
	}

	// switch level {
	// case 1:
	// 	rw.Level1++
	// case 2:
	// 	rw.Level2++
	// case 3:
	// 	rw.Level3++
	// case 4:
	// 	rw.Level4++
	// case 5:
	// 	rw.Level5++
	// case 6:
	// 	rw.Level6++
	// case 7:
	// 	rw.Level7++
	// case 8:
	// 	rw.Level8++
	// case 9:
	// 	rw.Level9++
	// case 10:
	// 	rw.Level10++
	// }
}

func (p *Player) AddRecordBullet(recordName string, baseAmount float64, recordType int) {
	if _, isok := p.RecordBullet[recordName]; !isok {
		bullet := &RecordBullet{
			RecordType: recordType,
			RecordName: recordName,
			BaseAmount: baseAmount,
		}
		p.RecordBullet[recordName] = bullet
	}
	p.RecordBullet[recordName].Nums++
}

func (p *Player) AddRoundPlayer(uid string, sid int32, nick string) {
	if _, isok := p.RoundPlayers[uid]; !isok {
		rp := RecordPlayer{
			Uid:  uid,
			Sid:  sid,
			Nick: nick,
		}

		p.RoundPlayers[uid] = rp
	}
}

func (p *Player) ChangeBalance(t int, balance int, orderID string, isEndRound bool) (int, error) {
	if balance == 0 {
		return 0, nil
	}

	timestamp := int(time.Now().Unix())

	req := &model2.ChangeBalanceReq{
		AppID:            data.BaseConfig.CenterGameAppid,
		Bet:              0,
		Type:             t,
		GameID:           consts.FISH_REAL_GAME_ID,
		Money:            balance,
		Currency:         p.currency,
		OrderID:          orderID,
		PlayerLoginToken: p.uid,
		SessionID:        fmt.Sprintf("rn%10d:%s", timestamp, p.sessionID),
		Timestamp:        timestamp,
		Uname:            p.name,
		EndRound:         isEndRound,
	}

	changeBalance, err := game_center.ChangeBalance(req)
	return changeBalance.Balance, err
}
