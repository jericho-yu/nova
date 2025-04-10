package redisPool

import "nova/src/util/honestMan"

type RedisSetting struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Prefix   string `yaml:"prefix"`
	Pool     []struct {
		Key    string `yaml:"key"`
		Prefix string `yaml:"prefix"`
		DbNum  int    `yaml:"dbNum"`
	}
}

var RedisSettingApp RedisSetting

// New 初始化：数据库配置
func (*RedisSetting) New(path string) *RedisSetting {
	var redisSetting *RedisSetting = &RedisSetting{}
	err := honestMan.HonestManApp.New(path).LoadYaml(redisSetting)
	if err != nil {
		return nil
	}

	return redisSetting
}

// ExampleYaml 示例配置文件
func (*RedisSetting) ExampleYaml() string {
	return `host: 127.0.0.1
port: 6379
password: ""
prefix: "abc-example"
pool:
  [
    {
      key: "auth",
      prefix: "auth",
      dbNum: 0
    }
  ]`
}
