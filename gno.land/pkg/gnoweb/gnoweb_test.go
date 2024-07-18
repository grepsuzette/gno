package gnoweb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gotuna/gotuna/test/assert"
)

// Keeping these aliases independent from the editorial content on https://gno.land is
// probably a good idea, as long as underlying realms remain the same.
var (
	miniAliases = map[string]string{
		"/":      "/r/gnoland/home",
		"/about": "/r/gnoland/pages:p/about",
		"/start": "/r/gnoland/pages:p/start",
	}
	miniRedirects = map[string]string{
		"/game-of-realms":  "/r/gnoland/pages:p/gor",
		"/getting-started": "/start",
		"/blog":            "/r/gnoland/blog",
		"/gor":             "/game-of-realms",
	}
)

// Launch a gnoland chain choosing a free random host:port, which is returned.
// Note: Make sure to call `defer node.Stop()`
func launchGnolandNode(t *testing.T) (node *node.Node, remoteAddr string) {
	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	return integration.TestingInMemoryNode(t, log.NewTestingLogger(t), config)
}

// modify a fresh NewDefaultConfig() to use specified remoteAddr
func configWith(remoteAddr string) Config {
	cfg := NewDefaultConfig()
	cfg.RemoteAddr = remoteAddr
	return cfg
}

func TestRoutes(t *testing.T) {
	const (
		ok       = http.StatusOK
		found    = http.StatusFound
		notFound = http.StatusNotFound
	)
	routes := []struct {
		route     string
		status    int
		substring string
	}{
		{"/", ok, "Welcome"}, // assert / gives 200 (OK). assert / contains "Welcome".
		{"/about", ok, "blockchain"},
		{"/r/gnoland/blog", ok, ""}, // whatever content
		{"/r/gnoland/blog?help", ok, "exposed"},
		{"/r/gnoland/blog/", ok, "admin.gno"},
		{"/r/gnoland/blog/admin.gno", ok, "func "},
		{"/r/demo/users:administrator", ok, "address"},
		{"/r/demo/users", ok, "manfred"},
		{"/r/demo/users/users.gno", ok, "// State"},
		{"/r/demo/deep/very/deep", ok, "it works!"},
		{"/r/demo/deep/very/deep:bob", ok, "hi bob"},
		{"/r/demo/deep/very/deep?help", ok, "exposed"},
		{"/r/demo/deep/very/deep/", ok, "render.gno"},
		{"/r/demo/deep/very/deep/render.gno", ok, "func Render("},
		{"/game-of-realms", found, "/r/gnoland/pages:p/gor"},
		{"/gor", found, "/game-of-realms"},
		{"/blog", found, "/r/gnoland/blog"},
		{"/404-not-found", notFound, "/404-not-found"},
		{"/아스키문자가아닌경로", notFound, "/아스키문자가아닌경로"},
		{"/%ED%85%8C%EC%8A%A4%ED%8A%B8", notFound, "/테스트"},
		{"/グノー", notFound, "/グノー"},
		{"/⚛️", notFound, "/⚛️"},
		{"/p/demo/flow/LICENSE", ok, "BSD 3-Clause"},
	}

	gnoland, remoteAddr := launchGnolandNode(t)
	defer gnoland.Stop()
	app := MakeAppWithOptions(log.NewTestingLogger(t), configWith(remoteAddr), Options{
		Aliases:   miniAliases,
		Redirects: miniRedirects,
	})
	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			app.Router.ServeHTTP(response, request)
			assert.Equal(t, r.status, response.Code)
			assert.Contains(t, response.Body.String(), r.substring)
		})
	}
}

func TestAnalytics(t *testing.T) {
	routes := []string{
		// special realms
		"/", // home
		"/about",
		"/start",

		// redirects
		"/game-of-realms",
		"/getting-started",
		"/blog",
		"/boards",

		// realm, source, help page
		"/r/gnoland/blog",
		"/r/gnoland/blog/admin.gno",
		"/r/demo/users:administrator",
		"/r/gnoland/blog?help",

		// special pages
		"/404-not-found",
	}

	gnoland, remoteAddr := launchGnolandNode(t)
	defer gnoland.Stop()
	cfg := configWith(remoteAddr)

	t.Run("with", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				ccfg := cfg // clone config
				ccfg.WithAnalytics = true
				app := MakeAppWithOptions(log.NewTestingLogger(t), ccfg, Options{
					Aliases:   miniAliases,
					Redirects: miniRedirects,
				})
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Contains(t, response.Body.String(), "sa.gno.services")
			})
		}
	})
	t.Run("without", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				ccfg := cfg // clone config
				ccfg.WithAnalytics = false
				app := MakeAppWithOptions(log.NewTestingLogger(t), ccfg, Options{
					Aliases:   miniAliases,
					Redirects: miniRedirects,
				})
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Equal(t, strings.Contains(response.Body.String(), "sa.gno.services"), false)
			})
		}
	})
}
