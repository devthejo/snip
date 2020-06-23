// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package expect is a Go version of the classic TCL Expect.
package expect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/goterm/term"
	"google.golang.org/grpc/codes"
)

// DefaultTimeout is the default Expect timeout.
const DefaultTimeout = 60 * time.Second

const (
	bufferSize    = 8192            // bufferSize sets the size of the io buffers.
	checkDuration = 2 * time.Second // checkDuration how often to check for new output.
)

// Expecter interface primarily to make testing easier.
type Expecter interface {
	// Expect reads output from a spawned session and tries matching it with the provided regular expression.
	// It returns  all output found until match.
	Expect(*regexp.Regexp, time.Duration) (string, []string, error)
	// ExpectBatch takes an array of BatchEntries and runs through them in order. For every Expect
	// command a BatchRes entry is created with output buffer and sub matches.
	// Failure of any of the batch commands will stop the execution, returning the results up to the
	// failure.
	ExpectBatch([]Batcher, time.Duration) ([]BatchRes, error)
	// ExpectSwitchCase makes it possible to Expect with multiple regular expressions and actions. Returns the
	// full output and submatches of the commands together with an index for the matching Case.
	ExpectSwitchCase([]Caser, time.Duration) (string, []string, int, error)
	// Send sends data into the spawned session.
	Send(string) error
	// Close closes the spawned session and files.
	Close() error
}

// GExpect implements the Expecter interface.
type GExpect struct {
	readout func(p []byte) (n int, err error)
	writein func(p []byte) (n int, err error)
	clean   func()
	wait    func() error
	// snd is the channel used by the Send command to send data into the spawned command.
	snd chan string
	// rcv is used to signal the Expect commands that new data arrived.
	rcv chan struct{}
	// chkMu lock protecting the check function.
	chkMu sync.RWMutex
	// chk contains the function to check if the spawned command is alive.
	chk func() bool
	// cls contains the function to close spawned command.
	cls func() error
	// timeout contains the default timeout for a spawned command.
	timeout time.Duration
	// sendTimeout contains the default timeout for a send command.
	sendTimeout time.Duration
	// chkDuration contains the duration between checks for new incoming data.
	chkDuration time.Duration
	// verbose enables verbose logging.
	verbose bool
	// verboseWriter if set specifies where to write verbose information.
	verboseWriter io.Writer
	// teeWriter receives a duplicate of the spawned process's output when set.
	teeWriter io.WriteCloser
	// PartialMatch enables the returning of unmatched buffer so that consecutive expect call works.
	partialMatch bool

	// mu protects the output buffer. It must be held for any operations on out.
	mu  sync.Mutex
	out bytes.Buffer
}

// String implements the stringer interface.
func (e *GExpect) String() string {
	res := fmt.Sprintf("%p: ", e)
	res += fmt.Sprintf("buf: %q", e.out.String())
	return res
}

// ExpectBatch takes an array of BatchEntry and executes them in order filling in the BatchRes
// array for any Expect command executed.
func (e *GExpect) ExpectBatch(batch []Batcher, timeout time.Duration) ([]BatchRes, error) {
	res := []BatchRes{}
	for i, b := range batch {
		switch b.Cmd() {
		case BatchExpect:
			re, err := regexp.Compile(b.Arg())
			if err != nil {
				return res, err
			}
			to := b.Timeout()
			if to < 0 {
				to = timeout
			}
			out, match, err := e.Expect(re, to)
			res = append(res, BatchRes{i, out, match})
			if err != nil {
				return res, err
			}
		case BatchSend:
			if err := e.Send(b.Arg()); err != nil {
				return res, err
			}
		case BatchSwitchCase:
			to := b.Timeout()
			if to < 0 {
				to = timeout
			}
			out, match, _, err := e.ExpectSwitchCase(b.Cases(), to)
			res = append(res, BatchRes{i, out, match})
			if err != nil {
				return res, err
			}
		default:
			return res, errors.New("unknown command:" + strconv.Itoa(b.Cmd()))
		}
	}
	return res, nil
}

