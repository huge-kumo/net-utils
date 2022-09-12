package config

var (
	Conf *Config
)

type Config struct {
	Storage *Storage
}

type Storage struct {
	FilePath string
	DBName   string
}

func init() {
	Conf = &Config{Storage: &Storage{
		FilePath: "test.db",
		DBName:   "test",
	}}
}
