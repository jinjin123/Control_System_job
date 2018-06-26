package main

import (
	"jiacrontab/client/store"
	"jiacrontab/libs/rpc"
	"log"
	"os"
	"os/signal"
	"time"

	"jiacrontab/model"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func newScheduler(config *config, crontab *crontab, store *store.Store) *scheduler {
	return &scheduler{
		config:  config,
		crontab: crontab,
		store:   store,
	}
}

type scheduler struct {
	config  *config
	crontab *crontab
	store   *store.Store
}

var globalConfig *config
var globalCrontab *crontab
var globalStore *store.Store
var globalDepend *depend
var globalDaemon *daemon
var startTime = time.Now()

func listenSignal(fn func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		sign := <-c
		log.Println("get signal:", sign)
		if fn != nil {
			fn()
		}
		log.Fatal("trying to exit gracefully...")

	}
}

func main() {

	model.CreateDB("sqlite3", "data/jiacrontab_client.db")
	model.DB().CreateTable(&model.DaemonTask{})
	model.DB().AutoMigrate(&model.DaemonTask{})

	globalConfig = newConfig()
	if globalConfig.debug == true {
		initPprof(globalConfig.pprofAddr)
	}

	globalStore = store.NewStore(globalConfig.dataFile)
	globalStore.Load()
	globalStore.Sync()

	globalCrontab = newCrontab(10)
	globalCrontab.run()

	globalDepend = newDepend()
	globalDepend.run()

	globalDaemon = newDaemon(100)
	globalDaemon.run()

	go listenSignal(func() {
		globalCrontab.lock.Lock()
		for k, v := range globalCrontab.handleMap {
			for _, item := range v.taskPool {
				item.cancel()
			}
			log.Printf("kill %s", k)
		}
		globalCrontab.lock.Unlock()
		globalStore.Update(func(s *store.Store) {
			for _, v := range s.TaskList {
				v.NumberProcess = 0
			}
		})

		globalDaemon.lock.Lock()
		for _, v := range globalDaemon.taskMap {
			v.cancel()
		}
		globalDaemon.lock.Unlock()
		globalDaemon.waitDone()

	})

	go RpcHeartBeat()
	rpc.ListenAndServe(globalConfig.rpcListenAddr, &DaemonTask{}, &Admin{}, &CrontabTask{})
}
