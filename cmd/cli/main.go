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

type ConsoleVisitor struct {
	llm.BaseDeltaVisitor

	tokens  uint64
	builder strings.Builder
}

func (c *ConsoleVisitor) VisitMessage(m llm.MessageDelta) error {
	c.builder.WriteString(m.Content)
	fmt.Print(m.Content)
	return nil
}

func (c *ConsoleVisitor) VisitUsage(u llm.UsageDelta) error {
	c.tokens += u.TotalTokens
	return nil
}

func (c *ConsoleVisitor) Summarize() (message string) {
	fmt.Printf("\n\nused %d tokens\n", c.tokens)
	message = c.builder.String()
	c.tokens = 0
	c.builder.Reset()
	return
}

func main() {
	ctx := context.Background()
	client := openai.NewClient(
		option.WithBaseURL(os.Getenv("BASE_URL")),
		option.WithAPIKey(os.Getenv("API_KEY")),
	)
	driver := chat.NewCompletionsDriver(&client)

	messages := []llm.Message{
		message{llm.RoleSystem, "you're a helpful assistant"},
	}

	reader := bufio.NewReader(os.Stdin)
	visitor := new(ConsoleVisitor)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if len(strings.TrimSpace(line)) == 0 {
			break
		}
		messages = append(messages, message{llm.RoleUser, line})

		fmt.Println()
		deltaChan, errChan := driver.Stream(ctx, messages, llm.WithModel("deepseek-v4-flash"))
	outer:
		for {
			select {
			case delta, ok := <-deltaChan:
				if !ok {
					break outer
				}
				if err := delta.Accept(visitor); err != nil {
					panic(err)
				}
			case err, ok := <-errChan:
				if ok && err != nil {
					panic(err)
				}
			}
		}
		messages = append(messages, message{llm.RoleAssistant, visitor.Summarize()})
		fmt.Println()
	}
}
