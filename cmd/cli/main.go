package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	dotenv "github.com/joho/godotenv"
	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/open-portfolios/codefolio/pkg/llm/openai/chat"
	"github.com/open-portfolios/codefolio/pkg/stdx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var (
	_ llm.Message = stdx.Zero[message]()
)

type message struct {
	role    string
	content string
}

func (m message) Role() string    { return m.role }
func (m message) Content() string { return m.content }

func init() {
	if err := dotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	client := openai.NewClient(
		option.WithBaseURL(os.Getenv("BASE_URL")),
		option.WithAPIKey(os.Getenv("API_KEY")),
	)
	driver := chat.NewCompletionsDriver(&client, chat.WithModel("deepseek-v4-flash"))

	messages := []llm.Message{
		message{llm.RoleSystem, "you're a helpful assistant"},
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if len(strings.TrimSpace(line)) == 0 {
			break
		}
		fmt.Println()
		messages = append(messages, message{llm.RoleUser, line})

		deltaChan, errChan := driver.Stream(ctx, messages)
		var resp strings.Builder
	outer:
		for {
			select {
			case delta, ok := <-deltaChan:
				if !ok {
					break outer
				}
				fmt.Print(delta.Content())
				if _, err := resp.WriteString(delta.Content()); err != nil {
					panic(err)
				}
			case err, ok := <-errChan:
				if ok && err != nil {
					panic(err)
				}
			}
		}
		messages = append(messages, message{llm.RoleAssistant, resp.String()})
		fmt.Println()
		fmt.Println()
	}
}
