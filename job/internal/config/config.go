package config

type Config struct {
	Log LogConfig
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
