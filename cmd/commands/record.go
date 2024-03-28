package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"github.com/wormbks/asciinema-edit/cast"
	"golang.org/x/term"
	"gopkg.in/urfave/cli.v1"
)

var Record = cli.Command{
	Name: "record",
	Usage: `
	Records cast to output file .

EXAMPLES:

     asciinema-edit rec  ./123.cast

`,
	ArgsUsage: "[filename]",
	Action:    recordAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "args",
			Usage: "shell command arguments",
		},
		cli.StringFlag{
			Name:   "shell",
			Usage:  "shell command",
			EnvVar: "SHELL",
			Value:  "bash",
		},
	},
}

func recordAction(c *cli.Context) (err error) {
	outputName := c.Args().First()
	shellArgs := c.String("out")
	shell := c.String("shell")

	if shell == "" {
		shell = "bash"
	}
	maxWindowSize := "200x50" //flag.String("max-win-size", "200x50", "The maximum window size for the terminal (columns x rows). Ex: 150x40")

	winChangedSig := make(chan os.Signal, 1)
	signal.Notify(winChangedSig, syscall.SIGWINCH)

	scriptWriter := &scriptWriter{
		outFileName: outputName,
		shell:       shell,
	}
	fmt.Printf("Script started, output file is %s\n\n\r", outputName)

	oldState, err := term.MakeRaw(0)
	defer term.Restore(0, oldState)

	cmd := exec.Command(shell, strings.Fields(shellArgs)...)
	scriptWriter.winSize, err = pty.GetsizeFull(os.Stdin)
	if err != nil {
		log.Printf("Can't get window size: %s", err.Error())
		return
	}

	ptyMaster, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("Cannot start the command: %s\n\r", err.Error())
		return
	}

	maxCols := -1
	maxRows := -1
	if maxWindowSize != "" {
		_, err := fmt.Sscanf(maxWindowSize, "%dx%d", &maxCols, &maxRows)

		if err != nil {
			fmt.Printf("Cannot parse <%s> for maximum window size", maxWindowSize)
			return err
		}
	}

	setSetTerminalSize := func(writeEvent bool) (cols, rows int) {
		winSize, err := pty.GetsizeFull(os.Stdin)

		if err != nil {
			log.Printf("Can't get window size: %s", err.Error())
			return
		}
		cols = int(winSize.Cols)
		rows = int(winSize.Rows)

		if maxWindowSize != "" {
			if maxCols != -1 && maxRows != -1 {
				rows = clamp(1, maxRows, rows)
				cols = clamp(1, maxCols, cols)
			}
		}

		winSize.Cols = uint16(cols)
		winSize.Rows = uint16(rows)

		pty.Setsize(ptyMaster, winSize)
		if writeEvent {
			scriptWriter.WriteSize(WindowSizeT{cols: cols, rows: rows})
		}
		return
	}

	go func() {
		for val := range winChangedSig {
			_ = val
			setSetTerminalSize(true)
		}
	}()

	cols, rows := setSetTerminalSize(false)
	err = scriptWriter.Begin(WindowSizeT{cols: cols, rows: rows}, shell)
	if err != nil {
		fmt.Printf("Cannot create output. Error: %s", err.Error())
		return
	}

	allWriter := io.MultiWriter(os.Stdout, scriptWriter)

	go func() {
		io.Copy(allWriter, ptyMaster)
	}()

	go func() {
		io.Copy(ptyMaster, os.Stdin)
	}()

	cmd.Wait()

	scriptWriter.End()
	fmt.Printf("\nScript done! output file is %s\n\r", outputName)

	return nil
}

type WindowSizeT struct {
	rows, cols int
}

func clamp(min, max, val int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

type scriptWriter struct {
	outFileName    string
	outputFile     *os.File
	shell          string
	timestampStart time.Time
	winSize        *pty.Winsize
}

func escapeNonPrintableChars(data []byte) string {
	var result string
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// Handle invalid UTF-8 sequence
			result += fmt.Sprintf("\\u%04X", data[i])
			i++
		} else {
			// Escape special JSON characters
			switch r {
			case '"':
				result += "\\\""
			case '\\':
				result += "\\\\"
			default:
				// Handle other non-printable characters
				if r < ' ' {
					result += fmt.Sprintf("\\u%04X", r)
				} else {
					result += string(r)
				}
			}
			i += size
		}
	}
	return result
}

// func escapeNonPrintableChars(data []byte) string {
// 	result := ""
// 	for i := 0; i < len(data); {
// 		r, size := utf8.DecodeRune(data[i:])
// 		if r == utf8.RuneError && size == 1 {
// 			// Handle non-printable characters
// 			result += fmt.Sprintf("\\u%04X", data[i])
// 			i++
// 		} else {
// 			// Escape special JSON characters
// 			if r == '"' {
// 				result += "\\\""
// 			} else if r == '\\' {
// 				result += "\\\\"
// 			} else if r < ' ' {
// 				// Handle other non-printable characters
// 				result += fmt.Sprintf("\\u%04X", r)
// 			} else {
// 				result += string(r)
// 			}
// 			i += size
// 		}
// 	}
// 	return result
// }

func (w *scriptWriter) WriteData(data []byte) {
	timestamp := time.Since(w.timestampStart).Seconds()

	// https://docs.asciinema.org/manual/asciicast/v2/
	fmt.Fprintf(w.outputFile, "[%f,\"o\",\"%s\"]\n", timestamp, escapeNonPrintableChars(data))
}

func (w *scriptWriter) WriteSize(size WindowSizeT) {
	ts := time.Since(w.timestampStart).Seconds()

	// https://docs.asciinema.org/manual/asciicast/v2/
	fmt.Fprintf(w.outputFile, "[%f,\"r\",\"%dx%d\"]\n", ts, size.cols, size.rows)
}

func (w *scriptWriter) Begin(size WindowSizeT, shell string) error {
	var err error
	w.outputFile, err = os.Create(w.outFileName)
	if err != nil {
		panic(err.Error())
	}

	w.timestampStart = time.Now()
	header := cast.Header{
		Version: 2,
		Width:   uint(size.cols),
		Height:  uint(size.rows),
	}
	header.Env.Term = os.Getenv("TERM")
	header.Env.Shell = shell
	header.Timestamp = uint(time.Now().Unix())
	encoder := json.NewEncoder(w.outputFile)
	encoder.SetIndent("", "")

	err = encoder.Encode(&header)
	if err != nil {

		return errors.Wrapf(err,
			"failed to encode header")
	}

	return nil
}

func (w *scriptWriter) End() error {

	return w.outputFile.Close()
}

func (w *scriptWriter) Write(data []byte) (n int, err error) {

	w.WriteData(data)
	return len(data), err
}
