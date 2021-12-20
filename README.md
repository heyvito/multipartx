# multipartx

**multipartx** implements a small implementation like `mime/multipart`, but
allows files from being read from disk directly to the request's body.

## Installing

```
go get github.com/heyvito/multipartx@v0.1.0
```

## Usage

When creating a request, create a new `Multipart` instance, and set fields/files
using the provided methods. For instance:

```go
package main

import (
    "net/http"
    "github.com/heyvito/multipartx"
)

func main() {
    m := &multipartx.Multipart{}
    m.AddField("name", "Paul Appleseed")
    m.AddField("email", "paul.a@example.org")
    err := m.AddFileFromDisk("file", "/path/to/some/large/file")
    if err != nil {
        // File could not be opened.
        // ...
    }

    req, err := http.NewRequest("POST", "https://example.org/foo", nil)
    if err != nil {
        // ...
    }
    m.AttachToRequest(req)
    response, err := http.DefaultClient.Do(req)
    // ...
}

```


## License

```
The MIT License (MIT)

Copyright (c) 2021 Victor Gama de Oliveira

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
