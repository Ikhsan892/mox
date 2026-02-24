package master

import (
	"context"
	"net"
	"sync"

	core "mox/internal"
	"mox/pkg/driver"
	"mox/use_cases/mastercore"
)

var _ (driver.IDriver) = (*MasterAdapter)(nil)

const MasterAdapterName = "MasterAdapter"

type MasterAdapter struct {
	app        core.App
	ctx        context.Context
	mastercore *mastercore.Master
	l          net.Listener // instance listener
	wg         *sync.WaitGroup
}

func NewMasterAdapter(ctx context.Context, app core.App) *MasterAdapter {
	return &MasterAdapter{app: app, ctx: ctx, wg: &sync.WaitGroup{}}
}

// Close implements [driver.IDriver].
func (m *MasterAdapter) Close() error {
	m.mastercore.Connections().CloseAllConnections()

	m.mastercore.Stop()

	m.app.Logger().Info("master closed")

	return nil
}

func (m *MasterAdapter) Init() error {
	ctx := m.app.Context()
	operations := RegisterCommand(m.app)

	master := mastercore.NewMasterCore(
		ctx,
		m.app,
	).SetOperations(operations)

	if err := master.Run(); err != nil {
		m.app.Logger().Error(err.Error())
		return err
	}

	m.app.Logger().Debug("master running")

	m.mastercore = master

	return nil
}

// Instance implements [driver.IDriver].
func (m *MasterAdapter) Instance() interface{} {
	return m.mastercore
}

// Name implements [driver.IDriver].
func (m *MasterAdapter) Name() string {
	return MasterAdapterName
}
