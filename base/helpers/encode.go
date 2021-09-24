package helpers

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"

	"github.com/gabriel-vasile/mimetype"
)

func Encode(bin []byte) []byte {
	e64 := base64.StdEncoding

	maxEncLen := e64.EncodedLen(len(bin))
	encBuf := make([]byte, maxEncLen)

	e64.Encode(encBuf, bin)
	return encBuf
}

func GetExportFilename(name, mimeType string) string {
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(extensions) == 0 {
		return name
	}

	return name + extensions[0]
}

// Convert an io.Reader to an array of bytes
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// Get the content type of the magic numbers of the file, return the content type, return the original file
func GetContentAndTypeByReader(reader io.Reader) (contentType string, multiReader io.Reader, err error) {
	// Set header size to 261 bytes.
	mimetype.SetLimit(261)
	testBytes := StreamToByte(reader)
	inputReader := bytes.NewReader(testBytes)
	// We only have to pass the file header = first 261 bytes
	header := bytes.NewBuffer(nil)
	// After DetectReader, the first bytes are stored in buf.
	mime, err := mimetype.DetectReader(io.TeeReader(inputReader, header))
	if err == nil {
		// Concatenate back the first bytes.
		// reusableReader now contains the complete, original data.
		reusableReader := io.MultiReader(header, inputReader)
		return mime.String(), reusableReader, nil
	} else {
		return "", nil, err
	}
}
