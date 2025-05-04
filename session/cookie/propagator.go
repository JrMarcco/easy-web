package cookie

import "net/http"

type CPropagator struct {
	cookieName string
	cookieOpt  func(*http.Cookie)
}

func (c *CPropagator) Inject(id string, w http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:  c.cookieName,
		Value: id,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (c *CPropagator) Extract(req *http.Request) (string, error) {
	cookie, err := req.Cookie(c.cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (c *CPropagator) Del(w http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:   c.cookieName,
		MaxAge: -1,
	}

	http.SetCookie(w, cookie)
	return nil
}

type CPropagatorOpt func(*CPropagator)

func CPropagatorWithCookieName(cookieName string) CPropagatorOpt {
	return func(c *CPropagator) {
		c.cookieName = cookieName
	}
}

func CPropagatorWithCookieOpt(cookieOpt func(*http.Cookie)) CPropagatorOpt {
	return func(c *CPropagator) {
		c.cookieOpt = cookieOpt
	}
}

func NewCPropagator(opts ...CPropagatorOpt) *CPropagator {
	cp := &CPropagator{
		cookieName: "session_id",
		cookieOpt:  func(c *http.Cookie) {},
	}

	for _, opt := range opts {
		opt(cp)
	}
	return cp
}
