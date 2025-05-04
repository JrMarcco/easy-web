package redis

import (
	easyweb "github.com/JrMarcco/easy-web"
	"github.com/JrMarcco/easy-web/session"
	"github.com/JrMarcco/easy-web/session/cookie"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"testing"
)

func TestSession_Redis(t *testing.T) {
	srv := easyweb.NewHttpServer()

	g := srv.Group("/")

	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.3.3:6379",
		Password: "<passwd>",
	})

	m := session.NewManager(
		NewRStore(client),
		cookie.NewCPropagator(),
	)
	g.Use(func(next easyweb.HandleFunc) easyweb.HandleFunc {
		return func(ctx *easyweb.Context) {
			if ctx.Req.URL.Path == "/login" || ctx.Req.URL.Path == "/logout" {
				next(ctx)
				return
			}

			_, err := m.GetSession(ctx)
			if err != nil {
				ctx.StatusCode = http.StatusUnauthorized
				ctx.Data = []byte("Unauthorized")
				return
			}

			err = m.RefreshSession(ctx)
			if err != nil {
				log.Println(err)
			}
			next(ctx)
		}
	})

	g.Route(http.MethodGet, "/login", func(ctx *easyweb.Context) {
		s, err := m.NewSession(ctx)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("Internal Server Error")
			return
		}

		err = s.Set(ctx.Req.Context(), "username", "jrmarcco")
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("Internal Server Error")
			return
		}

		ctx.StatusCode = http.StatusOK
		ctx.Data = []byte("login success")
	})

	g.Route(http.MethodGet, "/logout", func(ctx *easyweb.Context) {
		err := m.DelSession(ctx)
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("failed to logout")
			return
		}

		ctx.StatusCode = http.StatusOK
		ctx.Data = []byte("logout success")
	})

	g.Route(http.MethodGet, "/user", func(ctx *easyweb.Context) {
		s, _ := m.GetSession(ctx)
		username, err := s.Get(ctx.Req.Context(), "username")
		if err != nil {
			ctx.StatusCode = http.StatusInternalServerError
			ctx.Data = []byte("Internal Server Error")
			return
		}

		ctx.StatusCode = http.StatusOK
		ctx.Data = []byte(username.(string))
		return
	})

	err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}
