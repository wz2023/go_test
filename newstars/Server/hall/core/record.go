package core

import (
	"database/sql"
	"fmt"
	"newstars/framework/glog"
)

/*
select roundcode from settlement_baccarat_t where userid = 1447106
union all select roundcode from settlement_benz_t where userid = 1447106
union all select roundcode from settlement_dt_t where userid = 1447106
union all select roundcode from settlement_ox100_t where userid = 1447106
union all select roundcode from settlement_ox5_t where userid = 1447106
union all select roundcode from settlement_redblack_t where userid = 1447106
union all select roundcode from settlement_t where userid = 1447106
union all select roundcode from settlement_threecard_t where userid = 1447106
limit 50;
*/

// OneDayTime 一天
const (
	OneDayTime     = 86400
	LimitFrequency = 3
)

// C0000038 获取投注记录
//func (w *HallCore) C0000038(s *session.Session, msg *plr.C0000038, mid uint) error {
//	ts, ok := w.rs[msg.GetUserID()]
//	now := time.Now().Unix()
//	if ok {
//		if (now - ts) < LimitFrequency {
//			ret := &plr.S0000038{}
//			ret.RetCode = errcode.FrequencyLimit
//			return s.Response(ret, mid)
//		}
//	}
//	w.rs[msg.GetUserID()] = now
//
//	go func() {
//		kid := msg.GetKindID()
//		uid := msg.GetUserID()
//		pages := msg.GetPageIndex()
//		nums := msg.GetNums()
//		rsp := &plr.S0000038{}
//		rsp.KindID = kid
//		rsp.Days = msg.GetDays()
//		t1 := time.Now()
//		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
//		sTime := t2.AddDate(0, 0, 1-int(rsp.Days)).Unix()
//		sIndex := (pages - 1) * nums
//		eTime := time.Now().Unix() - 15
//
//		switch kid {
//		//case 0:
//		//	err := w.db2.QueryRow(`select sum(sum1),sum(bets),sum(payoffs) from  (select count(1) as sum1,sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_baccarat_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_benz_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_dt_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(betamount) as bets,sum(payoffvalue) as payoffs from settlement_ox100_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(0) as bets,sum(payoffvalue) as payoffs from settlement_ox5_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_redblack_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(0) as bets,sum(payoffvalue) as payoffs from settlement_t where userid = ? and settletime between ? and ?
//		//union all select  count(1) as sum1 ,sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_threecard_t where userid = ? and settletime between ? and ? ) as t`,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//		//	if err != nil {
//		//		rsp.RetCode = errcode.DBError
//		//		glog.SErrorf("Query all failed %v.param:%v", err, msg)
//		//		s.Response(rsp, mid)
//		//		return
//		//	}
//		//	rows, err := w.db2.Query(`select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_baccarat_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_benz_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_dt_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,betamount from settlement_ox100_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_ox5_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_redblack_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_t where userid = ? and settletime between ? and ?
//		//union all select settletime,roundcode,bettype,odds,payoffvalue,actualamount from settlement_threecard_t where userid = ? and settletime between ? and ?
//		//order by settletime desc limit ?,?`,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime,
//		//		uid, sTime, eTime, sIndex, nums,
//		//	)
//		//	if err != nil {
//		//		rsp.RetCode = errcode.DBError
//		//		glog.SErrorf("Query all failed %v.param:%v", err, msg)
//		//		s.Response(rsp, mid)
//		//		return
//		//	}
//		//	defer rows.Close()
//		//
//		//	for rows.Next() {
//		//		item := Settlement{}
//		//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Bettype, &item.Odds, &item.Payoffvalue, &item.Actualamount)
//		//		if err != nil {
//		//			rsp.RetCode = errcode.DBError
//		//			glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//		//			s.Response(rsp, mid)
//		//			return
//		//		}
//		//		rec := &plr.S0000038_RecordRound{}
//		//		rec.SettleTime = item.Settletime
//		//		rec.RoundName = item.Roundcode
//		//		ss := strings.Split(item.Roundcode, "_")
//		//		if len(ss) != 3 {
//		//			rsp.RetCode = errcode.DBError
//		//			glog.SErrorf("Query all failed.roundcode:%v.param:%v", item.Roundcode, msg)
//		//			s.Response(rsp, mid)
//		//			return
//		//		}
//		//		rid, _ := strconv.Atoi(ss[1])
//		//		rec.RoomID = int32(rid)
//		//		switch queryKindByRID(w.db2, rec.RoomID) {
//		//		case 1:
//		//			rec.RoundResult = landlordResult(item.Bettype, item.Payoffvalue)
//		//			rec.Play = landlordPlay(item.Bettype)
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = landlordAmount(item.Odds)
//		//		case 2:
//		//			rec.Play, rec.RoundResult = threePlayAndResult(w.db2, item.Roundcode, uid)
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//		case 3:
//		//			var bAmount float64
//		//			rec.RoundResult, bAmount = ox100Result(w.db2, item.Roundcode)
//		//			playDef := []string{
//		//				"水",
//		//				"火",
//		//				"风",
//		//				"雷",
//		//				"庄"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			if item.Bettype == 4 {
//		//				rec.BetAmount = fmt.Sprintf("%v", bAmount)
//		//			} else {
//		//				rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//			}
//		//		case 4:
//		//			rec.RoundResult = redblackResult(w.db2, item.Roundcode)
//		//			playDef := []string{
//		//				"红方",
//		//				"黑方",
//		//				"幸运一击"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//		case 5:
//		//			playDef := []string{
//		//				"闲家",
//		//				"抢庄"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v倍", item.Odds)
//		//			rec.RoundResult = ox5Result(w.db2, item.Roundcode)
//		//		case 6:
//		//			rec.RoundResult = dtResult(w.db2, item.Roundcode)
//		//			playDef := []string{
//		//				"龙",
//		//				"虎",
//		//				"和"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//		case 7:
//		//			rec.RoundResult = benzResult(w.db2, item.Roundcode)
//		//			playDef := []string{"小大众", "小宝马", "小奔驰", "小保时捷", "大奔驰", "大宝马", "大奔驰", "大保时捷"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//		case 9:
//		//			rec.RoundResult = bacResult(w.db2, item.Roundcode)
//		//			playDef := []string{"庄", "闲", "和", "庄对", "闲对"}
//		//			rec.Play = playDef[item.Bettype]
//		//			rec.PayOff = item.Payoffvalue
//		//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//		//		}
//		//
//		//		rsp.Records = append(rsp.Records, rec)
//		//	}
//		//	if rows.Err() != nil {
//		//		rsp.RetCode = errcode.DBError
//		//		glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//		//		s.Response(rsp, mid)
//		//		return
//		//	}
//		//	s.Response(rsp, mid)
//		//	return
//		case 1:
//			//斗地主
//			err := w.db2.QueryRow(`select count(1),sum(0) as bets,sum(payoffvalue) as payoffs from settlement_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			if err != nil {
//				rsp.RetCode = errcode.DBError
//				glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//				s.Response(rsp, mid)
//				return
//			}
//			rows, err := w.db2.Query(`select settletime,roundcode,bettype,odds,payoffvalue
//		from settlement_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			if err != nil {
//				rsp.RetCode = errcode.DBError
//				glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//				s.Response(rsp, mid)
//				return
//			}
//			defer rows.Close()
//
//			for rows.Next() {
//				item := Settlement{}
//				err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Bettype, &item.Odds, &item.Payoffvalue)
//				if err != nil {
//					rsp.RetCode = errcode.DBError
//					glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//					s.Response(rsp, mid)
//					return
//				}
//				rec := &plr.S0000038_RecordRound{}
//				rec.SettleTime = item.Settletime
//				rec.RoundName = item.Roundcode
//				ss := strings.Split(item.Roundcode, "_")
//				if len(ss) != 3 {
//					rsp.RetCode = errcode.DBError
//					glog.SErrorf("Query settlement_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//					s.Response(rsp, mid)
//					return
//				}
//				rid, _ := strconv.Atoi(ss[1])
//				rec.RoomID = int32(rid)
//				rec.RoundResult = landlordResult(item.Bettype, item.Payoffvalue)
//				rec.Play = landlordPlay(item.Bettype)
//				rec.PayOff = item.Payoffvalue
//				rec.BetAmount = landlordAmount(item.Odds)
//				rsp.Records = append(rsp.Records, rec)
//			}
//			if rows.Err() != nil {
//				rsp.RetCode = errcode.DBError
//				glog.SErrorf("Query settlement_t failed %v.param:%v", err, msg)
//				s.Response(rsp, mid)
//				return
//			}
//			s.Response(rsp, mid)
//			return
//			//case 2:
//			//	//炸金花
//			//	err := w.db2.QueryRow(`select count(1),sum(actualamount) as bets,sum(payoffvalue) as payoffs from settlement_threecard_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_threecard_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue
//			//from settlement_threecard_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_threecard_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_threecard_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_threecard_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		rec.Play, rec.RoundResult = threePlayAndResult(w.db2, item.Roundcode, uid)
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_threecard_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 3:
//			//	//百人牛牛
//			//	err := w.db2.QueryRow(`select count(1),sum(betamount) as bets,sum(payoffvalue) as payoffs from settlement_ox100_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox100_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,betamount,payoffvalue,bettype
//			//from settlement_ox100_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox100_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_ox100_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_ox100_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		var bAmount float64
//			//		rec.RoundResult, bAmount = ox100Result(w.db2, item.Roundcode)
//			//		playDef := []string{
//			//			"水",
//			//			"火",
//			//			"风",
//			//			"雷",
//			//			"庄"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		if item.Bettype == 4 {
//			//			rec.BetAmount = fmt.Sprintf("%v", bAmount)
//			//		} else {
//			//			rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		}
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox100_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 4:
//			//	//红黑大战
//			//	err := w.db2.QueryRow(`select count(1),sum(actualamount),sum(payoffvalue) from settlement_redblack_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_redblack_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue,bettype
//			//from settlement_redblack_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_redblack_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_redblack_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_redblack_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		rec.RoundResult = redblackResult(w.db2, item.Roundcode)
//			//		playDef := []string{
//			//			"红方",
//			//			"黑方",
//			//			"幸运一击"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_redblack_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 5:
//			//	//抢庄牛牛
//			//	err := w.db2.QueryRow(`select count(1),sum(0),sum(payoffvalue) from settlement_ox5_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox5_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue,bettype,odds
//			//from settlement_ox5_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox5_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype, &item.Odds)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_ox5_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_ox5_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		playDef := []string{
//			//			"闲家",
//			//			"抢庄"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v倍", item.Odds)
//			//		rec.RoundResult = ox5Result(w.db2, item.Roundcode)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_ox100_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 6:
//			//	//龙虎
//			//	err := w.db2.QueryRow(`select count(1),sum(actualamount),sum(payoffvalue) from settlement_dt_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_dt_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue,bettype
//			//from settlement_dt_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_dt_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_dt_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_dt_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		rec.RoundResult = dtResult(w.db2, item.Roundcode)
//			//		playDef := []string{
//			//			"龙",
//			//			"虎",
//			//			"和"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_dt_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 7:
//			//	//奔驰宝马
//			//	err := w.db2.QueryRow(`select count(1),sum(actualamount),sum(payoffvalue) from settlement_benz_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_benz_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue,bettype
//			//from settlement_benz_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_benz_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_benz_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_benz_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		rec.RoundResult = benzResult(w.db2, item.Roundcode)
//			//		playDef := []string{"小大众", "小宝马", "小奔驰", "小保时捷", "大奔驰", "大宝马", "大奔驰", "大保时捷"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_benz_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//			//case 9:
//			//	//百家乐
//			//	err := w.db2.QueryRow(`select count(1),sum(actualamount),sum(payoffvalue) from settlement_baccarat_t where userid = ? and settletime between ? and ?`, uid, sTime, eTime).Scan(&rsp.TotalNums, &rsp.TotalBets, &rsp.ToTalPayoff)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_baccarat_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	rows, err := w.db2.Query(`select settletime,roundcode,actualamount,payoffvalue,bettype
//			//from settlement_baccarat_t where userid = ? and settletime between ? and ? order by settletime desc limit ?,?`, uid, sTime, eTime, sIndex, nums)
//			//	if err != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_baccarat_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	defer rows.Close()
//			//
//			//	for rows.Next() {
//			//		item := Settlement{}
//			//		err = rows.Scan(&item.Settletime, &item.Roundcode, &item.Actualamount, &item.Payoffvalue, &item.Bettype)
//			//		if err != nil {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_baccarat_t failed %v.param:%v", err, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rec := &plr.S0000038_RecordRound{}
//			//		rec.SettleTime = item.Settletime
//			//		rec.RoundName = item.Roundcode
//			//		ss := strings.Split(item.Roundcode, "_")
//			//		if len(ss) != 3 {
//			//			rsp.RetCode = errcode.DBError
//			//			glog.SErrorf("Query settlement_baccarat_t failed.roundcode:%v.param:%v", item.Roundcode, msg)
//			//			s.Response(rsp, mid)
//			//			return
//			//		}
//			//		rid, _ := strconv.Atoi(ss[1])
//			//		rec.RoomID = int32(rid)
//			//		rec.RoundResult = bacResult(w.db2, item.Roundcode)
//			//		playDef := []string{"庄", "闲", "和", "庄对", "闲对"}
//			//		rec.Play = playDef[item.Bettype]
//			//		rec.PayOff = item.Payoffvalue
//			//		rec.BetAmount = fmt.Sprintf("%v", item.Actualamount)
//			//		rsp.Records = append(rsp.Records, rec)
//			//	}
//			//	if rows.Err() != nil {
//			//		rsp.RetCode = errcode.DBError
//			//		glog.SErrorf("Query settlement_baccarat_t failed %v.param:%v", err, msg)
//			//		s.Response(rsp, mid)
//			//		return
//			//	}
//			//	s.Response(rsp, mid)
//			//	return
//		}
//		s.Response(rsp, mid)
//		return
//	}()
//	return nil
//}

