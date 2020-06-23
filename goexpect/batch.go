package expect

import (
	"log"
	"regexp"
	"time"

	"google.golang.org/grpc/codes"
)

// BatchCommands.
const (
	// BatchSend for invoking Send in a batch
	BatchSend = iota
	// BatchExpect for invoking Expect in a batch
	BatchExpect
	// BatchSwitchCase for invoking ExpectSwitchCase in a batch
	BatchSwitchCase
)

// BatchRes returned from ExpectBatch for every Expect command executed.
type BatchRes struct {
	// Idx is used to match the result with the []Batcher commands sent in.
	Idx int
	// Out output buffer for the expect command at Batcher[Idx].
	Output string
	// Match regexp matches for expect command at Batcher[Idx].
	Match []string
}

// Batcher interface is used to make it more straightforward and readable to create
// batches of Expects.
//
// var batch = []Batcher{
//	&BExpT{"password",8},
//	&BSnd{"password\n"},
//	&BExp{"olakar@router>"},
//	&BSnd{ "show interface description\n"},
//	&BExp{ "olakar@router>"},
// }
//
// var batchSwCaseReplace = []Batcher{
//	&BCasT{[]Caser{
//		&BCase{`([0-9]) -- .*\(MASTER\)`, `\1` + "\n"}}, 1},
//	&BExp{`prompt/>`},
// }
type Batcher interface {
	// cmd returns the Batch command.
	Cmd() int
	// Arg returns the command argument.
	Arg() string
	// Timeout returns the timeout duration for the command , <0 gives default value.
	Timeout() time.Duration
	// Cases returns the Caser structure for SwitchCase commands.
	Cases() []Caser
}

// BExp implements the Batcher interface for Expect commands using the default timeout.
type BExp struct {
	// R contains the Expect command regular expression.
	R string
}

// Cmd returns the Expect command (BatchExpect).
func (be *BExp) Cmd() int {
	return BatchExpect
}

// Arg returns the Expect regular expression.
func (be *BExp) Arg() string {
	return be.R
}

// Timeout always returns -1 which sets it to the value used to call the ExpectBatch function.
func (be *BExp) Timeout() time.Duration {
	return -1
}

// Cases always returns nil for the Expect command.
func (be *BExp) Cases() []Caser {
	return nil
}

// BExpT implements the Batcher interface for Expect commands adding a timeout option to the BExp
// type.
type BExpT struct {
	// R contains the Expect command regular expression.
	R string
	// T holds the Expect command timeout in seconds.
	T int
}

// Cmd returns the Expect command (BatchExpect).
func (bt *BExpT) Cmd() int {
	return BatchExpect
}

// Timeout returns the timeout in seconds.
func (bt *BExpT) Timeout() time.Duration {
	return time.Duration(bt.T) * time.Second
}

// Arg returns the Expect regular expression.
func (bt *BExpT) Arg() string {
	return bt.R
}

// Cases always return nil for the Expect command.
func (bt *BExpT) Cases() []Caser {
	return nil
}

// BSnd implements the Batcher interface for Send commands.
type BSnd struct {
	S string
}

// Cmd returns the Send command(BatchSend).
func (bs *BSnd) Cmd() int {
	return BatchSend
}

// Arg returns the data to be sent.
func (bs *BSnd) Arg() string {
	return bs.S
}

// Timeout always returns 0 , Send doesn't have a timeout.
func (bs *BSnd) Timeout() time.Duration {
	return 0
}

// Cases always returns nil , not used for Send commands.
func (bs *BSnd) Cases() []Caser {
	return nil
}

// BCas implements the Batcher interface for SwitchCase commands.
type BCas struct {
	// C holds the Caser array for the SwitchCase command.
	C []Caser
}

// Cmd returns the SwitchCase command(BatchSwitchCase).
func (bc *BCas) Cmd() int {
	return BatchSwitchCase
}

// Arg returns an empty string , not used for SwitchCase.
func (bc *BCas) Arg() string {
	return ""
}

// Timeout returns -1 , setting it to the default value.
func (bc *BCas) Timeout() time.Duration {
	return -1
}

// Cases returns the Caser structure.
func (bc *BCas) Cases() []Caser {
	return bc.C
}

// BCasT implements the Batcher interfacs for SwitchCase commands, adding a timeout option
// to the BCas type.
type BCasT struct {
	// Cs holds the Caser array for the SwitchCase command.
	C []Caser
	// Tout holds the SwitchCase timeout in seconds.
	T int
}

