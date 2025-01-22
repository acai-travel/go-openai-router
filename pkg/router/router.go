package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/acai-travel/go-openai-router/v2/pkg/server"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
)

type Router struct {
	servers      []*server.RouterServer
	serverCount  int
	requestCount int
	strategy     routerStrategy
}

// NewRouter creates a new Router instance with the given server configurations and strategy type.
// It returns a pointer to the Router and an error if any.
// The serverConfigs parameter is a slice of server.ServerConfig that contains the configurations for each server.
// The strategyType parameter is the type of router strategy to be used.
// If the serverConfigs slice is empty, it returns an error with the message "empty server config".
// Otherwise, it creates a new RouterServer for each server configuration and adds them to the servers slice.
// Finally, it initializes the Router with the servers, serverCount, requestCount, and strategy.
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
func (r *Router) GetChatCompletions(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
	r.requestCount++
	modelName := body.Model.String()
	server := r.strategy.GetAvailableServer(r, modelName)
	if server == nil {
		return nil, fmt.Errorf("no server available for model %s", modelName)
	}
	return server.NewCompletion(ctx, body, opts...)
}

// GetChatCompletionsStream - Return the chat completions for a given prompt as a sequence of events.
// If the operation fails it returns an *azcore.ResponseError type.
//   - options - GetCompletionsOptions contains the optional parameters for the Client.GetCompletions method.
func (r *Router) GetChatCompletionsStream(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	r.requestCount++
	modelName := body.Model.String()
	server := r.strategy.GetAvailableServer(r, modelName)
	if server == nil {
		return nil, fmt.Errorf("no server available for model %s", modelName)
	}
	return server.NewStreamingCompletion(ctx, body, opts...), nil
}
