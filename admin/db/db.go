package db

import (
	"admin/db/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"xorm.io/core"
	"xorm.io/xorm"
)

var (
	dbmod  *DBModule
	db     *xorm.Engine
	logger *log.Entry
	Ctx    iris.Context
)

type DBModule struct {
	connString   string
	maxIdleConns int
	maxOpenConns int
	showSQL      bool
}

func NewDBModule() *DBModule { // constructor
	logger = log.WithField("source", "db")
	dbmod = &DBModule{}
	if err := dbmod.configuration(); err != nil {
		return nil
	}
	err := dbmod.Init()
	if err != nil {
		logger.Fatalf("init db module err:%v", err)
	}
	return dbmod
}

func (d *DBModule) configuration() error {
	d.connString = viper.GetString("database.connect")
	d.maxIdleConns = viper.GetInt("database.max_idle_conns")
	d.maxOpenConns = viper.GetInt("database.max_open_conns")
	d.showSQL = viper.GetBool("database.showsql")
	return nil
}

func (d *DBModule) Init() error {
	//logger.Debugf("mysql:%s", d.connString)
	database, err := xorm.NewEngine("mysql", d.connString)
	if err != nil {
		return err
	}
	db = database
	db.SetMapper(core.GonicMapper{})
	db.SetMaxIdleConns(d.maxIdleConns)
	db.SetMaxOpenConns(d.maxOpenConns)
	db.ShowSQL(d.showSQL)

	err = d.syncSchema()
	
	return err
}

func (d *DBModule) syncSchema() error {
	err := db.StoreEngine("InnoDB").Sync2(
		new(model.Config),
	) // init model here

	// logger.Println("sync schema")
	return err
}
