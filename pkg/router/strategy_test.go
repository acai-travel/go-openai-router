package router

import (
	"reflect"
	"testing"

	"github.com/acai-travel/go-openai-router/pkg/server"
)

func TestNewRouterStrategy(t *testing.T) {
	s := newRouterStrategy(RoundRobinStrategy)
	if reflect.TypeOf(s) != reflect.TypeOf(&simpleRoundRobinRouterStrategy{}) {
		t.Fatalf("Incorect Strategy Object for RoundRobinStrategy %v", reflect.TypeOf(s))
	}

	s = newRouterStrategy(LeastConnectionStrategy)
	if reflect.TypeOf(s) != reflect.TypeOf(&leastConnectionServerStrategy{}) {
		t.Fatalf("Incorect Strategy Object for LeastConnectionStrategy %v", reflect.TypeOf(s))
	}

	s = newRouterStrategy(LeastConnectionStrategy)
	if reflect.TypeOf(s) != reflect.TypeOf(&leastConnectionServerStrategy{}) {
		t.Fatalf("Incorect Strategy Object for LeastConnectionStrategy %v", reflect.TypeOf(s))
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	strategy := newRouterStrategy(RoundRobinStrategy)
	r := getRouterForRoundRobinStrategy()
	r.requestCount = 0
	s := strategy.GetAvailableServer(r, "gpt-3.5-turbo")
	if s.Type != server.OpenAiServerType {
		t.Fatalf("Incorrect server returned by Round Robin - %s", s.Type)
	}
	r.requestCount = 1
	s = strategy.GetAvailableServer(r, "gpt-3.5-turbo")
	if s.Type != server.AzureOpenAiServerType {
		t.Fatalf("Incorrect server returned by Round Robin - %s", s.Type)
	}
	r.requestCount = 2
	s = strategy.GetAvailableServer(r, "gpt-3.5-turbo")
	if s.Type != server.OpenAiServerType {
		t.Fatalf("Incorrect server returned by Round Robin - %s", s.Type)
	}
}

func TestLeastActiveConnectionsStrategy(t *testing.T) {
	r := getRouterForActiveConnectionsStrategy()
	strategy := newRouterStrategy(LeastConnectionStrategy)
	s := strategy.GetAvailableServer(r, "gpt-3.5-turbo")
	if s.Type != server.AzureOpenAiServerType {
		t.Fatalf("Incorrect server returned by Least Active Connection Strategy - %s", s.Type)
	}

}

func TestLeastLatencyStrategy(t *testing.T) {
	r := getRouterForLeastLatencyStrategy()
	strategy := newRouterStrategy(LeastLatencyStrategy)
	s := strategy.GetAvailableServer(r, "gpt-3.5-turbo")
	if s.Type != server.OpenAiServerType {
		t.Fatalf("Incorrect server returned by Least Latency Strategy - %s", s.Type)
	}

}

func getRouterForActiveConnectionsStrategy() *Router {
	s1, _ := server.NewRouterServer(
		server.ServerConfig{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		})
	s1.ActiveConnections = 10
	s2, _ := server.NewRouterServer(
		server.ServerConfig{
			Type:            server.AzureOpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		})
	s2.ActiveConnections = 8
	return &Router{servers: []*server.RouterServer{s1, s2}}
}

func getRouterForLeastLatencyStrategy() *Router {
	s1, _ := server.NewRouterServer(
		server.ServerConfig{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		})
	s1.Latency = 100
	s2, _ := server.NewRouterServer(
		server.ServerConfig{
			Type:            server.AzureOpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		})
	s2.Latency = 6000
	return &Router{servers: []*server.RouterServer{s1, s2}}
}

func getRouterForRoundRobinStrategy() *Router {
	router, err := NewRouter([]server.ServerConfig{
		{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		},
		{
			Type:            server.AzureOpenAiServerType,
			Endpoint:        "https://azure-openai.com",
			ApiKey:          "azure-openai-key",
			AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-turbo"},
		},
	}, RoundRobinStrategy)
	if err != nil {
		panic(err)
	}
	return router
}