func landlordResult(betType int32, payoff float64) string {
	switch betType {
	case 1:
		if payoff > 0 {
			return "地主胜利"
		}
		return "农民胜利"

	case 2:
		if payoff > 0 {
			return "农民胜利"
		}
		return "地主胜利"
	}
	return ""
}

func landlordPlay(betType int32) string {
	if betType == 1 {
		return "地主"
	}
	return "农民"
}

func landlordAmount(odds float64) string {
	return fmt.Sprintf("%v倍", odds)
}

//func threePlayAndResult(db *sql.DB, roundcode string, uid string) (play, rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := three.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//
//	for i := len(his.Actions) - 1; i >= 0; i-- {
//		if his.Actions[i].UserID == uid {
//			ty := his.Actions[i].OpType
//			switch ty {
//			case three.OpAllin:
//				play = "全押"
//			case three.OpBet:
//				play = "跟注"
//			case three.OpFold:
//				play = "弃牌"
//			case three.OpRaise:
//				play = "加注"
//			case three.OpVs:
//				play = "比牌"
//			}
//		}
//		if play != "" {
//			break
//		}
//	}
//
//	for i := range his.Players {
//		if his.Players[i].UserID == uid {
//			rc = fmt.Sprintf("%v-%v", his.Players[i].HandType, his.Players[i].Hand)
//		}
//	}
//
//	return
//}

