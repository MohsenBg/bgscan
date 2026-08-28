package startup

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

const EnableReportDelay = false

const (
	minReportDelay = 200 * time.Millisecond
	maxReportDelay = 300 * time.Millisecond
)

type logMsg struct {
	categoryID string
	status     categoryStatus
	line       string
	critical   bool
}

type allChecksDoneMsg struct{}

type reporter struct {
	categoryID string
	ch         chan<- tea.Msg
	abort      *atomic.Bool

	// status tracks the final (non-transient) status this category should
	// end with consumed via categoryEndMsg once the check function returns.
	status categoryStatus
}

func newReporter(categoryID string, ch chan<- tea.Msg, abort *atomic.Bool) *reporter {
	return &reporter{categoryID: categoryID, ch: ch, abort: abort, status: catRunning}
}

func (r *reporter) report(status categoryStatus, msg string) {
	r.ch <- logMsg{categoryID: r.categoryID, status: status, line: msg}

	if status != catRunning || r.status == catRunning {
		r.status = status
	}

	if EnableReportDelay {
		time.Sleep(randomReportDelay())
	}
}

func randomReportDelay() time.Duration {
	return minReportDelay + time.Duration(rand.Int64N(int64(maxReportDelay-minReportDelay)))
}

func (r *reporter) info(msg string)             { r.report(catRunning, msg) }
func (r *reporter) wait(msg string)             { r.report(catWait, msg) }
func (r *reporter) infof(f string, a ...any)    { r.info(fmt.Sprintf(f, a...)) }
func (r *reporter) success(msg string)          { r.report(catOK, msg) }
func (r *reporter) successf(f string, a ...any) { r.success(fmt.Sprintf(f, a...)) }
func (r *reporter) warn(msg string)             { r.report(catWarn, msg) }
func (r *reporter) warnf(f string, a ...any)    { r.warn(fmt.Sprintf(f, a...)) }

func (r *reporter) errMsg(msg string, err error) {
	if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}
	r.report(catError, msg)
}

func (r *reporter) binaryMissing(component, binary string) {
	r.errMsg(fmt.Sprintf("%s binary %q not found", component, binary), nil)
}

func (r *reporter) critical(msg string, err error) {
	if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}
	if r.abort != nil {
		r.abort.Store(true)
	}
	r.status = catError
	r.ch <- logMsg{categoryID: r.categoryID, status: catError, line: msg, critical: true}
}
