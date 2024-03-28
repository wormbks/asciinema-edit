// Package cast contains the essential structures for dealing with
// an asciinema cast of the v2 format.
//
// The current implementation is based on the V2 format as of July 2nd, 2018.
//
// From [1], asciicast v2 file is a newline-delimited JSON file where:
//
//   - first line contains header (initial terminal size, timestamp and other
//     meta-data), encoded as JSON object; and
//   - all following lines form an event stream, each line representing a separate
//     event, encoded as 3-element JSON array.
//
// [1]: https://github.com/asciinema/asciinema/blob/49a892d9e6f57ab3a774c0835fa563c77cf6a7a7/doc/asciicast-v2.md.
package cast

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

// Header represents the asciicast header - a JSON-encoded object containing
// recording meta-data.
type Header struct {
	// Version represents the version of the current ascii cast format
	// (must be `2`).
	//
	// This field is required for a valid header.
	Version uint8 `json:"version"`

	// With is the initial terminal width (number of columns).
	//
	// This field is required for a valid header.
	Width uint `json:"width"`

	// Height is the initial terminal height (number of rows).
	//
	// This field is required for a valid header.
	Height uint `json:"height"`

	// Timestamp is the unix timestamp of the beginning of the
	// recording session.
	Timestamp uint `json:"timestamp,omitempty"`

	// Command corresponds to the name of the command that was
	// recorded.
	Command string `json:"command,omitempty"`

	// Theme describes the color theme of the recorded terminal.
	// Theme struct {
	// 	// Fg corresponds to the normal text color (foreground).
	// 	Fg string `json:"fg,omitempty"`

	// 	// Bg corresponds to the normal background color.
	// 	Bg string `json:"bg,omitempty"`

	// 	// Palette specifies a list of 8 or 16 colors separated by
	// 	// colon character to apply a theme to the session
	// 	Palette string `json:"palette,omitempty"`
	// } `json:"theme,omitempty"`

	// // Title corresponds to the title of the cast.
	Title string `json:"title,omitempty"`

	// // IdleTimeLimit specifies the maximum amount of idleness between
	// one command and another.
	IdleTimeLimit float64 `json:"idle_time_limit,omitempty"`

	// Env specifies a map of environment variables captured by the
	// asciinema command.
	//
	// ps.: the official asciinema client only captures `SHELL` and `TERM`.
	Env struct {
		// Shell corresponds to the captured SHELL environment variable.
		Shell string `json:"SHELL,omitempty"`

		// Term corresponds to the captured TERM environment variable.
		Term string `json:"TERM,omitempty"`
	} `json:"env,omitempty"`
}

// Event represents terminal inputs that get recorded by asciinema.
type Event struct {
	// Time indicates when this event happened, represented as the number
	// of seconds since the beginning of the recording session.
	Time float64

	// Type represents the type of the data that's been recorded.
	//
	// Four types are possible:
	//   - "o": data written to stdout; and
	//   - "i": data read from stdin.
	//   - "r": change window size
	//   - "m": marker
	Type string

	// Data represents the data recorded from the terminal.
	Data string
}

// Cast represents the whole asciinema session.
type Cast struct {
	// Header presents the recording metadata.
	Header Header

	// EventStream contains all the events that were generated during
	// the recording.
	EventStream []*Event
}

// ValidateHeader verifies whether the provided `cast` header structure is valid
// or not based on the asciinema cast v2 protocol.
func (header *Header) ValidateHeader() error {

	if header.Version != 2 {
		return errors.Errorf("only casts with version 2 are valid")
	}

	if header.Width == 0 {
		return errors.Errorf("a valid width (>0) must be specified")
	}

	if header.Height == 0 {
		return errors.Errorf("a valid height (>0) must be specified")
	}

	return nil
}

func (header *Header) Encode(e *json.Encoder) error {
	if e == nil {
		return errors.Errorf("encoder must not be nil")
	}
	return e.Encode(header)
}

// ValidateEvent checks whether the provided `Event` is properly formed.
func (event *Event) ValidateEvent() error {
	if event == nil {
		return errors.Errorf("event must not be nil")
	}

	switch event.Type {
	case "i", "o", "r", "m":
		return nil
	default:
		return errors.Errorf("type must either be 'o', 'i', 'r', or 'm'")
	}
}

func (ev *Event) Encode(e *json.Encoder) error {
	if e == nil {
		return errors.Errorf("encoder must not be nil")
	}
	return e.Encode([]interface{}{ev.Time, ev.Type, ev.Data})
}

// ValidateEventStream makes sure that a given set of events (event stream)
// is valid.
//
// A valid stream must:
// - be ordered by time; and
// - have valid events.
func ValidateEventStream(eventStream []*Event) error {
	var lastTime float64

	for _, ev := range eventStream {
		if ev.Time < lastTime {
			return errors.Errorf("events must be ordered by time")
		}

		err := ev.ValidateEvent()
		if err != nil {
			return errors.Wrapf(err, "invalid event")
		}

		lastTime = ev.Time
	}

	return nil
}

// Validate makes sure that the supplied cast is valid.
func (cast *Cast) Validate() error {
	err := cast.Header.ValidateHeader()
	if err != nil {
		return errors.Wrapf(err, "invalid header")
	}

	err = ValidateEventStream(cast.EventStream)
	if err != nil {
		return errors.Wrapf(err, "invalid event stream")
	}

	return nil
}

// Encode writes the encoding of `Cast` into the writer passed as an argument.
//
// ps.: this method **will not** validate whether the cast is a valid V2
// cast or not. Make sure you call `Validate` before.
func (cast *Cast) Encode(writer io.Writer) error {
	if writer == nil {
		return errors.New("a writer must be specified")
	}

	if cast == nil {
		return errors.New("a cast must be specified")
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "")

	if err := cast.Header.Encode(encoder); err != nil {
		return errors.Wrap(err, "failed to encode header")
	}

	for _, ev := range cast.EventStream {
		if err := ev.Encode(encoder); err != nil {
			return errors.Wrap(err, "failed to encode event")
		}
	}

	return nil
}

// Decode reads the whole contents of the reader passed as argument, validates
// whether the stream contains a valid asciinema cast and then unmarshals it
// into a cast struct.
func Decode(reader io.Reader) (*Cast, error) {

	cast := &Cast{
		EventStream: make([]*Event, 0),
		Header: Header{
			Version: 2,
		},
	}

	if reader == nil {
		return nil, errors.New("a reader must be specified")
	}

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&cast.Header)
	if err != nil {
		return nil, errors.Wrapf(err,
			"couldn't decode header")
	}

	var (
		ev     = new([3]interface{})
		ok     bool
		time   float64
		evType string
		data   string
	)

	for {
		err = decoder.Decode(ev)
		if err != nil {
			if err == io.EOF {
				return cast, nil
			}

			return nil, errors.Wrapf(err,
				"failed to parse ev line")
		}

		time, ok = ev[0].(float64)
		if !ok {
			return nil, errors.Errorf("first element of event is not a float64")
		}

		evType, ok = ev[1].(string)
		if !ok {
			return nil, errors.Errorf("second element of event is not a string")
		}

		data, ok = ev[2].(string)
		if !ok {
			return nil, errors.Errorf("third element of event is not a string")
		}
		ev := &Event{
			Time: time,
			Type: evType,
			Data: data,
		}
		if err = ev.ValidateEvent(); err != nil {
			return nil, errors.Wrapf(err,
				"invalid event")
		}
		cast.EventStream = append(cast.EventStream, ev)
	}

}
