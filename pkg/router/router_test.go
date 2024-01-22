package router

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/acai-travel/go-openai-router/pkg/server"
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
	router.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{}, nil)
	if router.requestCount != 1 {
		t.Fatalf("Incorrect requests count %d", router.requestCount)
	}
}

func TestGetChatCompletionsStream(t *testing.T) {
	router := getRouter()
	router.GetChatCompletionsStream(context.TODO(), azopenai.ChatCompletionsOptions{}, nil)
	if router.requestCount != 1 {
		t.Fatalf("Incorrect requests count %d", router.requestCount)
	}
}

func getRouter() *Router {
	router, _ := NewRouter([]server.ServerConfig{
		{
			Type: server.OpenAiServerType,
		},
	}, RoundRobinStrategy)
	return router
}
