package game_center

import (
	"context"
	"encoding/json"
	"newstars/framework/glog"
	"newstars/framework/model"
	"newstars/framework/model/cachekey"
	"newstars/framework/redisx"
	"time"
)

func getUserCustomClaimsByID(uid string) (*model.CustomClaims, error) {
	data, err := redisx.Get(context.Background(), cachekey.GetUserKey(uid)).Result()
	if err != nil {
		glog.SErrorf("Failed to get user info, uid:%v err:%v", uid, err)
		return nil, err
	}

	var customClaims model.CustomClaims
	err = json.Unmarshal([]byte(data), &customClaims)
	if err != nil {
		glog.SErrorf("Failed Json Unmarshal,%v", err)
		return nil, err
	}
	return &customClaims, nil
}

func GetUserInfoByID(uid string) (*model.UserInfo, error) {
	// 尝试从缓存获取
	cacheKey := cachekey.GetUserInfoKey(uid)
	cachedData, err := redisx.Get(context.Background(), cacheKey).Result()
	if err == nil && cachedData != "" {
		var user model.UserInfo
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			return &user, nil
		}
	}

	claims, err := getUserCustomClaimsByID(uid)
	if err != nil {
		return nil, err
	}

	balanceResp, err := GetBalance(claims.Token)
	if err != nil {
		glog.SErrorf("Failed to GetBalance,%v", err)
		return nil, err
	}

	nickName := claims.NickName
	if len(nickName) == 0 {
		nickName = claims.AccountName
	}

	userInfo := &model.UserInfo{
		UserID:      claims.Token,
		NickName:    claims.NickName,
		DisPlayName: claims.AccountName,
		Currency:    claims.Currency,
		Wealth:      float64(balanceResp.Balance),
		Status:      1,
		FaceID:      0,
		Sexuality:   0,
		AccType:     0,
		GameID:      0,
		Profit:      0,
		FaceFrameID: 0,
	}
	marshal, _ := json.Marshal(userInfo)

	// 缓存不存在则保存进缓存
	redisx.Set(context.Background(), cacheKey, string(marshal), 30*time.Minute)
	return userInfo, nil
}
