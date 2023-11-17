package config

var Config = struct {
	RedisConfig
}{}

type RedisConfig struct {
	Expire int `mapstructure:"expire" json:"expire"`
}
