package commands

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/wormbks/asciinema-edit/cast"
	"golang.org/x/term"
	"gopkg.in/urfave/cli.v1"
)

var Play = cli.Command{
	Name: "play",
	Usage: `
	Plays cast from a file .

EXAMPLES:

     asciinema-edit play  ./123.cast

`,
	ArgsUsage: "[filename]",
	Action:    playAction,
	Flags: []cli.Flag{
		cli.Float64Flag{
			Name:  "speed",
			Usage: "speed of playback",
			Value: 1.0,
		},
		cli.Float64Flag{
			Name:  "idle-time-limit",
			Usage: "limit idle time during playback to given number of seconds",
			Value: 10.0,
		},
	},
}

// playAction plays a cast from a file. It opens the file, decodes
// the cast, creates a castPlayer with the decoded cast and specified
// playback options, and calls play() on it.
func playAction(cc *cli.Context) error {
	inputFile := cc.Args().First()
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	ct, err := cast.Decode(f)
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}
	p := &castPlayer{
		cast:          ct,
		speed:         cc.Float64("speed"),
		idleTimeLimit: cc.Float64("idle-time-limit"),
	}
	return p.play()
}

// castPlayer holds the state needed to play a cast.
type castPlayer struct {
	cast          *cast.Cast
	speed         float64
	idleTimeLimit float64
}

func (p *castPlayer) play() error {
	lastTime := float64(0.0)
	if err := p.cast.Validate(); err != nil {
		return err
	}

	if p.speed == 0.0 {
		p.speed = 1.0
	}

	oldState, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, oldState)

	winSize := &pty.Winsize{
		Rows: uint16(p.cast.Header.Height),
		Cols: uint16(p.cast.Header.Width),
	}
	err = pty.Setsize(os.Stdout, winSize)
	if err != nil {
		return err
	}

	for _, ev := range p.cast.EventStream {
		delay := ev.Time - lastTime
		if delay > p.idleTimeLimit {
			delay = p.idleTimeLimit
		}
		if err := p.playEvent(ev, delay); err != nil {
			return err
		}
		lastTime = ev.Time
	}
	return nil
}

// playEvent plays a cast event, introducing a delay before playing if needed.
// It dispatches to playOutput or resizeTerm based on the event type.
func (p *castPlayer) playEvent(ev *cast.Event, delay float64) error {
	switch ev.Type {
	case "o":
		return p.playOutput(ev, delay)
	case "r":
		return p.resizeTerm(ev, delay)
	default:
		return nil
	}
}

// playOutput plays a cast output event by writing the output to stdout after
// introducing a delay. It unescapes the output data, sleeps for the specified
// delay, and then writes the output to stdout.
func (p *castPlayer) playOutput(ev *cast.Event, delay float64) error {
	buf, err := unescapeString(ev.Data)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(delay * float64(time.Second)))

	_, err = os.Stdout.Write(buf)
	return err
}

func (p *castPlayer) resizeTerm(ev *cast.Event, delay float64) error {
	_ = ev
	// width, heigh, err := parseSize(ev.Data)
	// if err != nil {
	// 	return err
	// }
	time.Sleep(time.Duration(delay * float64(time.Second)))
	// winSize := &pty.Winsize{
	// 	Rows: uint16(width),
	// 	Cols: uint16(heigh),
	// }

	return nil
}

// unescapeString unescapes a string that contains escaped unicode sequences
// and escaped quote characters. It supports the following escapes:
//
// - \\ - Escapes a literal backslash character
// - \" - Escapes a double quote character
// - \uXXXX - Escapes a unicode code point, where XXXX is a 4 digit hex value
//
// Any invalid escape sequences will return an error.
func unescapeString(s string) ([]byte, error) {
	var result []byte
	for i := 0; i < len(s); {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) {
				return nil, fmt.Errorf("invalid escape sequence at position %d", i)
			}
			switch s[i+1] {
			case 'u':
				if i+5 >= len(s) {
					return nil, fmt.Errorf("invalid unicode escape sequence at position %d", i)
				}
				var value rune
				_, err := fmt.Sscanf(s[i+2:i+6], "%04X", &value)
				if err != nil {
					return nil, fmt.Errorf("error parsing unicode escape sequence at position %d: %v", i, err)
				}
				if !unicode.IsPrint(value) {
					return nil, fmt.Errorf("invalid unicode character at position %d", i)
				}
				var buf [utf8.UTFMax]byte
				size := utf8.EncodeRune(buf[:], value)
				result = append(result, buf[:size]...)
				i += 6
			case '\\', '"':
				result = append(result, s[i+1])
				i += 2
			default:
				return nil, fmt.Errorf("invalid escape sequence at position %d", i)
			}
		default:
			result = append(result, s[i])
			i++
		}
	}
	return result, nil
}

// parseSize parses a size string of the form "WxH" into width and height
// integers. It returns the width and height, or an error if the string is
// malformed.
func parseSize(s string) (int, int, error) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid size format: %s", s)
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width value: %v", err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height value: %v", err)
	}

	return width, height, nil
}
