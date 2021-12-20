package multipartx

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const rn = "\r\n"
const contentTypePrefix = "multipart/form-data; boundary="
const octetType = "application/octet-stream"

type Multipart struct {
	fields     []io.Reader
	boundary   string
	writingIdx int
	closed     bool
	toClose    []*os.File
}

func fieldHeader(fieldName string) string {
	return fmt.Sprintf("Content-Disposition: form-data; name=\"" + fieldName + "\"")
}

func fileHeader(fieldName, fileName, contentType string) string {
	return "Content-Disposition: form-data; name=\"" + fieldName + "\"; filename=\"" + fileName + "\"\r\nContent-Type: " + contentType
}

func (m *Multipart) ensureBoundary() {
	if m.boundary == "" {
		m.boundary = randomBoundary()
	}
}

// AddField adds a new field to the Multipart content with a given name and
// value.
func (m *Multipart) AddField(name, value string) {
	m.ensureBoundary()
	m.fields = append(m.fields,
		bytes.NewReader([]byte(rn+"--"+m.boundary+rn+fieldHeader(name)+rn+rn+value)),
	)
}

// AddBytes adds a new file field to the Multipart content with a given field
// name, filename, and contents. The Content-Type for the part is assumed as
// application/octet-stream. To set the Content-Type for this part, use
// AddBytesContentType.
func (m *Multipart) AddBytes(fieldName string, fileName string, data []byte) {
	m.AddBytesContentType(fieldName, fileName, octetType, data)
}

// AddBytesContentType adds a new file field to the Multipart content with a
// given name, filename, content type and contents.
func (m *Multipart) AddBytesContentType(fieldName, fileName, contentType string, data []byte) {
	m.ensureBoundary()
	val := []byte(rn + "--" + m.boundary + rn)
	val = append(val, []byte(fileHeader(fieldName, fileName, contentType)+rn+rn)...)
	val = append(val, data...)
	m.fields = append(m.fields,
		bytes.NewReader(val),
	)
}

// AddFile adds a new file to the Multipart content. The part's Content-Type is
// assumed as application/octet-stream. To set a specific Content-Type, use
// AddFileContentType.
func (m *Multipart) AddFile(fieldName, fileName string, reader io.Reader) {
	m.AddFileContentType(fieldName, fileName, octetType, reader)
}

// AddFileContentType adds a new file to the Multipart content with a given
// field name, filename, content-type and contents from an io.Reader.
func (m *Multipart) AddFileContentType(fieldName, fileName, contentType string, reader io.Reader) {
	m.ensureBoundary()
	val := []byte(rn + "--" + m.boundary + rn)
	val = append(val, []byte(fileHeader(fieldName, fileName, contentType)+rn+rn)...)
	m.fields = append(m.fields,
		io.MultiReader(
			bytes.NewReader(val),
			reader,
		),
	)
}

// AddFileFromDisk adds a new file to the Multipart content with a given field
// name. It reads the provided filePath from disk, and assumes the part's
// filename to be the same as the one present in filePath, and the Content-Type
// to be application/octet-stream.
// Returns an error in case opening the file under the specified path fails.
// To define a custom Content-Type, use AddFileFromDiskContentType.
func (m *Multipart) AddFileFromDisk(fieldName, filePath string) error {
	return m.AddFileFromDiskContentType(fieldName, octetType, filePath)
}

// AddFileFromDiskContentType adds a file to the Multipart content with a given
// field name. It reads the provided filePath from disk, and assumes the part's
// filename to be the same as the one present in filePath, reporting the
// part's Content-Type as per provided to the contentType parameter.
// Returns an error in case opening the file under the specified path fails.
func (m *Multipart) AddFileFromDiskContentType(fieldName, contentType, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	m.toClose = append(m.toClose, f)
	m.AddFileContentType(fieldName, filepath.Base(filePath), contentType, f)
	return nil
}

// taken from mime/multipart/writer.go
func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

// Done adds the final boundary part to the Multipart stream. It must be called
// before using the io.Reader implementation of Multipart.
func (m *Multipart) Done() {
	m.ensureBoundary()
	if !m.closed {
		m.closed = true
		m.fields = append(m.fields, bytes.NewReader([]byte(rn+"--"+m.boundary+"--"+rn)))
	}
}

// ContentTypeHeaderValue returns the required Content-Type to be used as the
// request's Content-Type header.
func (m *Multipart) ContentTypeHeaderValue() string {
	m.ensureBoundary()
	return contentTypePrefix + m.boundary
}

// AttachToRequest attaches the current multipart stream to the provided request
// by setting its body, Transfer-Encoding, and Content-Type headers. This
// method automatically calls Done.
func (m *Multipart) AttachToRequest(r *http.Request) {
	m.Done()
	r.Header.Set("Content-Type", m.ContentTypeHeaderValue())
	r.TransferEncoding = []string{"chunked"}
	r.Body = io.NopCloser(m)
}

// Read implements io.Reader for Multipart. Either Done or AttachToRequest must
// be called before attempting to read from the Multipart instance. Reading the
// content's of the Multipart instance yields the multipart/form-data contents
// properly formatted.
func (m *Multipart) Read(p []byte) (n int, err error) {
	if !m.closed {
		return 0, fmt.Errorf("multipartx: cannot read from multipartx.Multipart before calling Done()")
	}
	if m.writingIdx >= len(m.fields) {
		for _, f := range m.toClose {
			if e := f.Close(); e != nil {
				return 0, err
			}
		}
		return 0, io.EOF
	}

	e := m.fields[m.writingIdx]
	n, err = e.Read(p)
	if err == io.EOF {
		err = nil
		m.writingIdx++
	}
	return
}
