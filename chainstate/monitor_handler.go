package chainstate

import "github.com/mrz1836/go-whatsonchain"

type MonitorHandler interface {
	whatsonchain.SocketHandler
	SetMonitor(monitor *Monitor)
}
