package server

import (
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	serverConfig := ServerConfig{
		Type:            OpenAiServerType,
		Endpoint:        "https://azure-openai.com",
		ApiKey:          "azure-openai-key",
		AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-turbo"},
	}
	_, err := NewRouterServer(serverConfig)
	if err != nil {
		t.Fatalf("Error was not expected %v", err)
	}

	serverConfig = ServerConfig{
		Type:            "Foo",
		Endpoint:        "https://azure-openai.com",
		ApiKey:          "azure-openai-key",
		AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-turbo"},
	}
	_, err = NewRouterServer(serverConfig)
	if err == nil {
		t.Fatalf("Error was expected for this Server Type %v", err)
	}
}

func TestPreFlight(t *testing.T) {
	s := getServer()
	s.preFlight()
	if s.ActiveConnections != 1 {
		t.Fatalf("Incorrect Active Connections calculations %d", s.ActiveConnections)
	}

}

func TestPostFlight(t *testing.T) {
	s := getServer()
	s.ActiveConnections = 10
	s.totalRequests = 20
	s.totalLatency = 41
	start := time.Now().Add(-15 * time.Second)
	s.postFlight(start)
	if s.ActiveConnections != 9 {
		t.Fatalf("Incorrect Active Connections calculations %d", s.ActiveConnections)
	}
	if s.totalRequests != 21 {
		t.Fatalf("Incorrect Total Requests calculations %d", s.totalRequests)
	}
	if s.totalLatency <= 41 {
		t.Fatalf("Incorrect Total Latency calculations %d", s.totalLatency)
	}
	if s.Latency <= 2 {
		t.Fatalf("Incorrect Latency calculations %d", s.Latency)
	}
}

func getServer() RouterServer {
	server, _ := NewRouterServer(
		ServerConfig{
			Type:            AzureOpenAiServerType,
			Endpoint:        "https://azure-openai.com",
			ApiKey:          "azure-openai-key",
			AvailableModels: []string{"gpt-3.5-turbo", "gpt-4-turbo"},
		},
	)
	return *server
}
