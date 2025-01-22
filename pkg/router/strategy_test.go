package router

import (
	"reflect"
	"slices"
	"testing"

	"github.com/acai-travel/go-openai-router/v2/pkg/server"
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

func TestRoundRobinWithModelAvailability(t *testing.T) {
	r := getRouterWithModelVariation()
	strategy := newRouterStrategy(RoundRobinStrategy)

	modelName := "gpt-4-vision-preview"
	s := strategy.GetAvailableServer(r, modelName)
	if s.Type != server.OpenAiServerType || !slices.Contains(s.AvailableModels, modelName) {
		t.Fatalf("Round Robin with Model Availability failed to select correct server for model %s", modelName)
	}

	modelName = "model-not-available"
	s = strategy.GetAvailableServer(r, modelName)
	if s != nil {
		t.Fatal("Round Robin with Model Availability selected a server for an unavailable model")
	}
}

func TestLoadBalancingWithModelAvailability(t *testing.T) {
	r := getRouterForModelBalancingTest()
	strategy := newRouterStrategy(LeastConnectionStrategy)
	modelName := "gpt-4-vision-preview"

	// Make several requests to simulate load balancing
	for i := 0; i < 6; i++ {
		s := strategy.GetAvailableServer(r, modelName)
		if s == nil {
			t.Fatal("Expected to select a server but got nil")
		}
		if !slices.Contains(s.AvailableModels, modelName) {
			t.Fatalf("Selected server does not support the requested model %s", modelName)
		}

		// Increment active connections to simulate load
		s.ActiveConnections++

		// Ensure the third server is never selected
		if s.Type == server.AzureOpenAiServerType {
			t.Fatal("Selected server should not have been chosen as it does not support the model")
		}
	}

	// After load balancing, check the distribution of active connections
	if r.servers[0].ActiveConnections == r.servers[1].ActiveConnections {
		t.Log("Load balancing between servers supporting the model confirmed")
	} else {
		t.Fatal("Expected even distribution of load between servers supporting the model")
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

func getRouterWithModelVariation() *Router {
	serverConfigs := []server.ServerConfig{
		{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com",
			ApiKey:          "openai-key",
			AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-vision-preview"},
		},
		{
			Type:            server.AzureOpenAiServerType,
			Endpoint:        "https://azure-openai.com",
			ApiKey:          "azure-openai-key",
			AvailableModels: []string{"gpt-3.5-turbo"},
		},
	}
	router, err := NewRouter(serverConfigs, RoundRobinStrategy)
	if err != nil {
		panic(err)
	}
	return router
}

func getRouterForModelBalancingTest() *Router {
	serverConfigs := []server.ServerConfig{
		{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com/1",
			ApiKey:          "key1",
			AvailableModels: []string{"gpt-4-vision-preview"},
		},
		{
			Type:            server.OpenAiServerType,
			Endpoint:        "https://api.openai.com/2",
			ApiKey:          "key2",
			AvailableModels: []string{"gpt-4-vision-preview"},
		},
		{
			Type:            server.AzureOpenAiServerType,
			Endpoint:        "https://api.openai.com/3",
			ApiKey:          "key3",
			AvailableModels: []string{"gpt-3.5-turbo"},
		},
	}
	router, err := NewRouter(serverConfigs, LeastConnectionStrategy)
	if err != nil {
		panic(err)
	}
	return router
}