//func redblackResult(db *sql.DB, roundcode string) (rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := rb.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//
//	if len(his.Cards) == 2 {
//		rc = fmt.Sprintf("%v-%v:%v-%v", his.Cards[0].Type, his.Cards[0].Cards, his.Cards[1].Type, his.Cards[1].Cards)
//	}
//
//	return
//}

//func ox100Result(db *sql.DB, roundcode string) (rc string, amount float64) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := ox100.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//	if len(his.Cards) == 5 {
//		for i := 0; i < 5; i++ {
//			rc += fmt.Sprintf("%v-%v", his.Cards[i].Type, his.Cards[i].Cards)
//			if i != 4 {
//				rc += ":"
//			}
//		}
//	}
//	for i := range his.Settles {
//		if i != 4 {
//			for _, v := range his.Settles[i] {
//				amount += v.BetAmount
//			}
//		}
//	}
//	return
//}

//func ox5Result(db *sql.DB, roundcode string) (rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := ox5.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//	for i := range his.Players {
//		if his.Players[i].UserID == his.BankerUid {
//			rc += fmt.Sprintf("%v-%v-%v-%v", his.Players[i].Kind, his.Players[i].Hand, his.Players[i].UserID, his.BankerScore)
//		}
//	}
//
//	for i := range his.Players {
//		if his.Players[i].UserID != his.BankerUid {
//			var score int32
//			for _, v := range his.BidActions {
//				if v.OpType == 1 && v.UserID == his.Players[i].UserID {
//					score = v.Value
//					break
//				}
//			}
//			if score == 0 {
//				score = 5
//			}
//			rc += fmt.Sprintf(":%v-%v-%v-%v", his.Players[i].Kind, his.Players[i].Hand, his.Players[i].UserID, score)
//		}
//	}
//	return
//}

