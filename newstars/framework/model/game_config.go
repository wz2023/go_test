package model

type GameConfig struct {
	Key      string `json:"key" gorm:"key"`
	Value    string `json:"value" gorm:"value"`
	Describe string `json:"describe" gorm:"describe"`
}
