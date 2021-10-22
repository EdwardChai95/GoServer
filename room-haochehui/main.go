package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/game"
)

func main() {

	viper.SetConfigType("toml")
	viper.SetConfigFile("./configs/config.toml")
	viper.ReadInConfig()

	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	if viper.GetBool("core.debug") {
		log.SetLevel(log.DebugLevel)
	}
	logger := log.WithField("source", "main")

	db := db.NewDBModule()
	err := db.Init()
	if err != nil {
		logger.Fatalf("init db module err:%v", err)
	}
	logger.Info("before wait")
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() { defer wg.Done(); game.Startup() }() // 开启web服务器

	// go func() {
	// 	defer wg.Done()
	// 	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// 	http.ListenAndServe(":3001", nil)
	// }()

	wg.Wait()
}
