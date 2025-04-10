package elfin

import (
	"bytes"
	"io"
	"net/http"
)

func CopyReadCloser(req *http.Request, fn func(body []byte)) error {
	if b, err := io.ReadAll(req.Body); err != nil {
		return err
	} else {
		fn(b)
		req.Body = io.NopCloser(bytes.NewBuffer(b))
		return nil
	}
}