// Timeout returns the timeout in seconds.
func (bct *BCasT) Timeout() time.Duration {
	return time.Duration(bct.T) * time.Second
}

// Cmd returns the SwitchCase command(BatchSwitchCase).
func (bct *BCasT) Cmd() int {
	return BatchSwitchCase
}

// Arg returns an empty string , not used for SwitchCase.
func (bct *BCasT) Arg() string {
	return ""
}

// Cases returns the Caser structure.
func (bct *BCasT) Cases() []Caser {
	return bct.C
}

// Tag represents the state for a Caser.
type Tag int32

const (
	// OKTag marks the desired state was reached.
	OKTag = Tag(iota)
	// FailTag means reaching this state will fail the Switch/Case.
	FailTag
	// ContinueTag will recheck for matches.
	ContinueTag
	// NextTag skips match and continues to the next one.
	NextTag
	// NoTag signals no tag was set for this case.
	NoTag
)

// OK returns the OK Tag and status.
func OK() func() (Tag, *Status) {
	return func() (Tag, *Status) {
		return OKTag, NewStatus(codes.OK, "state reached")
	}
}

// Fail returns Fail Tag and status.
func Fail(s *Status) func() (Tag, *Status) {
	return func() (Tag, *Status) {
		return FailTag, s
	}
}

// Continue returns the Continue Tag and status.
func Continue(s *Status) func() (Tag, *Status) {
	return func() (Tag, *Status) {
		return ContinueTag, s
	}
}

// Next returns the Next Tag and status.
func Next() func() (Tag, *Status) {
	return func() (Tag, *Status) {
		return NextTag, NewStatus(codes.Unimplemented, "Next returns not implemented")
	}
}

// LogContinue logs the message and returns the Continue Tag and status.
func LogContinue(msg string, s *Status) func() (Tag, *Status) {
	return func() (Tag, *Status) {
		log.Print(msg)
		return ContinueTag, s
	}
}

// Caser is an interface for ExpectSwitchCase and Batch to be able to handle
// both the Case struct and the more script friendly BCase struct.
type Caser interface {
	// RE returns a compiled regexp
	RE() (*regexp.Regexp, error)
	// Send returns the send string
	String() string
	// Tag returns the Tag.
	Tag() (Tag, *Status)
	// Retry returns true if there are retries left.
	Retry() bool
}

// Case used by the ExpectSwitchCase to take different Cases.
// Implements the Caser interface.
type Case struct {
	// R is the compiled regexp to match.
	R *regexp.Regexp
	// S is the string to send if Regexp matches.
	S string
	// T is the Tag for this Case.
	T func() (Tag, *Status)
	// Rt specifies number of times to retry, only used for cases tagged with Continue.
	Rt int
}

// Tag returns the tag for this case.
func (c *Case) Tag() (Tag, *Status) {
	if c.T == nil {
		return NoTag, NewStatus(codes.OK, "no Tag set")
	}
	return c.T()
}

// RE returns the compiled regular expression.
func (c *Case) RE() (*regexp.Regexp, error) {
	return c.R, nil
}

// Retry decrements the Retry counter and checks if there are any retries left.
func (c *Case) Retry() bool {
	defer func() { c.Rt-- }()
	return c.Rt > 0
}

// Send returns the string to send if regexp matches
func (c *Case) String() string {
	return c.S
}

// BCase with just a string is a bit more friendly to scripting.
// Implements the Caser interface.
type BCase struct {
	// R contains the string regular expression.
	R string
	// S contains the string to be sent if R matches.
	S string
	// T contains the Tag.
	T func() (Tag, *Status)
	// Rt contains the number of retries.
	Rt int
}

// RE returns the compiled regular expression.
func (b *BCase) RE() (*regexp.Regexp, error) {
	if b.R == "" {
		return nil, nil
	}
	return regexp.Compile(b.R)
}

// Send returns the string to send.
func (b *BCase) String() string {
	return b.S
}

// Tag returns the BCase Tag.
func (b *BCase) Tag() (Tag, *Status) {
	if b.T == nil {
		return NoTag, NewStatus(codes.OK, "no Tag set")
	}
	return b.T()
}

// Retry decrements the Retry counter and checks if there are any retries left.
func (b *BCase) Retry() bool {
	b.Rt--
	return b.Rt > -1
}
