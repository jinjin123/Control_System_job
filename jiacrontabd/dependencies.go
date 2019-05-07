package jiacrontabd

import (
	"bytes"
	"context"
	"jiacrontab/pkg/proto"
	"time"

	"github.com/iwannay/log"
)

type depEntry struct {
	jobID       uint   // 定时任务id
	jobUniqueID string // job的唯一标志
	processID   int    // 当前依赖的父级任务（可能存在多个并发的task)
	id          string // depID
	workDir     string
	user        string
	env         []string
	from        string
	commands    []string
	dest        string
	done        bool
	timeout     int64
	err         error
	name        string
	logPath     string
	logContent  []byte
}

func newDependencies(jd *Jiacrontabd) *dependencies {
	return &dependencies{
		jd:  jd,
		dep: make(chan *depEntry, 100),
	}
}

type dependencies struct {
	jd  *Jiacrontabd
	dep chan *depEntry
}

func (d *dependencies) add(t *depEntry) {
	d.dep <- t
}

func (d *dependencies) run() {
	go func() {
		for {
			select {
			case t := <-d.dep:
				go d.exec(t)
			}
		}
	}()
}

// TODO: 主任务退出杀死依赖
func (d *dependencies) exec(task *depEntry) {

	var (
		reply     bool
		myCmdUnit cmdUint
		err       error
	)

	if task.timeout == 0 {
		// 默认超时10分钟
		task.timeout = 600
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.timeout)*time.Second)
	defer cancel()

	myCmdUnit.args = [][]string{task.commands}
	myCmdUnit.ctx = ctx
	myCmdUnit.dir = task.workDir
	myCmdUnit.user = task.user
	myCmdUnit.logPath = task.logPath
	myCmdUnit.jd = d.jd

	err = myCmdUnit.launch()

	log.Infof("exec %s %s cost %.4fs %v", task.name, task.commands, float64(myCmdUnit.costTime)/1000000000, err)

	task.logContent = bytes.TrimRight(myCmdUnit.content, "\x00")
	task.done = true
	task.err = err

	task.dest, task.from = task.from, task.dest

	if !d.jd.SetDependDone(task) {
		err = d.jd.rpcCall("Srv.SetDependDone", proto.DepJob{
			Name:        task.name,
			Dest:        task.dest,
			From:        task.from,
			ID:          task.id,
			JobUniqueID: task.jobUniqueID,
			ProcessID:   task.processID,
			JobID:       task.jobID,
			Commands:    task.commands,
			LogContent:  task.logContent,
			Err:         err,
			Timeout:     task.timeout,
		}, &reply)

		if err != nil {
			log.Error("Srv.SetDependDone error:", err, "server addr:", d.jd.getOpts().AdminAddr)
		}

		if !reply {
			log.Errorf("task %s %v call Srv.SetDependDone failed! err:%v", task.name, task.commands, err)
		}
	}
}
