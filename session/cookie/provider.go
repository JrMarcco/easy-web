package cookie

import "net/http"

type Provider struct {
	cookieName string
	cookieOpt  func(cookie *http.Cookie)
}

type ProviderOpt func(provider *Provider)

func NewProvider(opts ...ProviderOpt) *Provider {
	provider := &Provider{
		cookieName: "_session_id",
	}

	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func ProviderWithName(name string) ProviderOpt {
	return func(provider *Provider) {
		provider.cookieName = name
	}
}

func ProviderWithCookieOpt(cookieOpt func(cookie *http.Cookie)) ProviderOpt {
	return func(provider *Provider) {
		provider.cookieOpt = cookieOpt
	}
}

func (p *Provider) Inject(id string, writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
	}

	p.cookieOpt(cookie)
	http.SetCookie(writer, cookie)
	return nil
}

func (p *Provider) Extract(req *http.Request) (string, error) {
	cookie, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, err
}

func (p *Provider) Remove(writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1,
	}

	http.SetCookie(writer, cookie)
	return nil
}
