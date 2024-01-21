package router

import (
	"log/slog"

	"github.com/acai-travel/go-openai-router/pkg/server"
)

type RouterStrategyType string

const (
	//Simple Round Robin Strategy to get a server
	RoundRobinStrategy RouterStrategyType = "round-robin"
	//Least Connection Strategy to get a server using active connections
	LeastConnectionStrategy RouterStrategyType = "least-connection"
	//Least Average Strategy to get a server
	LeastLatencyStrategy RouterStrategyType = "least-latency"
)

type routerStrategy interface {
	GetAvailableServer(router *Router) *server.RouterServer
}

func newRouterStrategy(strategyType RouterStrategyType) routerStrategy {
	switch strategyType {
	case RoundRobinStrategy:
		return &simpleRoundRobinRouterStrategy{}
	case LeastConnectionStrategy:
		return &leastConnectionServerStrategy{}
	case LeastLatencyStrategy:
		return &leastLatencyServerStrategy{}
	default:
		return &simpleRoundRobinRouterStrategy{}
	}
}

type simpleRoundRobinRouterStrategy struct{}

// Implement simple round robin
func (s *simpleRoundRobinRouterStrategy) GetAvailableServer(r *Router) *server.RouterServer {
	serverIndex := r.requestCount % r.serverCount
	slog.Debug("Simple Round Robin Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]
}

type leastConnectionServerStrategy struct{}

// Implement least busy using active connections
func (s *leastConnectionServerStrategy) GetAvailableServer(r *Router) *server.RouterServer {
	var serverIndex = 0
	for k, server := range r.servers {
		if server.ActiveConnections <= r.servers[serverIndex].ActiveConnections {
			serverIndex = k
		}
	}
	slog.Debug("Least Busy Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]
}

type leastLatencyServerStrategy struct{}

// Implement least latency using calculated average latency
func (s *leastLatencyServerStrategy) GetAvailableServer(r *Router) *server.RouterServer {
	var serverIndex = 0
	for k, server := range r.servers {
		if server.Latency <= r.servers[serverIndex].Latency {
			serverIndex = k
		}
	}
	slog.Debug("Least Latency Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]

}
