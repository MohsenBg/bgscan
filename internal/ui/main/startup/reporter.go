package startup

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

// EnableReportDelay adds a small delay between startup reports.
const EnableReportDelay = true

const (
	minReportDelay = 50 * time.Millisecond
	maxReportDelay = 80 * time.Millisecond
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

	// status is the final non-transient status for this category.
	status categoryStatus
}

func newReporter(categoryID string, ch chan<- tea.Msg, abort *atomic.Bool) *reporter {
	return &reporter{
		categoryID: categoryID,
		ch:         ch,
		abort:      abort,
		status:     catRunning,
	}
}

func (r *reporter) report(status categoryStatus, msg string) {
	r.ch <- logMsg{
		categoryID: r.categoryID,
		status:     status,
		line:       msg,
	}

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
	r.report(catError, formatError(msg, err))
}

func (r *reporter) binaryMissing(component, binary string) {
	r.errMsg(fmt.Sprintf("%s binary %q not found", component, binary), nil)
}

func (r *reporter) critical(msg string, err error) {
	if r.abort != nil {
		r.abort.Store(true)
	}

	r.status = catError

	r.ch <- logMsg{
		categoryID: r.categoryID,
		status:     catError,
		line:       formatError(msg, err),
		critical:   true,
	}
}

func formatError(msg string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: %v", msg, err)
	}

	return msg
}
