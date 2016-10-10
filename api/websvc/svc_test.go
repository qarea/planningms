package websvc

import (
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"

	"../../cfg"
)

type TestSuite struct {
	url string
	w   *httptest.ResponseRecorder
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {
	s.url = "http://websvc.test" + cfg.HTTP.BasePath + "/web"
}

func (s *TestSuite) SetUpTest(c *C) {
	s.w = httptest.NewRecorder()
}

func (s *TestSuite) Test404(c *C) {
	cases := []struct {
		url string
	}{
		{s.url + "x"},
		{s.url + "/"},
		{s.url + "/x"},
	}
	for _, v := range cases {
		s.w = httptest.NewRecorder()
		r, err := http.NewRequest("GET", v.url, nil)
		c.Assert(err, IsNil)
		r.RemoteAddr = "1.2.3.4:0"
		r.Header.Set(cfg.HTTP.RealIPHeader, "1.2.3.4")
		http.DefaultServeMux.ServeHTTP(s.w, r)
		c.Check(s.w.Code, Equals, 404)
		c.Check(s.w.Header().Get("Content-Type"), Equals, "text/plain; charset=utf-8")
		c.Check(s.w.Body.String(), Equals, "404 page not found\n", Commentf("url=%q", v.url))
	}
}
