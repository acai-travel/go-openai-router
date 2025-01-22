package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/azure"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
)

type ServerConfigType string

const (
	AzureOpenAiServerType ServerConfigType = "azure-openai"
	OpenAiServerType      ServerConfigType = "openai"
)

// ServerConfig represents the configuration for the server.
type ServerConfig struct {
	Endpoint        string
	ApiKey          string
	Type            ServerConfigType
	AvailableModels []string // AvailableModels is a list of models that are available for the Azure endpoint. The list of models will vary based on the endpoint.
}

// RouterServer represents the server that the router will use to send requests.
type RouterServer struct {
	client            *openai.Client
	ActiveConnections int
	Latency           int64
	Type              ServerConfigType
	totalRequests     int64
	totalLatency      int64
	AvailableModels   []string // AvailableModels is a list of models that are available for the Azure endpoint. The list of models will vary based on the endpoint.
}

// Hardcoded, should fix this
const azureOpenAIAPIVersion = "2024-06-01"

func NewRouterServer(serverConfig ServerConfig) (*RouterServer, error) {
	var err error
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
	switch serverConfig.Type {
	case AzureOpenAiServerType:
		client := openai.NewClient(
			azure.WithEndpoint(serverConfig.Endpoint, azureOpenAIAPIVersion),
			azure.WithAPIKey(serverConfig.ApiKey),
		)
		server.client = client
		return server, err
	case OpenAiServerType:
		client := openai.NewClient(
			option.WithAPIKey(serverConfig.ApiKey),
		)
		server.client = client
		return server, err
	default:
		return nil, fmt.Errorf("server type %s is not supported", serverConfig.Type)
	}
}

// Returns the completion.
// If the operation fails it returns an error type
//   - options - ChatCompletionNewParams contains the optional parameters for the Client.Chat.Completions.New method.
func (s *RouterServer) NewCompletion(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
	s.preFlight()
	start := time.Now()
	defer s.postFlight(start)
	return s.client.Chat.Completions.New(ctx, body, opts...)
}

// Streams the completion.
// If the operation fails it returns an error type
//   - options - ChatCompletionNewParams contains the optional parameters for the Client.Chat.Completions.NewStreaming method.
func (s *RouterServer) NewStreamingCompletion(ctx context.Context, body openai.ChatCompletionNewParams, options ...option.RequestOption) *ssestream.Stream[openai.ChatCompletionChunk] {
	s.preFlight()
	start := time.Now()
	defer s.postFlight(start)
	return s.client.Chat.Completions.NewStreaming(ctx, body, options...)
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
