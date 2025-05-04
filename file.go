package easyweb

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type FileUploader struct {
	FileField   string
	DstPathFunc func(mph *multipart.FileHeader) string
}

func (f FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		file, header, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to upload file: " + err.Error())
			return
		}
		defer func() {
			_ = file.Close()
		}()

		dstPath := f.DstPathFunc(header)
		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		defer func() {
			_ = dstFile.Close()
		}()

		_, err = io.CopyBuffer(dstFile, file, nil)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to upload file: " + err.Error())
			return
		}

		ctx.StatusCode = http.StatusOK
		ctx.Data = []byte("file uploaded successfully")
	}
}
