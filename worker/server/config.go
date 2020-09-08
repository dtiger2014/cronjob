package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config : worker config
type Config struct {
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	MysqlUser       string   `json:"mysql_user"`
	MysqlPass       string   `json:"mysql_pass"`
	MysqlHost       string   `json:"mysql_host"`
	MysqlPort       string   `json:"mysql_port"`
	MysqlDatabase   string   `json:"mysql_database"`
	MysqlCharset    string   `json:"mysql_charset"`
}

var (
	// GConfig : global config
	GConfig *Config
)

// InitConfig : init config
func InitConfig(filename string) error {
	var (
		content []byte
		conf    Config
		err     error
	)

	// read file
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	// json unmarshal
	if err = json.Unmarshal(content, &conf); err != nil {
		return err
	}

	// set
	GConfig = &conf

	fmt.Printf("%+v\n", *GConfig)

	return nil
}
