package easyweb

import (
	"fmt"
	"log"

	lru "github.com/hashicorp/golang-lru"

	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
		filename, err := ctx.QueryParam(f.fieldName).String()
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to download file: " + err.Error())
			return
		}
		if filename == "" {
			ctx.StatusCode = http.StatusBadRequest
			ctx.Data = []byte("filename is empty")
			return
		}

		filename = filepath.Clean(filename)
		path, err := validateFileName(f.filePath, filename)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to download file: " + err.Error())
			return
		}

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

type StaticResourceHandler struct {
	fieldName       string
	filePath        string
	extContentTypes map[string]string
	cache           *lru.Cache
	maxCacheSize    int
}

type cacheItem struct {
	asbPath     string
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

func (srh *StaticResourceHandler) Handle() HandleFunc {
	return func(ctx *Context) {
		fileName, err := ctx.PathParam(srh.fieldName).String()
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to serve static resource: " + err.Error())
			return
		}
		if fileName == "" {
			ctx.StatusCode = http.StatusBadRequest
			ctx.Data = []byte("filename is empty")
		}

		path, err := validateFileName(srh.filePath, fileName)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to serve static resource: " + err.Error())
			return
		}

		if ci, ok := srh.getCachedItem(path); ok {
			srh.writeResp(ctx, ci)
			return
		}

		f, err := os.Open(path)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to serve static resource: " + err.Error())
			return
		}
		defer func() {
			_ = f.Close()
		}()

		ext := srh.getExt(fileName)
		t, ok := srh.extContentTypes[ext]
		if !ok {
			ctx.StatusCode = http.StatusBadRequest
			return
		}

		data, err := io.ReadAll(f)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to serve static resource: " + err.Error())
		}

		ci := &cacheItem{
			asbPath:     path,
			fileName:    fileName,
			fileSize:    len(data),
			contentType: t,
			data:        data,
		}

		srh.cacheFile(ci)

		srh.writeResp(ctx, ci)
	}
}

func (srh *StaticResourceHandler) getExt(fileName string) string {
	index := strings.LastIndex(fileName, ".")
	if index == len(fileName)-1 {
		return ""
	}
	return fileName[index+1:]
}

func (srh *StaticResourceHandler) getCachedItem(absPath string) (*cacheItem, bool) {
	if srh.cache == nil {
		return nil, false
	}

	if ci, ok := srh.cache.Get(absPath); ok {
		return ci.(*cacheItem), true
	}
	return nil, false
}

func (srh *StaticResourceHandler) cacheFile(ci *cacheItem) {
	if srh.cache == nil || srh.maxCacheSize <= 0 || ci.fileSize > srh.maxCacheSize {
		return
	}

	srh.cache.Add(ci.asbPath, ci)
}

func (srh *StaticResourceHandler) writeResp(ctx *Context, ci *cacheItem) {
	ctx.Resp.WriteHeader(http.StatusOK)
	ctx.Resp.Header().Set("Content-Type", ci.contentType)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprintf("%d", ci.fileSize))
	_, _ = ctx.Resp.Write(ci.data)
}

type StaticResourceHandlerOpt func(*StaticResourceHandler)

func StaticResourceHandlerWithFieldName(fieldName string) StaticResourceHandlerOpt {
	return func(srh *StaticResourceHandler) {
		srh.fieldName = fieldName
	}
}

func StaticResourceHandlerWithFilePath(filePath string) StaticResourceHandlerOpt {
	return func(srh *StaticResourceHandler) {
		srh.filePath = filePath
	}
}

func StaticResourceHandlerWithExtContentTypes(extContentTypes map[string]string) StaticResourceHandlerOpt {
	return func(srh *StaticResourceHandler) {
		srh.extContentTypes = extContentTypes
	}
}

func StaticResourceHandlerWithCache(maxCacheCnt, maxCacheSize int) StaticResourceHandlerOpt {
	return func(srh *StaticResourceHandler) {
		c, err := lru.New(maxCacheCnt)
		if err != nil {
			log.Println("failed to create cache: ", err)
		}

		srh.cache = c
		srh.maxCacheSize = maxCacheSize
	}
}

func NewStaticResourceHandler(opts ...StaticResourceHandlerOpt) *StaticResourceHandler {
	srh := &StaticResourceHandler{
		fieldName: "file",
		filePath:  filepath.Join("testdata", "statics"),
		extContentTypes: map[string]string{
			"jpg":  "image/jpeg",
			"jpeg": "image/jpeg",
			"png":  "image/png",
			"gif":  "image/gif",
			"svg":  "image/svg+xml",
			"css":  "text/css",
			"js":   "application/javascript",
		},
	}

	for _, opt := range opts {
		opt(srh)
	}

	return srh
}

// validateFileName validate filename to prevent accessing a system file.
// returns a file absolute path if the filename is validated.
func validateFileName(resourcePath, filename string) (string, error) {
	filename = filepath.Clean(filename)
	path, err := filepath.Abs(filepath.Join(resourcePath, filename))
	if err != nil {
		return "", err
	}

	if !strings.Contains(path, resourcePath) {
		return "", err
	}

	return path, nil
}
