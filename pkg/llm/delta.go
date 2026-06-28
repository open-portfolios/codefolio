package llm

type Delta interface {
	Role() string
	Content() string
	Usage() int64
}
