# Go OpenAI Router 

## What is it?
A simple routing layer for using OpenAI/AzureOpenAI endpoints in a go project in production. Wraps the official Azure OpenAI SDK for go - https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai

## When should you use it?
* When a single Azure/OpenAI API Endpoint doesn't have enough tokens/minute bandwidth for your use case and you need a router/load-balancer to manage requests between different Azure/OpenAI Endpoints.
* Like us, you are building AI/GenAI products using Golang instead of Python.

## Usage
The router provides three different load balancing strategies -
1. Round Robin
2. Least Busy
3. Least Latency

The router expects that `<DEPLOYMENT_NAME>` exists in all the underlying servers that the router uses.

### Example -
```golang

//Create server configuration - 1
config1 := server.ServerConfig{
    Endpoint: "https://<YOUR_AZURE_RESOURCE>.openai.azure.com/",
    ApiKey:   "<YOUR_AZURE_KEY>",
    Type:     server.AzureOpenAiServerType,
}
//Create server configuration - 2
config2 := server.ServerConfig{
    Endpoint: "https://api.openai.com/v1",
    ApiKey:   "<YOUR_OPENAI_KEY",
    Type:     server.OpenAiServerType,
}

//Create the router using the conifgurations and a strategy
router, _ := router.NewRouter(
    []server.ServerConfig{config1, config2}, 
    router.LeastLatencyStrategy,
)

//Prepare a sample azure request
azureMessages := []azopenai.ChatRequestMessageClassification{}
azureMessages = append(azureMessages, &azopenai.ChatRequestUserMessage{
    Content: azopenai.NewChatRequestUserMessageContent("Who wrote the Jungle Book?"),
})
body := azopenai.ChatCompletionsOptions{
		Messages:       azureMessages,
		DeploymentName: "<DEPLOYMENT_NAME>",
}

//Make the call using the router
router.GetChatCompletions(context.TODO(), body, nil)
```


## Contribution
We decided to build and open-source this project since we believe this is a key challenge people will face when they want to deploy their GenAI products in production to large enterprises/userbases and since we didn't find a suitable alternative in Golang for utilities that exist for python, for example - https://github.com/BerriAI/litellm

If you are interested in contributing to the project, feel free to open an issue, request or a PR.

### Roadmap
Currently we don't have a structured/formalized roadmap for the project, we will be adding features as we need them. But some of the things that we believe would happens soon are - 
1. Routing based on token usage/requests per minute.
2. Rate limiting/throttling.




