package transformer

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wormbks/asciinema-edit/cast"
)

type DummyTransformation struct{}

func (t *DummyTransformation) Transform(c *cast.Cast) error {
	return nil
}

func TestTransformer(t *testing.T) {

	t.Run("with nil transform", func(t *testing.T) {
		_, err := New(nil, "", "")
		assert.Error(t, err)
	})

	t.Run("with transformation", func(t *testing.T) {
		var (
			transformation Transformation
			tempDir        string
			err            error
		)

		setup := func() {
			transformation = &DummyTransformation{}
			tempDir, err = ioutil.TempDir("", "")
			assert.NoError(t, err)
		}

		teardown := func() {
			os.RemoveAll(tempDir)
		}

		t.Run("having empty input and output", func(t *testing.T) {
			setup()
			defer teardown()

			_, err := New(transformation, "", "")
			assert.NoError(t, err)
		})

		t.Run("with input specified", func(t *testing.T) {
			var input string

			t.Run("fails if it doesn't exist", func(t *testing.T) {
				setup()
				defer teardown()

				input = path.Join(tempDir, "inexistent")

				_, err := New(transformation, input, "")
				assert.Error(t, err)
			})

			t.Run("fails if is a directory", func(t *testing.T) {
				setup()
				defer teardown()

				input = tempDir
				_, err := New(transformation, input, "")
				assert.Error(t, err)
			})

			t.Run("succeeds if file exists", func(t *testing.T) {
				setup()
				defer teardown()

				inputFile, err := ioutil.TempFile(tempDir, "")
				assert.NoError(t, err)

				input = inputFile.Name()

				_, err = New(transformation,
					inputFile.Name(), "")
				assert.NoError(t, err)
			})
		})

		t.Run("with output specified", func(t *testing.T) {
			var output string

			t.Run("fails if directory doesn't exist", func(t *testing.T) {
				setup()
				defer teardown()

				output = "/inexistent/directory/file.txt"

				_, err = New(transformation, "", output)
				assert.Error(t, err)
			})

			t.Run("creates file if it doesn't exist in existing directory", func(t *testing.T) {
				setup()
				defer teardown()

				output = path.Join(tempDir, "output-file")

				_, err = New(transformation, "", output)
				assert.NoError(t, err)

				_, err = os.Stat(output)
				assert.NoError(t, err)
			})

			t.Run("succeeds if file exists", func(t *testing.T) {
				setup()
				defer teardown()

				outputFile, err := ioutil.TempFile(tempDir, "")
				assert.NoError(t, err)

				output = outputFile.Name()

				_, err = New(transformation, "", output)
				assert.NoError(t, err)
			})
		})
	})

	t.Run("transform", func(t *testing.T) {
		var (
			trans  *Transformer
			input  string
			output = "/dev/null"
			err    error
		)

		setup := func(content string) {
			input, err = createTempFileWithContent(content)
			assert.NoError(t, err)

			trans, err = New(
				&DummyTransformation{},
				input,
				output)
			assert.NoError(t, err)
		}

		teardown := func() {
			os.Remove(input)
		}

		t.Run("with malformed input", func(t *testing.T) {
			setup("malformed")
			defer teardown()

			err = trans.Transform()
			assert.Error(t, err)
		})

		t.Run("with malformed event stream", func(t *testing.T) {
			setup(`{"version": 2, "width": 123, "height": 123}
[1, "o", "aaa"]
[3, "o", "ccc"]
[2, "o", "bbb"]`)
			defer teardown()

			err = trans.Transform()
			assert.Error(t, err)
		})

		t.Run("with well formed event stream", func(t *testing.T) {
			setup(`{"version": 2, "width": 123, "height": 123}
[1, "o", "aaa"]
[2, "o", "bbb"]
[3, "o", "ccc"]`)
			defer teardown()

			err = trans.Transform()
			assert.NoError(t, err)
		})
	})
}

func createTempFileWithContent(content string) (res string, err error) {
	var file *os.File

	file, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	if err != nil {
		return
	}

	res = file.Name()
	return
}
