package clipboard

import atottoclipboard "github.com/atotto/clipboard"

var readClipboard = atottoclipboard.ReadAll

func ReadText() (string, error) {
	return readClipboard()
}
