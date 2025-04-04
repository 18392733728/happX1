package config

type Config struct {
	Server struct {
		Port string `yaml:"port" default:":8080"`
	} `yaml:"server"`

	Database struct {
		Host     string `yaml:"host" default:"localhost"`
		Port     string `yaml:"port" default:"5432"`
		User     string `yaml:"user" default:"postgres"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname" default:"happx1"`
	} `yaml:"database"`
}

var AppConfig Config
