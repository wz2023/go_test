package config

type Config struct {
	Zap    Zap             `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis  Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Mysql  Mysql           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Pgsql  Pgsql           `mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
	DBList []SpecializedDB `mapstructure:"db-list" json:"db-list" yaml:"db-list"`
	System System          `mapstructure:"system" json:"system" yaml:"system"`
}
