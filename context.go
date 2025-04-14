package easy_web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Context struct {
	Req *http.Request
	Rsp http.ResponseWriter

	StatusCode int
	Data       []byte

	matchedPath string
	pathParams  map[string]string
	queryParams map[string][]string
}

// BindJson bind json request body to v
func (c *Context) BindJson(v any) error {
	if c.Req.Body == nil {
		return errors.New("[easy_web] request body is nil")
	}

	dec := json.NewDecoder(c.Req.Body)
	return dec.Decode(v)
}

func (c *Context) FormParam(key string) ParamVal {
	if err := c.Req.ParseForm(); err != nil {
		return ParamVal{
			err: err,
		}
	}

	return ParamVal{
		val: c.Req.FormValue(key),
		err: nil,
	}
}

func (c *Context) PathParam(key string) ParamVal {
	if val, ok := c.pathParams[key]; ok {
		return ParamVal{val: val, err: nil}
	}

	return ParamVal{
		err: errors.New("[easy_web] path param not found"),
	}
}

func (c *Context) QueryParam(key string) ParamVal {
	if c.queryParams == nil {
		c.queryParams = c.Req.URL.Query()
	}

	if vals, ok := c.queryParams[key]; ok {
		return ParamVal{val: vals[0], err: nil}
	}

	return ParamVal{
		err: errors.New("[easy_web] query param not found"),
	}
}

// RspBytes response with bytes
func (c *Context) RspBytes(code int, data []byte) error {
	c.StatusCode = code
	c.Data = data
	return nil
}

// RspJson json response
func (c *Context) RspJson(code int, data any) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.RspBytes(code, bs)
}

// Ok response with code 200
func (c *Context) Ok() error {
	return c.RspBytes(http.StatusOK, nil)
}

// OkJson json response with code 200
func (c *Context) OkJson(data any) error {
	return c.RspJson(http.StatusOK, data)
}

type ParamVal struct {
	val string
	err error
}

func (s *ParamVal) String() (string, error) {
	return s.val, s.err
}

func (s *ParamVal) AsInt() (int, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.Atoi(s.val)
}

func (s *ParamVal) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseInt(s.val, 10, 64)
}

func (s *ParamVal) AsUint64() (uint64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseUint(s.val, 10, 64)
}

func (s *ParamVal) AsFloat64() (float64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseFloat(s.val, 64)
}
