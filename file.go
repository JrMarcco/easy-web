package easyweb

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type FileUploader struct {
	fieldName   string
	dstPathFunc func(mph *multipart.FileHeader) string
}

func (f *FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		file, header, err := ctx.Req.FormFile(f.fieldName)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to upload file: " + err.Error())
			return
		}
		defer func() {
			_ = file.Close()
		}()

		dstPath := f.dstPathFunc(header)
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

type FileUploaderOpt func(*FileUploader)

func FileUploaderWithFieldName(fieldName string) FileUploaderOpt {
	return func(f *FileUploader) {
		f.fieldName = fieldName
	}
}

func FileUploaderWithDstPathFunc(dstPathFunc func(mph *multipart.FileHeader) string) FileUploaderOpt {
	return func(f *FileUploader) {
		f.dstPathFunc = dstPathFunc
	}
}

func NewFileUploader(opts ...FileUploaderOpt) *FileUploader {
	fu := &FileUploader{
		fieldName: "file",
		dstPathFunc: func(mph *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", mph.Filename)
		},
	}

	for _, opt := range opts {
		opt(fu)
	}

	return fu
}

type FileDownloader struct {
	filePath  string
	fieldName string
}

func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		filename, _ := ctx.QueryParam(f.fieldName).String()
		path := filepath.Join(f.filePath, filepath.Clean(filename))

		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment; filename="+filepath.Base(path))
		header.Set("Content-description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}

type FileDownloaderOpt func(*FileDownloader)

func FileDownloaderWithFilePath(filePath string) FileDownloaderOpt {
	return func(fd *FileDownloader) {
		fd.filePath = filePath
	}
}

func FileDownloaderWithFieldName(fieldName string) FileDownloaderOpt {
	return func(fd *FileDownloader) {
		fd.fieldName = fieldName
	}
}

func NewFileDownloader(opts ...FileDownloaderOpt) *FileDownloader {
	fd := &FileDownloader{
		filePath:  filepath.Join("testdata", "upload"),
		fieldName: "file",
	}

	for _, opt := range opts {
		opt(fd)
	}
	return fd
}
