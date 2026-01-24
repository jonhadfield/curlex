package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Progress displays a modern progress indicator during test execution
type Progress struct {
	writer     io.Writer
	total      int
	current    int
	mu         sync.Mutex
	active     bool
	done       chan bool
	isTerminal bool
	noColor    bool
}

// NewProgress creates a new progress indicator
func NewProgress(total int, noColor bool) *Progress {
	return &Progress{
		writer:     os.Stderr,
		total:      total,
		current:    0,
		done:       make(chan bool),
		isTerminal: term.IsTerminal(int(os.Stderr.Fd())),
		noColor:    noColor,
	}
}

// Start begins showing the progress indicator
func (p *Progress) Start() {
	if !p.isTerminal {
		return // Don't show progress in non-terminal environments
	}

	p.mu.Lock()
	p.active = true
	p.mu.Unlock()

	go p.render()
}

// Increment updates progress by one test
func (p *Progress) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
}

// Stop halts and clears the progress indicator
func (p *Progress) Stop() {
	p.mu.Lock()
	if !p.active {
		p.mu.Unlock()
		return
	}
	p.active = false
	p.mu.Unlock()

	p.done <- true
	p.clear()
}

// render continuously updates the progress display
func (p *Progress) render() {
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.mu.Lock()
			if !p.active {
				p.mu.Unlock()
				return
			}

			current := p.current
			total := p.total
			p.mu.Unlock()

			// Choose display based on whether we have a known total
			var display string
			if total > 0 {
				display = p.formatProgressBar(current, total, spinnerFrames[frameIdx])
			} else {
				display = p.formatSpinner(current, spinnerFrames[frameIdx])
			}

			// Print with carriage return to overwrite
			fmt.Fprintf(p.writer, "\r%s", display)

			frameIdx = (frameIdx + 1) % len(spinnerFrames)
		}
	}
}

// formatProgressBar creates a progress bar display
func (p *Progress) formatProgressBar(current, total int, spinner string) string {
	percentage := float64(current) / float64(total) * 100
	barWidth := 30
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	if p.noColor {
		return fmt.Sprintf("%s Running tests [%s] %d/%d (%.0f%%)",
			spinner, bar, current, total, percentage)
	}

	// Color the spinner cyan, progress bar green
	spinnerColored := fmt.Sprintf("\033[36m%s\033[0m", spinner)
	barColored := fmt.Sprintf("\033[32m%s\033[90m%s\033[0m",
		strings.Repeat("█", filled),
		strings.Repeat("░", barWidth-filled))

	return fmt.Sprintf("%s Running tests [%s] \033[1m%d/%d\033[0m \033[90m(%.0f%%)\033[0m",
		spinnerColored, barColored, current, total, percentage)
}

// formatSpinner creates a spinner display for unknown progress
func (p *Progress) formatSpinner(current int, spinner string) string {
	if p.noColor {
		return fmt.Sprintf("%s Running tests... %d completed", spinner, current)
	}

	spinnerColored := fmt.Sprintf("\033[36m%s\033[0m", spinner)
	return fmt.Sprintf("%s Running tests... \033[1m%d\033[0m completed", spinnerColored, current)
}

// clear removes the progress line
func (p *Progress) clear() {
	if !p.isTerminal {
		return
	}
	// Clear line and move cursor to beginning
	fmt.Fprintf(p.writer, "\r\033[K")
}
