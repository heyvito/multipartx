package multipartx

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormFields(t *testing.T) {
	m := &Multipart{}
	m.AddField("test1", "value1")
	m.AddField("test2", "value2")
	m.AddBytes("bytes", "data.bin", []byte("hello!"))
	err := m.AddFileFromDisk("file", "./fixture/test.txt")
	require.NoError(t, err)
	m.Done()
	req := httptest.NewRequest("POST", "/", m)
	m.AttachToRequest(req)
	d, err := httputil.DumpRequest(req, true)
	require.NoError(t, err)
	r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(d)))
	require.NoError(t, err)
	require.NoError(t, r.ParseMultipartForm(3.2e+07))
	assert.Equal(t, "value1", r.Form.Get("test1"))
	assert.Equal(t, "value2", r.Form.Get("test2"))
	f := r.MultipartForm.File["bytes"]
	assert.Len(t, f, 1)
	f0 := f[0]
	assert.Equal(t, "data.bin", f0.Filename)
	fCont, err := f[0].Open()
	assert.NoError(t, err)
	data, err := io.ReadAll(fCont)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello!"), data)
}
