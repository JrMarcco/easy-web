package easy_web

import "net/http"

type Context struct {
	Req *http.Request
	Rsp http.ResponseWriter
}
