package modules

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/services"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/signals"
)

// ********************************************************************************************************************
// ignore code below this line, it's essentially a direct copy from loki:
// https://github.com/grafana/loki/blob/d6a60b4a9f7d29bba1900b25eeadc35bbd5c858f/pkg/loki/loki.go#L448-L543
// ********************************************************************************************************************
func (m *Modules) Run() error {
	fmt.Println("Running modules", m.Targets)
	serviceMap, err := m.ModuleManager.InitModuleServices(m.Targets...)
	if err != nil {
		return err
	}

	m.ServiceMap = serviceMap

	var servs []services.Service
	for _, s := range serviceMap {
		servs = append(servs, s)
	}

	sm, err := services.NewManager(servs...)
	if err != nil {
		return err
	}

	m.ServiceManager = sm

	// Let's listen for events from this manager, and log them.
	healthy := func() { level.Info(goKitLogger).Log("msg", "started") }
	stopped := func() { level.Info(goKitLogger).Log("msg", "stopped") }
	serviceFailed := func(service services.Service) {
		// if any service fails, stop entire
		sm.StopAsync()

		// let's find out which module failed
		for m, s := range serviceMap {
			if s == service {
				if service.FailureCase() == modules.ErrStopProcess {
					level.Info(goKitLogger).Log("msg", "received stop signal via return error", "module", m, "error", service.FailureCase())
				} else {
					level.Error(goKitLogger).Log("msg", "module failed", "module", m, "error", service.FailureCase())
				}
				return
			}
		}

		level.Error(goKitLogger).Log("msg", "module failed", "module", "unknown", "error", service.FailureCase())
	}

	sm.AddListener(services.NewManagerListener(healthy, stopped, serviceFailed))

	signalLogger := logging.GoKit(goKitLogger)

	// Setup signal handler. If signal arrives, we stop the manager, which stops all the services.
	signalHandler := signals.NewHandler(signalLogger)
	go func() {
		signalHandler.Loop()
		sm.StopAsync()
	}()

	// Start all services. This can really only fail if some service is already
	// in other state than New, which should not be the case.
	err = sm.StartAsync(context.Background())
	if err == nil {
		// Wait until service manager stops. It can stop in two ways:
		// 1) Signal is received and manager is stopped.
		// 2) Any service fails.
		err = sm.AwaitStopped(context.Background())
	}

	// If there is no error yet (= service manager started and then stopped without problems),
	// but any service failed, report that failure as an error to caller.
	if err == nil {
		if failed := sm.ServicesByState()[services.Failed]; len(failed) > 0 {
			for _, f := range failed {
				if f.FailureCase() != modules.ErrStopProcess {
					// Details were reported via failure listener before
					err = errors.New("failed services")
					break
				}
			}
		}
	}

	return err
}

// ********************************************************************************************************************
// copied from loki
// https://github.com/grafana/loki/blob/d6a60b4a9f7d29bba1900b25eeadc35bbd5c858f/pkg/loki/loki.go#L316-L318
// ********************************************************************************************************************
func (m *Modules) isModuleEnabled(name string) bool {
	return StringsContain(m.Targets, name)
}

// ********************************************************************************************************************
// copied from loki
// https://github.com/grafana/loki/blob/d6a60b4a9f7d29bba1900b25eeadc35bbd5c858f/pkg/util/string.go#L28-L36
// ********************************************************************************************************************
func StringsContain(values []string, search string) bool {
	for _, v := range values {
		if search == v {
			return true
		}
	}

	return false
}
