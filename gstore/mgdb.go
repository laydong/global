package gstore

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDb struct {
	Dsn             string `mapstructure:"dsn" json:"dsn" yaml:"dsn"`                                              // 服务器信息
	ConnTimeOut     uint64 `mapstructure:"conn-time-out" json:"conn_time_out" yaml:"conn-time-out"`                // 空闲中的最大连接数
	ConnMaxPoolSize uint64 `mapstructure:"conn-max-pool-size" json:"conn_max_pool_size" yaml:"conn-max-pool-size"` // 打开到数据库的最大连接数
}

func InitMongoDb(mongodb MongoDb) *mongo.Client {
	MdbOptions := options.Client().
		ApplyURI(mongodb.Dsn).
		SetMaxPoolSize(mongodb.ConnMaxPoolSize).
		SetMinPoolSize(mongodb.ConnTimeOut)
	db, err := mongo.NewClient(MdbOptions)
	if err != nil {
		log.Printf("[app.gstore] mgdb error: %v", err.Error())
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = db.Connect(ctx)
	if err != nil {
		log.Printf("[app.gstore] mgdb error: %v", err.Error())
		panic(err)
	}
	log.Printf("[app.gstore] mongo success")
	return db
}
