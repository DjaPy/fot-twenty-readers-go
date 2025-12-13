package config

type Conf struct {
	System struct {
		BaseUrl string `yaml:"baseUrl"`
	}
	Telegram struct {
		BotToken string `yaml:"botToken"`
	}
}

func NewConf() {
	println("new conf")
}
