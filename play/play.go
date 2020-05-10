package play

type Play struct {
	Name  string
	Title string

	Vars         map[string]*Var
	RegisterVars []string

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
	Name string

	DefaultFromVar string
	Default        string
	Required       bool

	Prompt                PromptType
	PromptForce           bool
	PromptMessage         string
	PromptSelectOptions   []string
	PromptMultiSelectGlue string
	PromptAnswer          string
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