//func dtResult(db *sql.DB, roundcode string) (rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := dt.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//
//	if len(his.Cards) == 2 {
//		rc = fmt.Sprintf("%v:%v", his.Cards[0], his.Cards[1])
//	}
//
//	return
//}

//func benzResult(db *sql.DB, roundcode string) (rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := benz.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//	rDef := []string{"小大众", "小宝马", "小奔驰", "小保时捷", "大奔驰", "大宝马", "大奔驰", "大保时捷"}
//	if his.Result >= 0 && his.Result < 8 {
//		rc = rDef[his.Result]
//	}
//	return
//}

//func bacResult(db *sql.DB, roundcode string) (rc string) {
//	var data []byte
//	err := db.QueryRow(`select history from round_record_t where roundname = ?`, roundcode).Scan(&data)
//	if err != nil {
//		glog.SErrorf("Query record failed %v", err)
//		return
//	}
//	his, err := bac.Decode(data)
//	if err != nil {
//		glog.SErrorf("Decode record failed %v", err)
//		return
//	}
//	rc = fmt.Sprintf("%v-%v:%v-%v", his.BankerCards.Point, his.BankerCards.Cards,
//		his.PlayerCards.Point, his.PlayerCards.Cards)
//	return
//}

func queryKindByRID(db *sql.DB, rid int32) int32 {
	var kid int32
	err := db.QueryRow(`select gamekindid from gameroom_t where gameroomid = ? `, rid).Scan(&kid)
	if err != nil {
		glog.SErrorf("Query kindid failed %v", err)
		return 0
	}
	return kid
}
