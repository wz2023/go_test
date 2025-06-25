package data

import (
	"github.com/mitchellh/mapstructure"
	"newstars/framework/db"
	"newstars/framework/glog"
	"newstars/framework/model"
)

var BaseConfig GameConfig

type GameConfig struct {
	SelfGameJwt     string `json:"self_game_jwt" gorm:"self_game_jwt"`
	CenterGameAppid string `json:"centergame_appid" gorm:"centergame_appid"`
	CenterGameKey   string `json:"centergame_key" gorm:"centergame_key"`
	CenterGameUrl   string `json:"centergame_url" gorm:"centergame_url"`
}

func InitBaseConfig() {
	rowList := make([]model.GameConfig, 0)
	db.GormDB.Table("game_config").Find(&rowList)

	cfgBaseMap := map[string]any{}
	for _, row := range rowList {
		cfgBaseMap[row.Key] = row.Value
	}

	gameConfig := GameConfig{}

	{
		mapstructureDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &gameConfig, WeaklyTypedInput: true})
		if err != nil {
			glog.SError("baseconfig数据初始化失败!", err)
			return
		}

		if err := mapstructureDecoder.Decode(cfgBaseMap); err != nil {
			glog.SError("baseconfig数据初始化失败!", err)
			return
		}
	}

	BaseConfig = gameConfig
}
