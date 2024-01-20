package router

import "log/slog"

type RouterStrategyType string

const (
	RoundRobinStrategy      RouterStrategyType = "round-robin"
	LeastConnectionStrategy RouterStrategyType = "least-connection"
)

type routerStrategy interface {
	GetAvailableServer(router *Router) *routerServer
}

func newRouterStrategy(strategyType RouterStrategyType) routerStrategy {
	switch strategyType {
	case RoundRobinStrategy:
		return &simpleRoundRobinRouterStrategy{}
	case LeastConnectionStrategy:
		return &leastConnectionServerStrategy{}
	default:
		return &simpleRoundRobinRouterStrategy{}
	}
}

type simpleRoundRobinRouterStrategy struct{}

// Implement simple round robin
func (s *simpleRoundRobinRouterStrategy) GetAvailableServer(r *Router) *routerServer {
	serverIndex := r.requestCount % r.serverCount
	slog.Debug("Simple Round Robin Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]
}

type leastConnectionServerStrategy struct{}

// Implement least busy using active connections
func (s *leastConnectionServerStrategy) GetAvailableServer(r *Router) *routerServer {
	var serverIndex = 0
	for k, server := range r.servers {
		if server.activeConnections <= r.servers[serverIndex].activeConnections {
			serverIndex = k
		}
	}
	slog.Debug("Least Busy Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]
}
