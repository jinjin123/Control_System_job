package kproc

import (
	"context"
	"jiacrontab/pkg/file"
	"os/exec"
)

type KCmd struct {
	ctx context.Context
	*exec.Cmd
	isKillChildProcess bool
	done               chan struct{}
}

// SetEnv 设置环境变量
func (k *KCmd) SetEnv(env []string) {
	if len(env) == 0 {
		return
	}
	k.Cmd.Env = env
}

// SetDir 设置工作目录
func (k *KCmd) SetDir(dir string) {
	if dir == "" {
		return
	}
	if file.Exist(dir) == false {
		return
	}
	k.Cmd.Dir = dir
}

// SetExitKillChildProcess 设置主进程退出时是否kill子进程,默认kill
func (k *KCmd) SetExitKillChildProcess(ok bool) {
	k.isKillChildProcess = ok
}
