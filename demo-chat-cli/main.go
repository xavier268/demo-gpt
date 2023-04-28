package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/xavier268/demo-openai/config"
)

const VERBOSE = 0
const model = openai.GPT3Dot5Turbo

var client = config.NewClient()
var conversation = []openai.ChatCompletionMessage{}
var wait = time.Second // time to wait if rate limit reached

func main() {
	fmt.Printf("Chat using %s\nEnter a question - an empty line will stop the conversation\n", model)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		input = strings.Trim(input, " \n\r\t")
		if input == "" {
			fmt.Println("Stop requested")
			break
		}

		resp, err := Ask(input)

		fmt.Printf("%s\n\n", resp)
		if err != nil {
			fmt.Println(err)
			break
		}

	}

	Summary()
	fmt.Println("\nDone.")

}

// Ask a new question, update the messages of the conversation, return the response.
// The context of the conversation is saved.
func Ask(question string) (string, error) {

	mess := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
		Name:    "",
	}
	conversation = append(conversation, mess)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       model,
			Temperature: 0.5,
			Messages:    conversation,
		},
	)

	if err != nil {
		if strings.Contains(err.Error(), "Rate limit reached") { // rate limit exceeded
			fmt.Println("... please wait - rate limit reached, retrying ...")
			time.Sleep(wait)
			wait = wait * 2
			s, err := Ask(question) // try again ...
			return s, err
		}
		return "", err
	}

	wait = time.Second // reset wait to default
	// update conversation
	conversation = append(conversation, resp.Choices[0].Message)

	if VERBOSE >= 1 {
		fmt.Printf("\nTRACE ----------\n%#v\nMESSAGES ------------\n%#v\n\n", resp, conversation)
	}

	return resp.Choices[0].Message.Content, nil

}

// Print summary of conversation.
func Summary() {

	fmt.Println("\nRésumé de la conversation :")
	for i, m := range conversation {
		fmt.Printf("\n%d\t%s %s:\n%s\n", i, m.Role, m.Name, m.Content)
	}
}