func (e *GExpect) check() bool {
	e.chkMu.RLock()
	defer e.chkMu.RUnlock()
	return e.chk()
}

// ExpectSwitchCase checks each Case against the accumulated out buffer, sending specified
// string back. Leaving Send empty will Send nothing to the process.
// Substring expansion can be used eg.
// 	Case{`vf[0-9]{2}.[a-z]{3}[0-9]{2}\.net).*UP`,`show arp \1`}
// 	Given: vf11.hnd01.net            UP      35 (4)        34 (4)          CONNECTED         0              0/0
// 	Would send: show arp vf11.hnd01.net
func (e *GExpect) ExpectSwitchCase(cs []Caser, timeout time.Duration) (string, []string, int, error) {
	// Compile all regexps
	rs := make([]*regexp.Regexp, 0, len(cs))
	for _, c := range cs {
		re, err := c.RE()
		if err != nil {
			return "", []string{""}, -1, err
		}
		rs = append(rs, re)
	}
	// Setup timeouts
	// timeout == 0 => Just dump the buffer and exit.
	// timeout < 0  => Set default value.
	if timeout < 0 {
		timeout = e.timeout
	}
	timer := time.NewTimer(timeout)
	check := e.chkDuration
	// Check if any new data arrived every checkDuration interval.
	// If timeout/4 is less than the checkout interval we set the checkout to
	// timeout/4. If timeout ends up being 0 we bump it to one to keep the Ticker from
	// panicking.
	// All this b/c of the unreliable channel send setup in the read function,making it
	// possible for Expect* functions to miss the rcv signal.
	//
	// from read():
	//		// Ping Expect function
	//		select {
	//		case e.rcv <- struct{}{}:
	//		default:
	//		}
	//
	// A signal is only sent if any Expect function is running. Expect could miss it
	// while playing around with buffers and matching regular expressions.
	if timeout>>2 < check {
		check = timeout >> 2
		if check <= 0 {
			check = 1
		}
	}
	chTicker := time.NewTicker(check)
	defer chTicker.Stop()
	// Read in current data and start actively check for matches.
	var tbuf bytes.Buffer
	if _, err := io.Copy(&tbuf, e); err != nil {
		return tbuf.String(), nil, -1, fmt.Errorf("io.Copy failed: %v", err)
	}
	for {
	L1:
		for i, c := range cs {
			if rs[i] == nil {
				continue
			}
			match := rs[i].FindStringSubmatch(tbuf.String())
			if match == nil {
				continue
			}

			t, s := c.Tag()
			if t == NextTag && !c.Retry() {
				continue
			}

			if e.verbose {
				if e.verboseWriter != nil {
					vStr := fmt.Sprintln(term.Green("Match for RE:").String() + fmt.Sprintf(" %q found: %q Buffer: %s", rs[i].String(), match, tbuf.String()))
					for n, bytesRead, err := 0, 0, error(nil); bytesRead < len(vStr); bytesRead += n {
						n, err = e.verboseWriter.Write([]byte(vStr)[n:])
						if err != nil {
							log.Printf("Write to Verbose Writer failed: %v", err)
							break
						}
					}
				} else {
					log.Printf("Match for RE: %q found: %q Buffer: %q", rs[i].String(), match, tbuf.String())
				}
			}

			tbufString := tbuf.String()
			o := tbufString

			if e.partialMatch {
				// Return the part of the buffer that is not matched by the regular expression so that the next expect call will be able to match it.
				matchIndex := rs[i].FindStringIndex(tbufString)
				o = tbufString[0:matchIndex[1]]
				e.returnUnmatchedSuffix(tbufString[matchIndex[1]:])
			}

			tbuf.Reset()

			st := c.String()
			// Replace the submatches \[0-9]+ in the send string.
			if len(match) > 1 && len(st) > 0 {
				for i := 1; i < len(match); i++ {
					// \(submatch) will be expanded in the Send string.
					// To escape use \\(number).
					si := strconv.Itoa(i)
					r := strings.NewReplacer(`\\`+si, `\`+si, `\`+si, `\\`+si)
					st = r.Replace(st)
					st = strings.Replace(st, `\\`+si, match[i], -1)
				}
			}
			// Don't send anything if string is empty.
			if st != "" {
				if err := e.Send(st); err != nil {
					return o, match, i, fmt.Errorf("failed to send: %q err: %v", st, err)
				}
			}
			// Tag handling.
			switch t {
			case OKTag, FailTag, NoTag:
				return o, match, i, s.Err()
			case ContinueTag:
				if !c.Retry() {
					return o, match, i, s.Err()
				}
				break L1
			case NextTag:
				break L1
			default:
				s = NewStatusf(codes.Unknown, "Tag: %d unknown, err: %v", t, s)
			}
			return o, match, i, s.Err()
		}
		if !e.check() {
			nr, err := io.Copy(&tbuf, e)
			if err != nil {
				return tbuf.String(), nil, -1, fmt.Errorf("io.Copy failed: %v", err)
			}
			if nr == 0 {
				return tbuf.String(), nil, -1, errors.New("expect: Process not running")
			}
		}
		select {
		case <-timer.C:
			// Expect timeout.
			nr, err := io.Copy(&tbuf, e)
			if err != nil {
				return tbuf.String(), nil, -1, fmt.Errorf("io.Copy failed: %v", err)
			}
			// If we got no new data we return otherwise give it another chance to match.
			if nr == 0 {
				return tbuf.String(), nil, -1, TimeoutError(timeout)
			}
			timer = time.NewTimer(timeout)
		case <-chTicker.C:
			// Periodical timer to make sure data is handled in case the <-e.rcv channel
			// was missed.
			if _, err := io.Copy(&tbuf, e); err != nil {
				return tbuf.String(), nil, -1, fmt.Errorf("io.Copy failed: %v", err)
			}
		case <-e.rcv:
			// Data to fetch.
			if _, err := io.Copy(&tbuf, e); err != nil {
				return tbuf.String(), nil, -1, fmt.Errorf("io.Copy failed: %v", err)
			}
		}
	}
}

func Spawn(opt *SpawnOptions) (*GExpect, <-chan error, error) {
	switch {
	case opt == nil:
		return nil, nil, errors.New("SpawnOptions is <nil>")
	case opt.In == nil:
		return nil, nil, errors.New("In can't be <nil>")
	case opt.Out == nil:
		return nil, nil, errors.New("Out can't be <nil>")
	case opt.Wait == nil:
		return nil, nil, errors.New("Wait can't be <nil>")
	case opt.Close == nil:
		return nil, nil, errors.New("Close can't be <nil>")
	case opt.Check == nil:
		return nil, nil, errors.New("Check can't be <nil>")
	}

	timeout := opt.Timeout
	if timeout < 1 {
		timeout = DefaultTimeout
	}

	res := make(chan error, 1)

	stdin := opt.In
	writein := func(p []byte) (n int, err error) {
		return stdin.Write(p)
	}

	stdout := opt.Out
	readout := func(p []byte) (n int, err error) {
		return stdout.Read(p)
	}

	var chkDuration time.Duration
	if opt.CheckDuration > 0 {
		chkDuration = opt.CheckDuration
	} else {
		chkDuration = checkDuration
	}

	e := &GExpect{
		readout:       readout,
		writein:       writein,
		rcv:           make(chan struct{}),
		snd:           make(chan string),
		timeout:       timeout,
		chkDuration:   chkDuration,
		cls:           opt.Close,
		chk:           opt.Check,
		wait:          opt.Wait,
		clean:         opt.Clean,
		sendTimeout:   opt.SendTimeout,
		verbose:       opt.Verbose,
		verboseWriter: opt.VerboseWriter,
		teeWriter:     opt.Tee,
		partialMatch:  opt.PartialMatch,
	}

	go func() {
		// Moving the go read/write functions here makes sure the command is started before first checking if it's running.
		clean := make(chan struct{})
		chDone := e.goIO(clean)
		// Signal command started
		res <- nil
		cErr := e.wait()
		close(chDone)
		// make sure the read/send routines are done before closing the pty.
		<-clean
		res <- cErr
	}()

	return e, res, <-res

}

// Close closes the Spawned session.
func (e *GExpect) Close() error {
	return e.cls()
}

// Change the check function using lock
func (e *GExpect) ChangeCheck(f func() bool) {
	e.chkMu.Lock()
	e.chk = f
	e.chkMu.Unlock()
}

// Read implements the reader interface for the out buffer.
func (e *GExpect) Read(p []byte) (nr int, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.out.Read(p)
}

func (e *GExpect) returnUnmatchedSuffix(p string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	newBuffer := bytes.NewBufferString(p)
	newBuffer.WriteString(e.out.String())
	e.out = *newBuffer
}

// Send sends a string to spawned process.
func (e *GExpect) Send(in string) error {
	if !e.check() {
		return errors.New("expect: Process not running")
	}

	if e.sendTimeout == 0 {
		e.snd <- in
	} else {
		select {
		case <-time.After(e.sendTimeout):
			return fmt.Errorf("send to spawned process command reached the timeout %v", e.sendTimeout)
		case e.snd <- in:
		}
	}

	if e.verbose {
		if e.verboseWriter != nil {
			vStr := fmt.Sprintln(term.Blue("Sent:").String() + fmt.Sprintf(" %q", in))
			_, err := e.verboseWriter.Write([]byte(vStr))
			if err != nil {
				log.Printf("Write to Verbose Writer failed: %v", err)
			}
		}
		log.Printf("Sent: %q", in)
	}

	return nil
}

// goIO starts the io handlers.
func (e *GExpect) goIO(clean chan struct{}) (done chan struct{}) {
	done = make(chan struct{})
	var ptySync sync.WaitGroup
	ptySync.Add(2)
	go e.read(done, &ptySync)
	go e.send(done, &ptySync)
	go func() {
		ptySync.Wait()
		if e.clean != nil {
			e.clean()
		}
		close(clean)
	}()
	return done
}

// Expect reads spawned processes output looking for input regular expression.
// Timeout set to 0 makes Expect return the current buffer.
// Negative timeout value sets it to Default timeout.
func (e *GExpect) Expect(re *regexp.Regexp, timeout time.Duration) (string, []string, error) {
	out, match, _, err := e.ExpectSwitchCase([]Caser{&Case{re, "", nil, 0}}, timeout)
	return out, match, err
}

// read reads from the PTY master and forwards to active Expect function.
func (e *GExpect) read(done chan struct{}, ptySync *sync.WaitGroup) {
	defer ptySync.Done()
	buf := make([]byte, bufferSize)
	for {
		nr, err := e.readout(buf)
		if err != nil && !e.check() {
			if e.teeWriter != nil {
				e.teeWriter.Close()
			}
			if err == io.EOF {
				if e.verbose {
					log.Printf("read closing down: %v", err)
				}
				return
			}
			return
		}
		// Tee output to writer
		if e.teeWriter != nil {
			e.teeWriter.Write(buf[:nr])
		}
		// Add to buffer
		e.mu.Lock()
		e.out.Write(buf[:nr])
		e.mu.Unlock()
		// Ping Expect function
		select {
		case e.rcv <- struct{}{}:
		default:
		}
	}
}

// send writes to the PTY master.
func (e *GExpect) send(done chan struct{}, ptySync *sync.WaitGroup) {
	defer ptySync.Done()
	for {
		select {
		case <-done:
			return
		case sstr, ok := <-e.snd:
			if !ok {
				return
			}
			if _, err := e.writein([]byte(sstr)); err != nil || !e.check() {
				log.Printf("send failed: %v", err)
				break
			}
		}
	}
}
