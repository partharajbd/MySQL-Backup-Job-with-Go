package configs

type DBConfig struct {
	Name string `json:"name"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Host string `json:"host"`
	Port string `json:"port"`
}
