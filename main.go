package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return fmt.Errorf("open /dev/tty: %w", err)
	}
	defer tty.Close()

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(int(tty.Fd()), oldState)

	oldStdout, err := term.MakeRaw(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("make raw stdout: %w", err)
	}
	defer term.Restore(int(os.Stdout.Fd()), oldStdout)

	_, height, err := term.GetSize(int(tty.Fd()))
	if err != nil {
		return fmt.Errorf("get term size: %w", err)
	}

	scrollHeight := height - 1

	fmt.Fprintf(os.Stdout, "\033[1;%dr", scrollHeight)
	// Clear scrollback and screen, same as redraw does on filter change
	fmt.Fprintf(os.Stdout, "\033[3J")
	for i := 1; i <= scrollHeight; i++ {
		fmt.Fprintf(os.Stdout, "\033[%d;1H\033[K", i)
	}
	defer func() {
		fmt.Fprintf(os.Stdout, "\033[r\033[%d;1H\r\n", height)
	}()

	// Buffer all log lines so we can re-render on filter change
	var allLines []string

	pipeLines := make(chan string)
	pipeErr := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			pipeLines <- scanner.Text()
		}
		pipeErr <- scanner.Err()
		close(pipeLines)
	}()

	input := []byte{}
	// activeFilters holds the last valid parsed filter, used for incoming lines
	var activeFilters []Filter

	renderPrompt := func() {
		prompt := ">"
		if len(input) > 0 {
			_, valid := parseFilters(string(input))
			if valid {
				prompt = "\033[32m>\033[0m" // green
			} else {
				prompt = "\033[31m>\033[0m" // red
			}
		}
		fmt.Fprintf(os.Stdout, "\033[%d;1H\033[K%s %s", height, prompt, input)
	}

	// Redraw the scroll region with lines matching the given filters.
	redraw := func(filters []Filter) {
		var matched []string
		for _, l := range allLines {
			if matchesFilter(l, filters) {
				matched = append(matched, l)
			}
		}
		// Clear scrollback buffer and entire scroll region
		fmt.Fprintf(os.Stdout, "\033[3J")
		for i := 1; i <= scrollHeight; i++ {
			fmt.Fprintf(os.Stdout, "\033[%d;1H\033[K", i)
		}
		// Show the last scrollHeight lines, bottom-aligned in the scroll region
		visible := matched
		if len(visible) > scrollHeight {
			visible = visible[len(visible)-scrollHeight:]
		}
		offset := scrollHeight - len(visible)
		for i, l := range visible {
			fmt.Fprintf(os.Stdout, "\033[%d;1H%s", offset+i+1, l)
		}
	}

	renderPrompt()

	ttyBytes := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := tty.Read(buf)
			if err != nil || n == 0 {
				return
			}
			ttyBytes <- buf[0]
		}
	}()

	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	// afterInput is called after any input change.
	// Only redraws if the new input parses into a valid, changed filter.
	afterInput := func() {
		filters, valid := parseFilters(string(input))
		if !valid {
			renderPrompt()
			return
		}
		if !filtersEqual(filters, activeFilters) {
			activeFilters = filters
			redraw(activeFilters)
		}
		renderPrompt()
	}

	for {
		select {
		case line, ok := <-pipeLines:
			if !ok {
				if err := <-pipeErr; err != nil {
					return err
				}
				return nil
			}
			allLines = append(allLines, line)
			if matchesFilter(line, activeFilters) {
				fmt.Fprintf(os.Stdout, "\033[%d;1H\r\n%s", scrollHeight, line)
			}
			renderPrompt()

		case b := <-ttyBytes:
			switch {
			case b == 3: // Ctrl+C
				return nil
			case b == 13: // Enter
				input = input[:0]
				afterInput()
			case b == 127 || b == 8: // Backspace
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
				afterInput()
			case b >= 32 && b < 127: // Printable ASCII
				input = append(input, b)
				afterInput()
			}

		case <-sigwinch:
			_, height, _ = term.GetSize(int(tty.Fd()))
			scrollHeight = height - 1
			fmt.Fprintf(os.Stdout, "\033[1;%dr", scrollHeight)
			redraw(activeFilters)
			renderPrompt()
		}
	}
}
