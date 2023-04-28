// Package config accesses a securely provided api key to provides an openai Client.
package config

import "github.com/sashabaranov/go-openai"

// This value should be initialized before running the demo.
// NEVER commit the file doing this initialization !
var api_secret_key string

func NewClient() *openai.Client {
	if api_secret_key == "" {
		panic("the API SECRET KEY is not available")
	}
	return openai.NewClient(api_secret_key)
}
