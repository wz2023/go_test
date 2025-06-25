package model

type VerifySessionResp struct {
	Data VerifySessionRespData `json:"data"`
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
}
type VerifySessionRespData struct {
	Uname                   string `json:"uname"`
	Nickname                string `json:"nickname"`
	Balance                 int    `json:"balance"`
	Currency                string `json:"currency"`
	Avatar                  string `json:"avatar"`
	UserCreatedAt           int    `json:"user_created_at"`
	UserChannelCode         string `json:"user_channel_code"`
	UserDistributionChannel string `json:"user_distribution_channel" form:"user_distribution_channel"`
	UserCountry             string `json:"user_country"`
	UserTagId               int    `json:"user_tag_id"`
}

type GetBalanceResp struct {
	Data GetBalanceRespData `json:"data"`
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
}
type GetBalanceRespData struct {
	Balance  int    `json:"balance"`
	Currency string `json:"currency"`
}

type ChangeBalanceResp struct {
	Data ChangeBalanceRespData `json:"data"`
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
}
type ChangeBalanceRespData struct {
	Rtp      int    `json:"rtp"`
	Balance  int    `json:"balance"`
	Currency string `json:"currency"`
}

type ChangeBalanceReq struct {
	AppID            string `json:"app_id"`
	Bet              int    `json:"bet"`
	Type             int    `json:"type"`
	GameID           int    `json:"game_id"`
	Money            int    `json:"money"`
	Currency         string `json:"currency"`
	OrderID          string `json:"order_id"`
	PlayerLoginToken string `json:"player_login_token"`
	SessionID        string `json:"session_id"`
	Timestamp        int    `json:"timestamp"`
	Uname            string `json:"uname"`
	EndRound         bool   `json:"end_round"`
	CancelOrderId    string `json:"cancel_order_id"`
}
