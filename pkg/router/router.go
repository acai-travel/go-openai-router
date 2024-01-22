package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/acai-travel/go-openai-router/pkg/server"
)

type Router struct {
	servers      []*server.RouterServer
	serverCount  int
	requestCount int
	strategy     routerStrategy
}

// NewRouter creates an instance of the Router that routes requests to one or more openai/azureopenai servers
func NewRouter(serverConfigs []server.ServerConfig, strategyType RouterStrategyType) (*Router, error) {
	servers := []*server.RouterServer{}
	if len(serverConfigs) == 0 {
		return nil, fmt.Errorf("empty server config")
	}
	for _, serverConfig := range serverConfigs {
		server, err := server.NewRouterServer(serverConfig)
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

// GetChatCompletions - Gets chat completions for the provided chat messages. Completions support a wide variety of tasks
// and generate text that continues from or "completes" provided prompt data.
// If the operation fails it returns an *azcore.ResponseError type.
func (r *Router) GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
	r.requestCount++
	return r.strategy.GetAvailableServer(r).GetChatCompletions(ctx, body, options)
}

// GetChatCompletionsStream - Return the chat completions for a given prompt as a sequence of events.
// If the operation fails it returns an *azcore.ResponseError type.
//   - options - GetCompletionsOptions contains the optional parameters for the Client.GetCompletions method.
func (r *Router) GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
	r.requestCount++
	return r.strategy.GetAvailableServer(r).GetChatCompletionsStream(ctx, body, options)
}
