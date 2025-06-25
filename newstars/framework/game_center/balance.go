package game_center

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"newstars/framework/consts"
	"newstars/framework/glog"
	"newstars/framework/model"
	"newstars/framework/model/data"
	"newstars/framework/util"
	"time"
)

func GetBalance(token string) (model.GetBalanceRespData, error) {
	defer util.PrintElapsedTime("GetBalance", time.Now())

	var response model.GetBalanceRespData
	url := data.BaseConfig.CenterGameUrl + "/v1/game/get_balance"
	key := data.BaseConfig.CenterGameKey

	d := make(map[string]any)
	d["app_id"] = data.BaseConfig.CenterGameAppid
	d["time"] = time.Now().Unix()
	d["non_str"] = util.GetTraceId()
	d["token"] = token

	d["sign"] = sign(d, key)
	jsonValue, _ := json.Marshal(d)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		glog.SError("[game_center] GetBalance Error! creating request:", err)
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		glog.SError("[game_center] GetBalance Error! making request:", err)
		return response, err
	}
	defer resp.Body.Close()

	// 读取响应
	var result model.CenterGetBalanceResp
	json.NewDecoder(resp.Body).Decode(&result)

	glog.SInfo("[game_center] GetBalance req:", url, result)

	if result.Code == 1 {
		response.Currency = result.Data.Currency
		response.Balance = GetRealMoney(result.Data.Balance)
		glog.SInfo("[game_center] GetBalance Success.", response)
		return response, nil
	}

	return response, fmt.Errorf("[game_center] GetBalance error:%v", result.Msg)
}

type ChangeBalanceReq struct {
	UID       string
	Type      int
	Money     int
	SessionID string
	EndRound  bool
	OrderId   string
}

//changeBalanceReq := model.ChangeBalanceReq{
//	AppID:            data.BaseConfig.CenterGameAppid,
//	Type:             req.Type,
//	GameID:           FISH_REAL_GAME_ID,
//	Money:            req.Money,
//	Currency:         claims.Currency,
//	PlayerLoginToken: claims.Token,
//	SessionID:        req.SessionID,
//	Timestamp:        int(time.Now().Unix()),
//	Uname:            claims.AccountName,
//	EndRound:         req.EndRound,
//}

func ChangeBalance(req *model.ChangeBalanceReq) (model.ChangeBalanceRespData, error) {
	defer util.PrintElapsedTime("ChangeBalance", time.Now())

	var response model.ChangeBalanceRespData

	if req.OrderID == "" {
		//回撤必须要带订单号(有回撤的业务需要自己记录下注订单);
		if req.Type == consts.CHANGE_BALANCE_REFUND {
			return response, errors.New("the withdrawal must have an order number")
		}
		req.OrderID = util.GetTraceId()
	}

	if req.Type == consts.CHANGE_BALANCE_REFUND {
		req.CancelOrderId = req.OrderID
		req.OrderID = ""
	}

	return changeBalance(req)
}

func changeBalance(info *model.ChangeBalanceReq) (model.ChangeBalanceRespData, error) {
	var response model.ChangeBalanceRespData
	betAmount := ""
	winAmount := ""
	money := info.Money
	if money < 0 {
		money = -money
	}
	if info.Type == consts.CHANGE_BALANCE_BET {
		betAmount = SetRealMoney(money)
	} else {
		winAmount = SetRealMoney(money)
	}

	url := data.BaseConfig.CenterGameUrl + "/v1/game/change_balance_v1"
	key := data.BaseConfig.CenterGameKey

	d := make(map[string]any)
	d["app_id"] = info.AppID
	d["time"] = info.Timestamp
	d["token"] = info.PlayerLoginToken
	d["game_id"] = info.GameID
	d["currency"] = info.Currency
	d["order_id"] = info.OrderID
	d["session_id"] = info.SessionID
	d["cancel_order_id"] = info.CancelOrderId
	d["bet_amount"] = betAmount
	d["win_amount"] = winAmount

	d["sign"] = sign(d, key)
	jsonValue, _ := json.Marshal(d)

	glog.SInfo("[game_center] balance Request:", url, string(jsonValue))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		glog.SError("[game_center] balance update Error! creating request:", err)
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		glog.SError("[game_center] balance update Error! making request:", err, req)
		return response, err
	}
	defer resp.Body.Close()

	// 读取响应
	var result model.CenterChangeBalanceResp
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Code == 1 {
		//
		response.Balance = GetRealMoney(result.Data.Balance)
		response.Currency = info.Currency
		response.Rtp = result.Data.Rtp

		glog.SInfo("[game_center] balance update success.", result)
		return response, nil
	}

	glog.SError("[game_center] balance update failed!", result)
	return response, fmt.Errorf("change balance error: %s", result.Msg)
}

func GetRealMoney(amount string) int {
	balance := 0
	balance = util.ToInt(util.ToFloat64(amount) * util.ToFloat64(100))
	return balance
}

func SetRealMoney(amount int) string {
	return fmt.Sprintf("%.2f", float32(amount)/float32(100))
}
