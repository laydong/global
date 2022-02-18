package gstore

import (
	"github.com/olivere/elastic/v6"
	"log"
)

type EsConfigDB struct {
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`             // 服务器地址:端口
	Dbname   string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`     // 默认索引数据库名
	Username string `mapstructure:"username" json:"username" yaml:"username"` // 数据库用户名
	Password string `mapstructure:"password" json:"password" yaml:"password"` // 数据库密码
}

func InitEs(esConfig EsConfigDB) *elastic.Client {
	// 创建client连接ES
	if esConfig.Username == "" || esConfig.Password == "" {
		db, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(esConfig.Addr))
		if err != nil {
			log.Printf("[app.gstore] elastic error: %v", err.Error())
			panic(err)
		}
		return db
	} else {
		db, err := elastic.NewClient(
			elastic.SetSniff(false),
			// elasticsearch 服务地址，多个服务地址使用逗号分隔
			elastic.SetURL(esConfig.Addr),
			// 基于http base auth验证机制的账号和密码
			elastic.SetBasicAuth(esConfig.Username, esConfig.Password),
		)
		if err != nil {
			log.Printf("[app.gstore] elastic error: %v", err.Error())
			panic(err)
		}
		return db
	}
}
