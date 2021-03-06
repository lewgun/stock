package db

import (
	"database/sql"
	"github.com/DejaMi/mymysql-pool"
	"github.com/carusyte/stock/model"
	"github.com/carusyte/stock/util"
	_ "github.com/ziutek/mymysql/godrv"
	_ "github.com/ziutek/mymysql/native"
	"gopkg.in/gorp.v2"
)

var p, e = pool.New(pool.Config{Address: "127.0.0.1:3306", Protocol: "tcp", Username: "mysql", Password: "123456",
Database: "secu", MaxConnections: 100, MaxConnectionAge: 60, ConnectTimeout: 60, RequestTimeout: 60,
KeepConnectionsAlive: true})

func Get(create, truncate bool) *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("mymysql", "tcp:localhost:3306*secu/mysql/123456")
	util.CheckErr(err, "sql.Open failed,")

	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(64)

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "utf8"}}

	dbmap.AddTableWithName(model.KlineW{}, "kline_w").SetKeys(false, "Code", "Date", "Klid")
	dbmap.AddTableWithName(model.KlineM{}, "kline_m").SetKeys(false, "Code", "Date", "Klid")
	dbmap.AddTableWithName(model.Indicator{}, "indicator_d").SetKeys(false, "Code", "Date", "Klid")
	dbmap.AddTableWithName(model.IndicatorW{}, "indicator_w").SetKeys(false, "Code", "Date", "Klid")
	dbmap.AddTableWithName(model.IndicatorM{}, "indicator_m").SetKeys(false, "Code", "Date", "Klid")
	dbmap.AddTableWithName(model.IndcFeatRaw{}, "indc_feat_raw").SetKeys(false, "Code", "Indc", "Fid")
	if create {
		err = dbmap.CreateTablesIfNotExists()
		util.CheckErr(err, "Create tables failed,")
	}
	if truncate {
		err = dbmap.TruncateTables()
		util.CheckErr(err, "Truncate tables failed,")
	}

	util.CheckErr(db.Ping(), "Failed to ping db,")

	return dbmap
}

func GetMySql() (c *pool.Conn) {
	c,e := p.Get()
	util.CheckErrNop(e,"failed to get connection from pool")
	return
}
