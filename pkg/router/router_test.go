package router

import (
	"context"
	"testing"

	"github.com/acai-travel/go-openai-router/v2/pkg/server"
	"github.com/openai/openai-go"
)

func TestNewRouter(t *testing.T) {
	_, err := NewRouter([]server.ServerConfig{}, RoundRobinStrategy)
	if err == nil {
		t.Fatal("NewRouter should have errored for an empty server config")
	}
	_, err = NewRouter([]server.ServerConfig{
		{
			Type: server.OpenAiServerType,
		},
		{
			Type: "foo",
		},
	}, RoundRobinStrategy)
	if err == nil {
		t.Fatal("NewRouter should have errored for incorrect type in server config")
	}
}

func TestGetChatCompletions(t *testing.T) {
	router := getRouter()
	deploymentName := openai.ChatModelGPT3_5Turbo
	router.GetChatCompletions(context.TODO(), openai.ChatCompletionNewParams{
		Model: openai.F(deploymentName),
	})
	if router.requestCount != 1 {
		t.Fatalf("Incorrect requests count %d", router.requestCount)
	}
}

func TestGetChatCompletionsStream(t *testing.T) {
	router := getRouter()
	deploymentName := openai.ChatModelGPT3_5Turbo
	router.GetChatCompletions(context.TODO(), openai.ChatCompletionNewParams{
		Model: openai.F(deploymentName),
	})
	if router.requestCount != 1 {
		t.Fatalf("Incorrect requests count %d", router.requestCount)
	}
}

func getRouter() *Router {
	router, _ := NewRouter([]server.ServerConfig{
		{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://azure-openai.com",
			ApiKey:          "azure-openai-key",
			AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-turbo"},
		},
	}, RoundRobinStrategy)
	return router
}
