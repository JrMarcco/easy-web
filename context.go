package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request
	Rsp http.ResponseWriter

	TraceCtx context.Context

	MatchedRoute  string
	RspStatusCode int
	RspData       []byte

	UserVals map[string]any

	pathParams map[string]string
	queryVals  url.Values
}

func (c *Context) RspJson(statusCode int, val any) error {

	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}

	c.RspStatusCode = statusCode
	c.RspData = bs

	return err
}

func (c *Context) OkJson(val any) error {
	return c.RspJson(http.StatusOK, val)
}

func (c *Context) PathVal(key string) StringVal {
	if val, ok := c.pathParams[key]; ok {
		return StringVal{
			val: val,
		}
	}

	return StringVal{
		err: errors.New("[ctx] path key not exist"),
	}
}

func (c *Context) QueryVal(key string) StringVal {
	if c.queryVals == nil {
		c.queryVals = c.Req.URL.Query()
	}

	if vals, ok := c.queryVals[key]; ok {
		return StringVal{
			val: vals[0],
		}
	}

	return StringVal{
		err: errors.New("[ctx] query key not exist"),
	}
}

func (c *Context) FormVal(key string) StringVal {
	if err := c.Req.ParseForm(); err != nil {
		return StringVal{
			err: err,
		}
	}

	return StringVal{
		val: c.Req.FormValue(key),
	}
}

func (c *Context) BindJson(val any) error {
	if c.Req.Body == nil {
		return errors.New("[ctx] request body is nil")
	}

	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

type StringVal struct {
	val string
	err error
}

func (s StringVal) AsInt() (int, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.Atoi(s.val)
}

func (s StringVal) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseInt(s.val, 10, 64)
}

func (s StringVal) AsFloat64() (float64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseFloat(s.val, 64)
}
