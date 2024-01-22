package server

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	serverConfig := ServerConfig{
		Type: OpenAiServerType,
	}
	_, err := NewRouterServer(serverConfig)
	if err != nil {
		t.Fatalf("Error was not expected %v", err)
	}

	serverConfig = ServerConfig{
		Type: "Foo",
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
	s.postFlight(1)
	if s.ActiveConnections != 9 {
		t.Fatalf("Incorrect Active Connections calculations %d", s.ActiveConnections)
	}
	if s.totalRequests != 21 {
		t.Fatalf("Incorrect Total Requests calculations %d", s.totalRequests)
	}
	if s.totalLatency != 42 {
		t.Fatalf("Incorrect Total Latency calculations %d", s.totalLatency)
	}
	if s.Latency != 2 {
		t.Fatalf("Incorrect Latency calculations %d", s.Latency)
	}
}

func getServer() RouterServer {
	server, _ := NewRouterServer(
		ServerConfig{
			Type: AzureOpenAiServerType,
		},
	)
	return *server
}
