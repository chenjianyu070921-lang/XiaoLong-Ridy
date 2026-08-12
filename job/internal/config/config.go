package config

type Config struct {
	DB    DBConfig
	Redis RedisConfig
	Log   LogConfig
}

type DBConfig struct {
	DataSource string
}

type RedisConfig struct {
	Host string
	Pass string
	Type string
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
