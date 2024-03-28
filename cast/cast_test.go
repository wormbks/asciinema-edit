package cast

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidVersion2Header(t *testing.T) {
	header := Header{Version: 2, Width: 80, Height: 24}
	err := header.ValidateHeader()
	assert.NoError(t, err)
}

func TestInvalidVersionHeader(t *testing.T) {
	header := Header{Version: 1, Width: 80, Height: 24}
	err := header.ValidateHeader()
	assert.Error(t, err)
}

func TestValidWidth(t *testing.T) {
	header := Header{Version: 2, Width: 80, Height: 24}
	err := header.ValidateHeader()
	assert.NoError(t, err)
}

func TestInvalidWidth(t *testing.T) {
	header := Header{Version: 2, Width: 0, Height: 24}
	err := header.ValidateHeader()
	assert.Error(t, err)
}

func TestValidHeight(t *testing.T) {
	header := Header{Version: 2, Width: 80, Height: 24}
	err := header.ValidateHeader()
	assert.NoError(t, err)
}

func TestInvalidHeight(t *testing.T) {
	header := Header{Version: 2, Width: 80, Height: 0}
	err := header.ValidateHeader()
	assert.Error(t, err)
}

func TestValidateEvent(t *testing.T) {
	validEvent := &Event{Type: "i"}
	err := validEvent.ValidateEvent()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	validEvent.Type = "o"
	err = validEvent.ValidateEvent()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	validEvent.Type = "r"
	err = validEvent.ValidateEvent()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	validEvent.Type = "m"
	err = validEvent.ValidateEvent()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	invalidEvent := &Event{Type: "invalid"}
	err = invalidEvent.ValidateEvent()
	if err == nil {
		t.Error("Expected an error, but got none")
	}

	nilEvent := (*Event)(nil)
	err = nilEvent.ValidateEvent()
	if err == nil {
		t.Error("Expected an error, but got none")
	}
}

func TestCast_Encode(t *testing.T) {
	cast := Cast{
		Header: Header{Version: 2, Width: 80, Height: 24},
		EventStream: []*Event{
			{Time: 1.0, Type: "o", Data: "data1"},
			{Time: 2.0, Type: "r", Data: "10x20"},
		},
	}

	// Test case 1: Testing with a nil writer
	err := cast.Encode(nil)
	if err == nil {
		t.Errorf("Expected an error when writer is nil, but got nil")
	}

	// Test case 2: Testing with a nil cast
	var nilCast *Cast
	err = nilCast.Encode(os.Stdout)
	if err == nil {
		t.Errorf("Expected an error when cast is nil, but got nil")
	}

	// Test case 3: Testing with a valid writer and cast
	writer := &bytes.Buffer{}
	err = cast.Encode(writer)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Add more test cases as needed
}
func TestCast_Decode(t *testing.T) {

	// Test case with nil reader
	_, err := Decode(nil)
	assert.Error(t, err)

	// Test case with invalid header
	reader := bytes.NewBuffer([]byte("invalid"))
	_, err = Decode(reader)
	assert.Error(t, err)

	// Test case with valid header but invalid event
	header := &Header{Version: 2, Width: 80, Height: 24}
	event := &Event{Type: "invalid"}
	writer := &bytes.Buffer{}
	e := json.NewEncoder(writer)
	err = header.Encode(e)
	assert.NoError(t, err)
	err = event.Encode(e)
	assert.NoError(t, err)

	_, err = Decode(writer)
	assert.Error(t, err)

	// Test case with valid header and events
	header = &Header{Version: 2, Width: 80, Height: 24}
	event1 := &Event{Time: 1.0, Type: "o", Data: "data1"}
	event2 := &Event{Time: 2.0, Type: "r", Data: "10x20"}

	writer.Reset()
	e = json.NewEncoder(writer)
	err = header.Encode(e)
	assert.NoError(t, err)
	err = event1.Encode(e)
	assert.NoError(t, err)
	err = event2.Encode(e)
	assert.NoError(t, err)

	cast, err := Decode(writer)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cast.EventStream))

}

func TestHeader_Encode(t *testing.T) {

	// Test case with nil writer
	var header Header
	err := header.Encode(nil)
	assert.Error(t, err)

	// Test case with valid header
	header = Header{Version: 2, Width: 80, Height: 24}
	writer := &bytes.Buffer{}
	e := json.NewEncoder(writer)

	err = header.Encode(e)
	assert.NoError(t, err)

	// Validate encoded output
	expected := []byte("{\"version\":2,\"width\":80,\"height\":24,\"env\":{}}\n")
	assert.Equal(t, expected, writer.Bytes())

}

func TestEvent_Encode(t *testing.T) {

	// Test case with nil writer
	var event Event
	err := event.Encode(nil)
	assert.Error(t, err)

	// Test case with valid event
	event = Event{Time: 1.0, Type: "o", Data: "data"}
	writer := &bytes.Buffer{}
	e := json.NewEncoder(writer)

	err = event.Encode(e)
	assert.NoError(t, err)

	// Validate encoded output
	expected := []byte("[1,\"o\",\"data\"]\n")
	assert.Equal(t, expected, writer.Bytes())

}
