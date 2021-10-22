package db

import (
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"xorm.io/xorm"

	"github.com/spf13/viper"
	"xorm.io/core"
)

var (
	db     *xorm.Engine
	logger *log.Entry
)

type DBModule struct {
	connString   string
	maxIdleConns int
	maxOpenConns int
	showSQL      bool
}

func ExportDb() *xorm.Engine {
	return db
}

func NewDBModule() *DBModule {
	logger = log.WithField("source", "db")
	db := &DBModule{}

	if err := db.configuration(); err != nil {
		return nil
	}
	return db
}
func (d *DBModule) configuration() error {
	d.connString = viper.GetString("database.connect")
	d.maxIdleConns = viper.GetInt("database.max_idle_conns")
	d.maxOpenConns = viper.GetInt("database.max_open_conns")
	d.showSQL = viper.GetBool("database.showsql")
	return nil
}

func (d *DBModule) Init() error {
	logger.Debugf("mysql:%s", d.connString)
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

func (d *DBModule) AfterInit() {
}

func (d *DBModule) BeforeShutdown() {
}

func (d *DBModule) Shutdown() error {
	db.Close()
	return nil
}

func (d *DBModule) syncSchema() error {
	err := db.StoreEngine("InnoDB").Sync2(
		new(model.GameCollection),
		new(model.DailyCollection),
		new(model.Account),
		new(model.User),
		new(model.UserProfile),
		new(model.UserGameData),
		new(model.Currency),
		// new(model.Payment),
		// new(model.PaymentLedger),
		new(model.Club),
		new(model.ClubUser),
		new(model.GameCoinLocker),
		new(model.GameCoinLockerHistory),
		new(model.Item),
		new(model.HighAndSpecialCustomer),
		new(model.GameCoinTransaction),
		new(model.CustomerServiceMessage),
		new(model.LogInformation),
		new(model.TempAnnouncement),
		new(model.ExchangeCode),
		new(model.NewYearEvent),
		new(model.Order),
		new(model.TotalGameCoin),
		new(model.VCode),
		new(model.Proxy),
		new(model.ProxyUser),
		new(model.CodeInformation),
	)

	logger.Println("re populating！！")
	return err
}
