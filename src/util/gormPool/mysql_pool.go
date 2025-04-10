package gormPool

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type MySqlPool struct {
	username  string
	password  string
	host      string
	port      uint16
	database  string
	charset   string
	sources   map[string]*MySqlConnection
	replicas  map[string]*MySqlConnection
	mainDsn   *Dsn
	mainConn  *gorm.DB
	dbSetting *DbSetting
}

var (
	mysqlPoolIns   *MySqlPool
	mysqlPoolOnce  sync.Once
	MySqlDsnFormat = "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local"
	MySqlPoolApp   MySqlPool
)

func (*MySqlPool) Once(dbSetting *DbSetting) GormPool { return OnceMySqlPool(dbSetting) }

// OnceMySqlPool 单例化：mysql链接池
//
//go:fix 推荐使用：Once方法
func OnceMySqlPool(dbSetting *DbSetting) GormPool {
	mysqlPoolOnce.Do(func() {
		mysqlPoolIns = &MySqlPool{
			username: dbSetting.MySql.Main.Username,
			password: dbSetting.MySql.Main.Password,
			host:     dbSetting.MySql.Main.Host,
			port:     dbSetting.MySql.Main.Port,
			database: dbSetting.MySql.Database,
			charset:  dbSetting.MySql.Charset,
			sources:  dbSetting.MySql.Sources,
			replicas: dbSetting.MySql.Replicas,

			dbSetting: dbSetting,
		}
	})

	var (
		err      error
		dbConfig *gorm.Config
	)

	// 配置主库
	mysqlPoolIns.mainDsn = &Dsn{
		Name: "main",
		Content: fmt.Sprintf(
			MySqlDsnFormat,
			dbSetting.MySql.Main.Username,
			dbSetting.MySql.Main.Password,
			dbSetting.MySql.Main.Host,
			dbSetting.MySql.Main.Port,
			dbSetting.MySql.Database,
			dbSetting.MySql.Charset,
		),
	}

	// 数据库配置
	dbConfig = &gorm.Config{
		PrepareStmt:                              true,  // 预编译
		CreateBatchSize:                          500,   // 批量操作
		DisableForeignKeyConstraintWhenMigrating: true,  // 禁止自动创建外键
		SkipDefaultTransaction:                   false, // 开启自动事务
		QueryFields:                              true,  // 查询字段
		AllowGlobalUpdate:                        false, // 不允许全局修改,必须带有条件
	}

	// 配置主库
	mysqlPoolIns.mainConn, err = gorm.Open(mysql.Open(mysqlPoolIns.mainDsn.Content), dbConfig)
	if err != nil {
		panic(fmt.Sprintf("配置主库失败：%s", err.Error()))
	}

	mysqlPoolIns.mainConn = mysqlPoolIns.mainConn.Session(&gorm.Session{})
	{
		sqlDb, _ := mysqlPoolIns.mainConn.DB()
		sqlDb.SetConnMaxIdleTime(time.Duration(mysqlPoolIns.dbSetting.Common.MaxIdleTime) * time.Hour)
		sqlDb.SetConnMaxLifetime(time.Duration(mysqlPoolIns.dbSetting.Common.MaxLifetime) * time.Hour)
		sqlDb.SetMaxIdleConns(mysqlPoolIns.dbSetting.Common.MaxIdleConnections)
		sqlDb.SetMaxOpenConns(mysqlPoolIns.dbSetting.Common.MaxOpenConnections)
	}

	return mysqlPoolIns
}

// GetConn 获取主数据库链接
func (my *MySqlPool) GetConn() *gorm.DB {
	my.getRws()
	return my.mainConn
}

// getRws 获取带有读写分离的数据库链接
func (my *MySqlPool) getRws() *gorm.DB {
	var (
		err                                 error
		sourceDialectors, replicaDialectors []gorm.Dialector
		sources                             []*Dsn
		replicas                            []*Dsn
	)
	// 配置写库
	if len(my.sources) > 0 {
		sources = make([]*Dsn, 0)
		for idx, item := range my.sources {
			sources = append(sources, &Dsn{
				Name: idx,
				Content: fmt.Sprintf(
					MySqlDsnFormat,
					item.Username,
					item.Password,
					item.Host,
					item.Port,
					my.dbSetting.MySql.Database,
					my.dbSetting.MySql.Charset,
				),
			})
		}
	}

	// 配置读库
	if len(my.replicas) > 0 {
		replicas = make([]*Dsn, 0)
		for idx, item := range my.replicas {
			replicas = append(replicas, &Dsn{
				Name: idx,
				Content: fmt.Sprintf(
					MySqlDsnFormat,
					item.Username,
					item.Password,
					item.Host,
					item.Port,
					my.dbSetting.MySql.Database,
					my.dbSetting.MySql.Charset,
				),
			})
		}
	}

	if len(sources) > 0 {
		sourceDialectors = make([]gorm.Dialector, len(sources))
		for i := 0; i < len(sources); i++ {
			sourceDialectors[i] = mysql.Open(sources[i].Content)
		}
	}

	if len(replicas) > 0 {
		replicaDialectors = make([]gorm.Dialector, len(replicas))
		for i := 0; i < len(replicas); i++ {
			replicaDialectors[i] = mysql.Open(replicas[i].Content)
		}
	}

	err = my.mainConn.Use(
		dbresolver.Register(dbresolver.Config{
			Sources:           sourceDialectors,          // 写库
			Replicas:          replicaDialectors,         // 读库
			Policy:            dbresolver.RandomPolicy{}, // 策略
			TraceResolverMode: true,
		}).
			SetConnMaxIdleTime(time.Duration(my.dbSetting.Common.MaxIdleTime) * time.Hour).
			SetConnMaxLifetime(time.Duration(my.dbSetting.Common.MaxLifetime) * time.Hour).
			SetMaxIdleConns(my.dbSetting.Common.MaxIdleConnections).
			SetMaxOpenConns(my.dbSetting.Common.MaxOpenConnections),
	)
	if err != nil {
		panic(fmt.Errorf("数据库链接错误：%s", err.Error()))
	}

	return my.mainConn
}

// Close 关闭数据库链接
func (my *MySqlPool) Close() error {
	if my.mainConn != nil {
		db, err := my.mainConn.DB()
		if err != nil {
			return fmt.Errorf("关闭数据库链接失败：获取数据库链接失败 %s", err.Error())
		}
		err = db.Close()
		if err != nil {
			return fmt.Errorf("关闭数据库连接失败 %s", err.Error())
		}
	}

	return nil
}
