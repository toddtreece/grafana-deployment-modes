package modules

import (
	"errors"
	"os"

	"github.com/go-kit/log"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/services"
	"github.com/grafana/grafana-deployment-modes/pkg/logger"
	"github.com/grafana/grafana-deployment-modes/pkg/sender"
)

const (
	All         = "all"
	Sender      = "sender"
	LocalLogger = "local-logger"
	Client      = "client"
	Server      = "server"
)

var goKitLogger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

type Modules struct {
	Targets        []string
	ModuleManager  *modules.Manager
	ServiceManager *services.Manager
	ServiceMap     map[string]services.Service
	logger         logger.Logger
}

func New(targets []string) (*Modules, error) {
	mm := modules.NewManager(goKitLogger)
	m := &Modules{
		Targets:       targets,
		ModuleManager: mm,
	}

	deps := map[string][]string{
		Server:      {},
		Client:      {},
		LocalLogger: {},
		Sender:      {},
		All:         {Sender},
	}

	if m.isModuleEnabled(All) {
		// if target "all" is enabled, we can use a local stdout logger
		deps[Sender] = append(deps[Sender], LocalLogger)
	} else {
		// otherwise, we need to use a TCP client logger
		deps[Sender] = append(deps[Sender], Client)
	}

	mm.RegisterModule(Server, m.initServer)
	mm.RegisterModule(LocalLogger, m.initLocalLogger, modules.UserInvisibleModule)
	mm.RegisterModule(Client, m.initClient, modules.UserInvisibleModule)
	mm.RegisterModule(Sender, m.initSender)
	mm.RegisterModule(All, nil)

	for mod, t := range deps {
		if err := mm.AddDependency(mod, t...); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Modules) initLocalLogger() (services.Service, error) {
	l := logger.NewLocalLogger()
	m.logger = l
	return l, nil
}

func (m *Modules) initClient() (services.Service, error) {
	l := logger.NewClient()
	m.logger = l
	return l, nil
}

func (m *Modules) initServer() (services.Service, error) {
	return logger.NewServer(), nil
}

func (m *Modules) initSender() (services.Service, error) {
	if m.logger == nil {
		return nil, errors.New("logger not initialized")
	}
	return sender.New(m.logger), nil
}
