package play

type Play struct {
	Name  string
	Title string

	Vars map[string]*Var

	CheckCommand string

	Dependencies []string
	PostInstall  []string

	Sudo bool
	SSH  bool

	Markdown   string
	CodeBlocks []*CodeBlock

	State PlayStateType
}

type Var struct {
	Required       bool
	Default        string
	DefaultFromVar string

	Prompt              PromptType
	PromptMessage       string
	PromptSelectOptions []string
}

type PromptType int

const (
	PromptInput PromptType = iota
	PromptMultiline
	PromptPassword
	PromptConfirm
	PromptSelect
	PromptMultiSelect
	PromptEditor
	Prompt
)

type CodeBlockType int

const (
	CodeBlockBash CodeBlockType = iota
)

type CodeBlock struct {
	Type    CodeBlockType
	Content string
}

type PlayStateType int

const (
	PlayStateReady PlayStateType = iota
	PlayStateUpdated
	PlayStateFailed
)
