package expect

import (
	"io"
	"time"
)

type SpawnOptions struct {
	Timeout time.Duration
	// In is where Expect Send messages will be written.
	In io.WriteCloser
	// Out will be read and matched by the expecter.
	Out io.Reader
	// Wait is used by expect to know when the session is over and cleanup of io Go routines should happen.
	Wait func() error
	// Close will be called as part of the expect Close, should normally include a Close of the In WriteCloser.
	Close func() error
	// Check is called everytime a Send or Expect function is called to makes sure the session is still running.
	Check func() bool
	Clean func()

	// CheckDuration changes the default duration checking for new incoming data.
	CheckDuration time.Duration

	SendTimeout   time.Duration
	Verbose       bool
	VerboseWriter io.Writer
	Tee           io.WriteCloser
	PartialMatch  bool
}
