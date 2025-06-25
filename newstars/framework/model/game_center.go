package model

import "github.com/golang-jwt/jwt/v4"

type CustomClaims struct {
	Token       string `json:"token"` //用户唯一标记
	Time        int    `json:"time"`
	NickName    string `json:"nickName"`
	AccountName string `json:"accountName"`
	MerchantId  string `json:"merchantId"`
	Currency    string `json:"currency"`
	jwt.RegisteredClaims
}

type CenterGetBalanceResp struct {
	Data CenterGetBalanceData `json:"data"`
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
}

type CenterGetBalanceData struct {
	Token    string `json:"token"`
	Balance  string `json:"balance"`
	Currency string `json:"currency"`
}

type CenterChangeBalanceResp struct {
	Data CenterChangeBalanceData `json:"data"`
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
}

type CenterChangeBalanceData struct {
	Rtp      int    `json:"rtp"`
	Balance  string `json:"balance"`
	Token    string `json:"token"`
	Currency string `json:"currency"`
}
