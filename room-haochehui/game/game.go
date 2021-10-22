package game

import (
	"fmt"
	// "github.com/lonng/nano/serialize/protobuf"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"

	"math/rand"
	"time"
	// "gitlab.com/wolfplus/gamespace-lhd/db"
)

var (
	version     = "" // 游戏版本
	forceUpdate = false
	logger      = log.WithField("source", "game")
)

func Startup() {

	rand.Seed(time.Now().Unix())
	version = viper.GetString("update.version")
	forceUpdate = viper.GetBool("update.force")

	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}

	logger.Infof("当前游戏服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("game service starup")
	// register game handler
	comps := &component.Components{}
	comps.Register(defaultManager)

	addr := fmt.Sprintf(":%d", viper.GetInt("game-server.port"))
	nano.Listen(addr,
		nano.WithIsWebsocket(true),
		// nano.WithDebugMode(),
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("source", "nano")),
		// nano.WithSerializer(protobuf.NewSerializer()),
		nano.WithSerializer(json.NewSerializer()), // override default serializer
		nano.WithComponents(comps),
	)
}
