package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type ServerConfigType string

const (
	AzureOpenAiServerType ServerConfigType = "azure-openai"
	OpenAiServerType      ServerConfigType = "openai"
)

type ServerConfig struct {
	Endpoint string
	ApiKey   string
	Type     ServerConfigType
}

type routerServer struct {
	client            *azopenai.Client
	activeConnections int
}

type Router struct {
	servers      []*routerServer
	serverCount  int
	requestCount int
	strategy     routerStrategy
}

// NewRouter creates an instance of the Router that routes requests to one or more openai/azureopenai servers
func NewRouter(serverConfigs []ServerConfig, strategyType RouterStrategyType) (*Router, error) {
	servers := []*routerServer{}
	for _, serverConfig := range serverConfigs {
		server, err := newRouterServer(serverConfig)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	slog.Debug("Creating New Router", "strategy", strategyType)
	return &Router{
		servers:      servers,
		serverCount:  len(servers),
		requestCount: 0,
		strategy:     newRouterStrategy(strategyType),
	}, nil
}

func newRouterServer(serverConfig ServerConfig) (*routerServer, error) {
	server := &routerServer{
		activeConnections: 0,
	}
	keyCredential := azcore.NewKeyCredential(serverConfig.ApiKey)
	switch serverConfig.Type {
	case AzureOpenAiServerType:
		client, err := azopenai.NewClientWithKeyCredential(serverConfig.Endpoint, keyCredential, nil)
		server.client = client
		return server, err
	case OpenAiServerType:
		client, err := azopenai.NewClientForOpenAI(serverConfig.Endpoint, keyCredential, nil)
		server.client = client
		return server, err
	default:
		return nil, fmt.Errorf("server type %s is not supported", serverConfig.Type)
	}
}

// GetChatCompletions - Gets chat completions for the provided chat messages. Completions support a wide variety of tasks
// and generate text that continues from or "completes" provided prompt data.
// If the operation fails it returns an *azcore.ResponseError type.
func (r *Router) GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
	r.requestCount++
	server := r.strategy.GetAvailableServer(r)
	server.activeConnections++
	resp, err := server.client.GetChatCompletions(ctx, body, options)
	server.activeConnections--
	return resp, err
}

// GetChatCompletionsStream - Return the chat completions for a given prompt as a sequence of events.
// If the operation fails it returns an *azcore.ResponseError type.
//   - options - GetCompletionsOptions contains the optional parameters for the Client.GetCompletions method.
func (r *Router) GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
	r.requestCount++
	server := r.strategy.GetAvailableServer(r)
	server.activeConnections++
	resp, err := server.client.GetChatCompletionsStream(ctx, body, options)
	server.activeConnections--
	return resp, err
}
