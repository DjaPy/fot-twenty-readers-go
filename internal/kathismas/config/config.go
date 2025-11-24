package config

type Conf struct {
	System struct {
		BaseUrl string `yaml:"baseUrl"`
	}
}

func NewConf() {
	println("new conf")
}
