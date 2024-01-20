package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type ServerConfigType string
type RouterStrategy string

const (
	AzureOpenAi ServerConfigType = "azure-openai"
	OpenAi      ServerConfigType = "openai"
	RoundRobin  RouterStrategy   = "round-robin"
	LeastBusy   RouterStrategy   = "least-busy"
)

type ServerConfig struct {
	Endpoint string
	ApiKey   string
	Type     ServerConfigType
}

type RouterServer struct {
	Client            *azopenai.Client
	activeConnections int
}

type Router struct {
	servers             []*RouterServer
	lastUsedServerIndex int
	serverCount         int
	RequestCount        int
	strategy            RouterStrategy
}

// NewRouter creates an instance of the Router that routes requests to one or more openai/azureopenai servers
func NewRouter(serverConfigs []ServerConfig, strategy RouterStrategy) (*Router, error) {
	servers := []*RouterServer{}
	for _, serverConfig := range serverConfigs {
		server, err := newRouterServer(serverConfig)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	slog.Debug("Creating New Router", "strategy", strategy)
	return &Router{
		servers:             servers,
		lastUsedServerIndex: 0,
		serverCount:         len(servers),
		RequestCount:        0,
		strategy:            strategy,
	}, nil
}

func newRouterServer(serverConfig ServerConfig) (*RouterServer, error) {
	server := &RouterServer{
		activeConnections: 0,
	}
	keyCredential := azcore.NewKeyCredential(serverConfig.ApiKey)
	switch serverConfig.Type {
	case AzureOpenAi:
		client, err := azopenai.NewClientWithKeyCredential(serverConfig.Endpoint, keyCredential, nil)
		server.Client = client
		return server, err
	case OpenAi:
		client, err := azopenai.NewClientForOpenAI(serverConfig.Endpoint, keyCredential, nil)
		server.Client = client
		return server, err
	default:
		return nil, fmt.Errorf("server type %s is not supported", serverConfig.Type)
	}
}

func (r *Router) getAvailableServer() *RouterServer {
	var serverIndex = 0
	switch r.strategy {
	case RoundRobin:
		serverIndex = r.getSimpleRoundRobinServerIndex()
	case LeastBusy:
		serverIndex = r.getLeastBusyServerIndex()
	default:
		serverIndex = r.getSimpleRoundRobinServerIndex()
	}
	r.lastUsedServerIndex = serverIndex
	slog.Debug("Providing Next Available Server", "serverIndex", serverIndex)
	return r.servers[serverIndex]
}

// Implement least busy using active connections
func (r *Router) getLeastBusyServerIndex() int {
	var serverIndex = 0
	for k, server := range r.servers {
		if server.activeConnections <= r.servers[serverIndex].activeConnections {
			serverIndex = k
		}
	}
	return serverIndex
}

// Implement simple round robin
func (r *Router) getSimpleRoundRobinServerIndex() int {
	return r.RequestCount % r.serverCount
}

// GetChatCompletions - Gets chat completions for the provided chat messages. Completions support a wide variety of tasks
// and generate text that continues from or "completes" provided prompt data.
// If the operation fails it returns an *azcore.ResponseError type.
func (r *Router) GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
	r.RequestCount++
	server := r.getAvailableServer()
	server.activeConnections++
	resp, err := server.Client.GetChatCompletions(ctx, body, options)
	server.activeConnections--
	return resp, err
}

// GetChatCompletionsStream - Return the chat completions for a given prompt as a sequence of events.
// If the operation fails it returns an *azcore.ResponseError type.
//   - options - GetCompletionsOptions contains the optional parameters for the Client.GetCompletions method.
func (r *Router) GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
	r.RequestCount++
	server := r.getAvailableServer()
	server.activeConnections++
	resp, err := server.Client.GetChatCompletionsStream(ctx, body, options)
	server.activeConnections--
	return resp, err
}
