package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"mox/drivers/worker"
	core "mox/internal"
	asyncexec "mox/pkg/async"
	"mox/pkg/driver"
	driverv2 "mox/pkg/driver/v2"
	"mox/tools/utils"
	"mox/use_cases/workercore"
)

var _ (driver.IDriver) = (*DaemonAdapter)(nil)

const DaemonAdapterName = "DaemonAdapter"

type DaemonAdapter struct {
	app    core.App
	cmd    *asyncexec.Cmd
	worker *workercore.Worker
	l      *sync.RWMutex
}

// Close implements [driver.IDriver].
func (d *DaemonAdapter) Close() error {
	if err := d.cmd.Cancel(); err != nil {
		return err
	}

	return nil
}

func (d *DaemonAdapter) runHaproxy(file *os.File) error {
	utils.LookupExecutablePathAbs("haproxy")

	// d.app.Driver().Instance()
	executable, err := utils.LookupExecutablePathAbs("haproxy")
	if err != nil {
		d.app.Logger().Error(err.Error())
		return err
	}

	argsValidate := []string{"-f", "haproxy.cfg"}
	cmd := asyncexec.Command(d.app.Context(), executable, argsValidate...)

	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// cmd.Stdout = &logs.SlogWriter{Logger: logger, Level: slog.LevelInfo, App: "HAPROXY"}
	// cmd.Stderr = &logs.SlogWriter{Logger: logger, Level: slog.LevelError, App: "HAPROXY"}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = []*os.File{file}

	fmt.Printf("Master: Oper FD dari file %s (FD asli: %d) ke ExtraFiles[0]\n", file.Name(), file.Fd())

	fdOrderEnv := fmt.Sprintf("FD_ORDER=%d", file.Fd())
	pidEnv := fmt.Sprintf("PID=%d", d.worker.PID())
	cmd.Env = append(os.Environ(), fdOrderEnv, "APP_VERSION=v1.1", pidEnv)

	if err := cmd.AsyncRun(); err != nil {
		d.app.Logger().Error("process starting failed", slog.String("err", err.Error()))
		return err
	}

	d.cmd = cmd

	go func(cmd *asyncexec.Cmd) {
		<-cmd.Terminated
		d.app.Logger().Info(fmt.Sprintf("process %d terminated : %s", cmd.Process.Pid, cmd.Status()))

		// exit if termination signal was received and the last process terminated abnormally
		if cmd.ProcessState.ExitCode() != 0 {
			d.app.Stop()
			return
		}

		// remove the process from tracking
		d.l.Lock()
		defer d.l.Unlock()
		// for i := range cmds {
		// 	if cmds[i].Process.Pid == cmd.Process.Pid {
		// 		cmds = append(cmds[:i], cmds[i+1:]...)
		// 		break
		// 	}
		// }
		//
		// // exit if there are no more processes running
		// if len(cmds) == 0 {
		// 	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		// 		os.Exit(cmd.ProcessState.ExitCode())
		// 	} else {
		// 		os.Exit(0)
		// 	}
		// }
	}(cmd)

	d.app.Logger().Info(fmt.Sprintf("process started with pid %d and status %s", cmd.Process.Pid, cmd.Status()))

	d.l.Lock()
	defer d.l.Unlock()

	return nil
}

// Init implements [driver.IDriver].
func (d *DaemonAdapter) Init() error {
	d.app.Logger().Info("running daemon driver")

	worker, err := driverv2.Get[*workercore.Worker](d.app.Driver(), worker.WorkerAdapterName)
	if err != nil {
		d.app.Logger().Error(err.Error())
		return err
	}

	if worker.ExtraFile != nil {
		d.worker = worker

		if err := d.runHaproxy(worker.ExtraFile); err != nil {
			return nil
		}
	}

	d.app.Logger().Info("daemon running")

	return nil
}

// Instance implements [driver.IDriver].
func (d *DaemonAdapter) Instance() interface{} {
	return nil
}

// Name implements [driver.IDriver].
func (d *DaemonAdapter) Name() string {
	return DaemonAdapterName
}

func NewDaemonAdapter(app core.App) *DaemonAdapter {
	return &DaemonAdapter{app: app, l: &sync.RWMutex{}}
}
