package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormats(t *testing.T) {
	assert.Contains(t, Formats, "json")
	assert.Contains(t, Formats, "toml")
	assert.Contains(t, Formats, "yaml")
}

func TestFormatRoundTrip(t *testing.T) {
	type value struct {
		Bool   bool
		Bytes  []byte
		Int    int
		Float  float64
		Object map[string]interface{}
		String string
	}

	for _, format := range []Format{
		formatGzippedJSON{},
		formatJSON{},
		formatTOML{},
		formatYAML{},
	} {
		t.Run(format.Name(), func(t *testing.T) {
			v := value{
				Bool:  true,
				Bytes: []byte("bytes"),
				Int:   1,
				Float: 2.3,
				Object: map[string]interface{}{
					"key": "value",
				},
				String: "string",
			}
			data, err := format.Marshal(v)
			require.NoError(t, err)
			var actualValue value
			require.NoError(t, format.Unmarshal(data, &actualValue))
			assert.Equal(t, v, actualValue)
		})
	}
}
