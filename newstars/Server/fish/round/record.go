package round

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"newstars/Server/fish/consts"
	"newstars/framework/glog"
	"newstars/framework/util/decimal"
	"time"
)

// 同局玩家
type RecordPlayer struct {
	Sid  int32
	Uid  string
	Nick string
}

// 消耗(包含子弹、技能、购买炮台)
type RecordBullet struct {
	RecordType int     //记录类型 0:子弹 1:技能 2:购买炮台
	RecordName string  //记录名称
	BaseAmount float64 //单价
	Nums       int32   //数量
}

type RecordReward struct {
	KindID int32
	Score  float64 //鱼分值
	Levels []int32 //击杀数量,共有十级
}

// RoundRecord
type RoundRecord struct {
	RoundName    string
	Bettime      int64 //开始时间
	SettleTime   int64 //结算时间
	RoomID       int32
	MinRatio     float64 //最小炮的倍率
	StartAmount  float64
	EndAmount    float64
	BankAmount   float64 //保险箱金币
	PayoffValue  float64 //输赢
	Player       RecordPlayer
	RoundPlayers []RecordPlayer  //同局玩家
	Rewards      []*RecordReward //奖励
	RecordBullet []*RecordBullet
	TitleRatios  []float64 //奖励金币
}

func NewRoundRecord(roundName string, minRatio float64, roomid int32, uid string, sid int32, nick string) RoundRecord {
	player := RecordPlayer{
		Sid:  sid,
		Uid:  uid,
		Nick: nick,
	}
	var ratios []float64
	if minRatio == consts.SpecialRatio {
		ratios = make([]float64, consts.SpecialRatioLevel)
		ratios[0] = 0.001
		ratios[1] = 0.005
		len := len(ratios)
		startRatio, _ := decimal.NewFromFloat(minRatio).Mul(decimal.New(10, 0)).Float64()
		for i := 2; i < len; i++ {
			ratios[i], _ = decimal.NewFromFloat(startRatio).Mul(decimal.New(int64(i-1), 0)).Float64()
		}
	} else {
		ratios = make([]float64, 10)
		for i := range ratios {
			ratios[i], _ = decimal.NewFromFloat(minRatio).Mul(decimal.New(int64(i+1), 0)).Float64()
		}
	}

	record := RoundRecord{
		RoundName:    roundName,
		MinRatio:     minRatio,
		RoomID:       roomid,
		Bettime:      time.Now().Unix(),
		Player:       player,
		RoundPlayers: make([]RecordPlayer, 0),
		Rewards:      make([]*RecordReward, 0),
		RecordBullet: make([]*RecordBullet, 0),
		TitleRatios:  ratios,
	}
	return record

}

func (p *RoundRecord) DumpInfo() {
	glog.SInfof("房间号:%v", p.RoomID)
	glog.SInfof("牌局号:%v", p.RoundName)
	glog.SInfof("开始时间:%v", p.Bettime)
	glog.SInfof("结算时间:%v", p.SettleTime)
	glog.SInfof("入场金币:%v", p.StartAmount)
	glog.SInfof("离场金币:%v", p.EndAmount)
	glog.SInfof("输赢:%0.3f", p.PayoffValue)
	glog.SInfof("保险箱金币:%v", p.BankAmount)
	glog.SInfof("炮%v ID:%v 昵称:%v", p.Player.Sid, p.Player.Uid, p.Player.Nick)
	glog.SInfof("\n局内其他玩家:")
	for _, v := range p.RoundPlayers {
		glog.SInfof("炮%v ID:%v 昵称:%v", v.Sid, v.Uid, v.Nick)
	}

	glog.SInfof("炮弹:\t使用数量:\t合计:\t")
	for _, v := range p.RecordBullet {
		glog.SInfof("%v\t%v\t%v\t", v.RecordName, v.Nums,
			decimal.NewFromFloat(v.BaseAmount).Mul(decimal.New(int64(v.Nums), 0)))
	}

	glog.SInfof("子弹最小倍率:%v", p.MinRatio)
	glog.SInfof("收获:")

	title := ""
	for _, v := range p.TitleRatios {
		title = fmt.Sprintf("%s  %0.3f", title, v)
	}
	glog.SInfof("标题:     %s", title)

	for _, v := range p.Rewards {
		sum := decimal.Zero
		if len(v.Levels) == len(p.TitleRatios) {
			for j, lvl := range v.Levels {
				sum = sum.Add(decimal.New(int64(lvl), 0).Mul(decimal.NewFromFloat(p.TitleRatios[j])))
			}
		}
		total, _ := sum.Mul(decimal.NewFromFloat(v.Score)).Float64()
		if len(v.Levels) == 10 {
			glog.SInfof("%2d:%2.0f      %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %0.3f", v.KindID, v.Score, v.Levels[0], v.Levels[1],
				v.Levels[2], v.Levels[3], v.Levels[4], v.Levels[5], v.Levels[6], v.Levels[7], v.Levels[8], v.Levels[9], total)
		} else if len(v.Levels) == consts.SpecialRatioLevel {
			glog.SInfof("%2d:%2.0f      %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %4d  %0.3f", v.KindID, v.Score, v.Levels[0], v.Levels[1],
				v.Levels[2], v.Levels[3], v.Levels[4], v.Levels[5], v.Levels[6], v.Levels[7], v.Levels[8], v.Levels[9], v.Levels[10], v.Levels[11], total)
		}
	}
}

// Encode 编码
func Encode(his RoundRecord) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(his)
	if err != nil {
		glog.SErrorf("Encode fail %v", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 解码
func Decode(data []byte) (RoundRecord, error) {
	var ret RoundRecord
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ret)
	if err != nil {
		glog.SErrorf("Decode fail %v", err)
		return ret, err
	}
	return ret, nil
}
