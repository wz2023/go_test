package config

type System struct {
	Env    string `mapstructure:"env" json:"env" yaml:"env"`             // 环境值
	Host   string `mapstructure:"host" json:"host" yaml:"host"`          // 主机地址
	DbType string `mapstructure:"db-type" json:"db-type" yaml:"db-type"` // 数据库类型:mysql(默认)|sqlite|sqlserver|postgresql
}
