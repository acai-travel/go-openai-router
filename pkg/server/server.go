package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type ServerConfigType string

const (
	AzureOpenAiServerType ServerConfigType = "azure-openai"
	OpenAiServerType      ServerConfigType = "openai"
)

type ServerConfig struct {
	Endpoint        string
	ApiKey          string
	Type            ServerConfigType
	AvailableModels []string // AvailableModels is a list of models that are available for the Azure endpoint. The list of models will vary based on the endpoint.
}

type RouterServer struct {
	client            *azopenai.Client
	ActiveConnections int
	Latency           int64
	Type              ServerConfigType
	totalRequests     int64
	totalLatency      int64
	AvailableModels   []string // AvailableModels is a list of models that are available for the Azure endpoint. The list of models will vary based on the endpoint.
}

func NewRouterServer(serverConfig ServerConfig) (*RouterServer, error) {
	if len(serverConfig.ApiKey) == 0 {
		return nil, fmt.Errorf("empty api key")
	}
	if len(serverConfig.Endpoint) == 0 {
		return nil, fmt.Errorf("empty endpoint")
	}
	if len(serverConfig.AvailableModels) == 0 {
		return nil, fmt.Errorf("empty available models")
	}
	server := &RouterServer{
		ActiveConnections: 0,
		totalRequests:     0,
		Latency:           0,
		totalLatency:      0,
		Type:              serverConfig.Type,
		AvailableModels:   serverConfig.AvailableModels,
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
func (s *RouterServer) GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
	s.preFlight()
	start := time.Now()
	defer s.postFlight(start)
	return s.client.GetChatCompletions(ctx, body, options)
}

// GetChatCompletionsStream - Return the chat completions for a given prompt as a sequence of events.
// If the operation fails it returns an *azcore.ResponseError type.
//   - options - GetCompletionsOptions contains the optional parameters for the Client.GetCompletions method.
func (s *RouterServer) GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
	s.preFlight()
	start := time.Now()
	defer s.postFlight(start)
	return s.client.GetChatCompletionsStream(ctx, body, options)
}

func (s *RouterServer) preFlight() {
	s.ActiveConnections++
}

func (s *RouterServer) postFlight(start time.Time) {
	elapsed := time.Since(start)
	s.ActiveConnections--
	s.totalRequests++
	s.totalLatency += elapsed.Milliseconds()
	s.Latency = s.totalLatency / s.totalRequests
	slog.Debug("Average Latency for Server", "averageLatency", s.Latency, "totalLatency", s.totalLatency, "numberOfRequests", s.totalRequests)
}
