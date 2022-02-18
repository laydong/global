package gstore

import (
	"github.com/climber-dong/global/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

const (
	defaultPoolMaxIdle     = 2                                 // 连接池空闲连接数量
	defaultPoolMaxOpen     = 13                                // 连接池最大连接数量4c*2+4只读副本+1主实例
	defaultConnMaxLifeTime = time.Second * time.Duration(7200) // MySQL默认长连接时间为8个小时,可根据高并发业务持续时间合理设置该值
	defaultConnMaxIdleTime = time.Second * time.Duration(600)  // 设置连接10分钟没有用到就断开连接(内存要求较高可降低该值)
	LevelInfo              = "info"
	LevelWarn              = "warn"
	LevelError             = "error"
)

type DbPoolCfg struct {
	MaxIdleConn int `json:"max_idle_conn"` //空闲连接数
	MaxOpenConn int `json:"max_open_conn"` //最大连接数
	MaxLifeTime int `json:"max_life_time"` //连接可重用的最大时间
	MaxIdleTime int `json:"max_idle_time"` //在关闭连接之前,连接可能处于空闲状态的最大时间
}

type dbConfig struct {
	poolCfg *DbPoolCfg
	gormCfg *gorm.Config
}

type DbConnFunc func(cfg *dbConfig)

// InitDB init db
func InitDB(dsn string, logLevel string, DbCfgFunc ...DbConnFunc) *gorm.DB {
	var err error
	var cfg dbConfig

	for _, f := range DbCfgFunc {
		f(&cfg)
	}

	var level logger.LogLevel
	switch logLevel {
	case LevelInfo:
		level = logger.Info
	case LevelWarn:
		level = logger.Warn
	case LevelError:
		level = logger.Error
	default:
		level = logger.Info
	}

	if cfg.gormCfg == nil {
		cfg.gormCfg = &gorm.Config{
			Logger: logx.Default(logx.GetSugar(), level),
		}
	} else {
		if cfg.gormCfg.Logger == nil {
			cfg.gormCfg.Logger = logx.Default(logx.GetSugar(), level)
		}
	}

	Db, err := gorm.Open(mysql.Open(dsn), cfg.gormCfg)
	if err != nil {
		log.Printf("[app.gstore] mysql open fail, err=%s", err)
		panic(err)
	}

	cfg.setDefaultPoolConfig(Db)

	err = DbSurvive(Db)
	if err != nil {
		log.Printf("[app.gstore] mysql survive fail, err=%s", err)
		panic(err)
	}

	log.Printf("[app.gstore] mysql success")
	return Db
}

// DbSurvive mysql survive
func DbSurvive(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// SetPoolConfig set poolx config
func SetPoolConfig(cfg DbPoolCfg) DbConnFunc {
	return func(c *dbConfig) {
		c.poolCfg = &cfg
	}
}

// SetGormConfig set gorm config
func SetGormConfig(cfg *gorm.Config) DbConnFunc {
	return func(c *dbConfig) {
		c.gormCfg = cfg
	}
}

func (c *dbConfig) setDefaultPoolConfig(db *gorm.DB) {
	d, err := db.DB()
	if err != nil {
		log.Printf("[app.gstore] mysql db fail, err=%s", err)
		panic(err)
	}
	var cfg = c.poolCfg
	if cfg == nil {
		d.SetMaxOpenConns(defaultPoolMaxOpen)
		d.SetMaxIdleConns(defaultPoolMaxIdle)
		d.SetConnMaxLifetime(defaultConnMaxLifeTime)
		d.SetConnMaxIdleTime(defaultConnMaxIdleTime)
		return
	}

	if cfg.MaxOpenConn == 0 {
		d.SetMaxOpenConns(defaultPoolMaxOpen)
	} else {
		d.SetMaxOpenConns(cfg.MaxOpenConn)
	}

	if cfg.MaxIdleConn == 0 {
		d.SetMaxIdleConns(defaultPoolMaxIdle)
	} else {
		d.SetMaxIdleConns(cfg.MaxIdleConn)
	}

	if cfg.MaxLifeTime == 0 {
		d.SetConnMaxLifetime(defaultConnMaxLifeTime)
	} else {
		d.SetConnMaxLifetime(time.Second * time.Duration(cfg.MaxLifeTime))
	}

	if cfg.MaxIdleTime == 0 {
		d.SetConnMaxIdleTime(defaultConnMaxIdleTime)
	} else {
		d.SetConnMaxIdleTime(time.Second * time.Duration(cfg.MaxIdleTime))
	}
}
