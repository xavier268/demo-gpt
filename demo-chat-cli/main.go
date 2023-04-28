package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
	fmt.Printf("Chat using %s\n", model)

	Learn(
		"Mon nom est Xavier",
		"Je veux que tu utilises mon nom le plus possible dans toutes tes réponses, et que tu ne me parle qu'en français.",
		"Par exemple, au lieu de dire simplement oui, tu repondras oui, Xavier.",
		"Mon langage de programation préféré est le go (golang). Si je te demande un exemple de code sans préciser, c'est que je veux du code en go.",
		"Fais toujours des réponses courtes et factuelles. Si tu as besoin de plus d'information, pose des questions.",
		"Je vais de donner un certain nombre de dates à retenir précisément ",
		"La date de l'anniversaire de mon chien, le 8 mars",
		"La date d'achat de ma voiture, le 15 février 1997",
		"Rappelle toi bien de ces deux dates",
		"Aujourd'hui, nous sommes le  "+time.Now().Format("2 janvier 2006"),
	)

	fmt.Println("Enter a question - an empty line will stop the conversation")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		input = strings.Trim(input, " \n\r\t")
		if input == "" {
			fmt.Println("Stop requested")
			break
		}

		resp, u, err := Ask(input)

		fmt.Printf("%s\n(%d tokens)\n\n", resp, u)
		if err != nil {
			fmt.Println(err)
			break
		}

	}

	Summary()
	fname := SaveToFile()
	fmt.Printf("\nSaved conversation to %s\nDone.", fname)

}

// Preload the conversation with some facts and assertions.
// This allows ChatGPT to "learn" a couple of facts and set some preferences.
// From the doc, it is better to use UserRole than SytemRole.
func Learn(facts ...string) {
	for _, f := range facts {
		conversation = append(conversation,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: f,
			},
		)
	}

	// add empty line for printing
	conversation = append(conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "",
		Name:    "",
	})
}

// Ask a new question, update the messages of the conversation, return the response.
// The context of the conversation is saved.
// return reponse, nb of tokens, and error.
func Ask(question string) (string, int, error) {

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
			Temperature: 0.2,
			Messages:    conversation,
		},
	)

	if err != nil {
		if strings.Contains(err.Error(), "Rate limit reached") { // rate limit exceeded
			fmt.Println("... please wait - rate limit reached, retrying ...")
			time.Sleep(wait)
			wait = wait * 2
			s, u, err := Ask(question) // try again ...
			return s, u, err
		}
		return "", 0, err
	}

	wait = time.Second // reset wait to default
	// update conversation
	conversation = append(conversation, resp.Choices[0].Message)

	if VERBOSE >= 1 {
		fmt.Printf("\nTRACE ----------\n%#v\nMESSAGES ------------\n%#v\n\n", resp, conversation)
	}

	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil

}

// Print summary of conversation.
func Summary() {

	fmt.Println("\nRésumé de la conversation :")
	for i, m := range conversation {
		fmt.Printf("\n%d\t%s %s:\n%s\n", i, m.Role, m.Name, m.Content)
	}
}

func Save(out io.Writer) {
	for _, m := range conversation {
		fmt.Fprintln(out, m.Content)
	}
}

func SaveToFile() string {
	fname := time.Now().Format("conv-2006-01-02.150405.txt")
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	Save(f)
	return fname
}
