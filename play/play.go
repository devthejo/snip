package play

type Play struct {
	Name     string
	Title    string
	FuncName string

	Vars         map[string]*Var
	RegisterVars []string

	CheckCommand string

	Dependencies []string
	PostInstall  []string

	Sudo bool
	SSH  bool

	Markdown         string
	SourceCodeBlocks []*CodeBlock
	CodeBlocks       []*CodeBlock

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
