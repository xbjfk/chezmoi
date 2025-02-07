package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestGPGEncryption(t *testing.T) {
	command := lookPathOrSkip(t, "gpg")

	tempDir := t.TempDir()
	key, passphrase, err := chezmoitest.GPGGenerateKey(command, tempDir)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		symmetric bool
	}{
		{
			name:      "asymmetric",
			symmetric: false,
		},
		{
			name:      "symmetric",
			symmetric: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testEncryption(t, &GPGEncryption{
				Command: command,
				Args: []string{
					"--homedir", tempDir,
					"--no-tty",
					"--passphrase", passphrase,
					"--pinentry-mode", "loopback",
				},
				Recipient: key,
				Symmetric: tc.symmetric,
			})
		})
	}
}
