package router

import (
	"log/slog"
	"slices"

	"github.com/acai-travel/go-openai-router/v2/pkg/server"
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
	GetAvailableServer(router *Router, modelName string) *server.RouterServer
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

// GetAvailableServer returns an available server from the router based on the provided model name.
// It filters the servers based on the available models and uses a simple round-robin strategy to select a server.
// If no server is available for the given model, it returns nil.
func (s *simpleRoundRobinRouterStrategy) GetAvailableServer(r *Router, modelName string) *server.RouterServer {
	filteredServers := []*server.RouterServer{}
	for _, server := range r.servers {
		if slices.Contains(server.AvailableModels, modelName) {
			filteredServers = append(filteredServers, server)
		}
	}
	if len(filteredServers) == 0 {
		return nil
	}
	serverIndex := r.requestCount % len(filteredServers)
	slog.Debug("Simple Round Robin Server", "serverIndex", serverIndex)
	return filteredServers[serverIndex]
}

type leastConnectionServerStrategy struct{}

// GetAvailableServer returns the server with the least active connections that supports the specified model.
// If no server is available for the model, it returns nil.
func (s *leastConnectionServerStrategy) GetAvailableServer(r *Router, modelName string) *server.RouterServer {
	filteredServers := make([]*server.RouterServer, 0)
	for _, server := range r.servers {
		if slices.Contains(server.AvailableModels, modelName) {
			filteredServers = append(filteredServers, server)
		}
	}

	if len(filteredServers) == 0 {
		return nil
	}

	minConnectionsServer := filteredServers[0]
	for _, server := range filteredServers {
		if server.ActiveConnections < minConnectionsServer.ActiveConnections {
			minConnectionsServer = server
		}
	}
	return minConnectionsServer
}

type leastLatencyServerStrategy struct{}

// GetAvailableServer returns the server with the least latency that supports the specified model.
// If no server is available for the model, it returns nil.
func (s *leastLatencyServerStrategy) GetAvailableServer(r *Router, modelName string) *server.RouterServer {
	filteredServers := make([]*server.RouterServer, 0)
	for _, server := range r.servers {
		if slices.Contains(server.AvailableModels, modelName) {
			filteredServers = append(filteredServers, server)
		}
	}

	if len(filteredServers) == 0 {
		return nil
	}

	leastLatencyServer := filteredServers[0]
	for _, server := range filteredServers {
		if server.Latency < leastLatencyServer.Latency {
			leastLatencyServer = server
		}
	}
	return leastLatencyServer
}
