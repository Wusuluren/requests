package requests

import (
	"fmt"
	"github.com/wusuluren/requests/py"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func mergeSetting(requestSetting, sessionSetting interface{}) interface{} {
	if requestSetting == nil {
		return sessionSetting
	}

	if sessionSetting == nil {
		return requestSetting
	}

	var b1, b2 bool
	if _, ok := sessionSetting.(py.Dict); ok {
		b1 = true
	}
	if _, ok := requestSetting.(py.Dict); ok {
		b2 = true
	}
	if !(b1 && b2) {
		return requestSetting
	}

	mergedSetting := sessionSetting.(py.Dict)
	mergedSetting.Update(requestSetting.(py.Dict))

	noneKeys := make([]string, 0)
	for k, v := range mergedSetting {
		if v == nil {
			noneKeys = append(noneKeys, k)
		}
	}
	for _, key := range noneKeys {
		delete(mergedSetting, key)
	}

	return mergedSetting
}

func mergeHooks(requestHooks, sessionHooks map[string][]hookFunc) interface{} {
	resps, ok := sessionHooks["response"]
	respNotExist := true
	if ok {
		respNotExist = len(resps) == 0
	}
	if sessionHooks == nil || respNotExist {
		return requestHooks
	}

	resps, ok = requestHooks["response"]
	respNotExist = true
	if ok {
		respNotExist = len(resps) == 0
	}
	if requestHooks == nil || respNotExist {
		return sessionHooks
	}

	return mergeSetting(requestHooks, sessionHooks)
}

type sessionRedirectMixin struct {
}

func (s *sessionRedirectMixin) getRedirectTarget(resp Response) {

}

func (s *sessionRedirectMixin) resolveRedirects(resp Response, req *Request, kwargs py.Dict) {
	//hist := make([]Response, 0)
	//
	//Url := s.getRedirectTarget(resp)
	//parsedUrl,_ := url.Parse(req.url)
	//previousFragment := parsedUrl.Fragment
	//for Url != "" {
	//	preparedRequest := *req
	//
	//	hist = append(hist, resp)
	//	resp.History = hist[1:]
	//
	//	if len(resp.History) >= defaultRedirectLimit {
	//		panic(fmt.Sprintf("Exceeded %s redirects.", defaultRedirectLimit))
	//	}
	//
	//	resp.Body.Close()
	//
	//	if strings.HasPrefix(Url, "//") {
	//		parsedUrl := url.Parse(resp.url)
	//		Url = fmt.Sprintf("%s:%s", parsedUrl.Scheme, Url)
	//	}
	//
	//	parsed,_ := url.Parse(Url)
	//	if parsed.Fragment == "" && previousFragment != ""{
	//		parsed.Fragment = previousFragment
	//	} else if parsed.Fragment != "" {
	//		previousFragment = parsed.Fragment
	//	}
	//	Url = parsed.String()
	//
	//	if parsed.Hostname() != "" {
	//		Url = resp.
	//	}
	//}
}

func (s *sessionRedirectMixin) rebuildAuth() {

}

func (s *sessionRedirectMixin) rebuildProxies() {

}

func (s *sessionRedirectMixin) rebuildMethod() {

}

// session
type session struct {
	sessionRedirectMixin
	headers      map[string]string
	cookies      CookieJar
	auth         []string
	proxies      map[string]string
	hooks        map[string][]hookFunc
	params       py.Dict
	verify       bool
	cert         []string
	prefetch     bool
	adapters     AdapterDict
	stream       bool
	trustEnv     bool
	maxRedirects int
}

func (s *session) init() {
	s.headers = defaultHeaders()

	s.auth = make([]string, 0)

	s.proxies = make(map[string]string)

	s.hooks = defaultHooks()

	s.params = make(py.Dict)

	s.stream = false

	s.verify = false

	s.cert = make([]string, 0)

	s.maxRedirects = defaultRedirectLimit

	s.trustEnv = true

	s.cookies = cookiejarFromDict(make(map[string]string), nil, false)

	s.adapters = AdapterDict{}

	s.mount("https://", &HttpAdapterImpl{})
	s.mount("http://", &HttpAdapterImpl{})
}

func (s *session) prepareRequest(req *Request) *PreparedRequest {
	cookies := CookieJar{}

	//todo
	mergedCookies := cookies

	p := &PreparedRequest{Request: *req}
	p.init()
	p.prepare(py.Dict{
		"method":  strings.ToUpper(req.method),
		"url":     req.url,
		"files":   req.files,
		"data":    req.data,
		"json":    req.json,
		"headers": mergeSetting(req.headers, p.headers),
		"params":  mergeSetting(req.params, p.params),
		"auth":    mergeSetting(req.auth, p.auth),
		"cookies": mergedCookies,
		"hooks":   mergeHooks(req.hooks, p.hooks),
	})
	return p
}

func (s *session) request(method, Url string, dftargs py.Dict) *Response {
	params := dftargs.Get("params", py.Dict{}).(py.Dict)
	data := dftargs.Get("data", io.ReadCloser(ioutil.NopCloser(nil))).(io.ReadCloser)
	headers := dftargs.Get("headers", make(map[string]string)).(map[string]string)
	cookies := dftargs.Get("cookies", CookieJar{}).(CookieJar)
	files := dftargs.Get("files", make(map[string]string)).(map[string]string)
	auth := dftargs.Get("auth", make([]string, 0)).([]string)
	timeout := dftargs.Get("timeout", time.Duration(0)).(time.Duration)
	allowRedirects := dftargs.Get("allowRedirects", true).(bool)
	proxies := dftargs.Get("proxies", py.Dict{}).(py.Dict)
	hooks := dftargs.Get("hooks", make(map[string][]hookFunc)).(map[string][]hookFunc)
	stream := dftargs.Get("stream", false).(bool)
	verify := dftargs.Get("verify", "").(string)
	cert := dftargs.Get("cert", make([]string, 0)).([]string)
	json := dftargs.Get("json", nil)

	req := &Request{
		Request: &http.Request{},
		method:  method,
		url:     Url,
		headers: headers,
		files:   files,
		data:    data,
		json:    json,
		params:  params,
		auth:    auth,
		cookies: cookies,
		hooks:   hooks,
	}
	prep := s.prepareRequest(req)

	settings := s.mergeEnvironmentSettings(prep.url, proxies, stream, verify, cert)

	sendKwargs := py.Dict{
		"timeout":        timeout,
		"allowRedirects": allowRedirects,
	}
	sendKwargs.Update(settings)
	resp := s.send(prep, sendKwargs)
	return resp
}

func (s *session) Get(url string, kwargs py.Dict) *Response {
	kwargs.SetDefault("allow_redirects", true)
	return s.request("get", url, kwargs)
}

func (s *session) Option(url string, kwargs py.Dict) *Response {
	kwargs.SetDefault("allow_redirects", true)
	return s.request("option", url, kwargs)
}

func (s *session) Head(url string, kwargs py.Dict) *Response {
	kwargs.SetDefault("allow_redirects", false)
	return s.request("head", url, kwargs)
}

func (s *session) Post(url string, kwargs py.Dict) *Response {
	return s.request("post", url, kwargs)
}

func (s *session) Put(url string, kwargs py.Dict) *Response {
	return s.request("put", url, kwargs)
}

func (s *session) Patch(url string, kwargs py.Dict) *Response {
	return s.request("patch", url, kwargs)
}

func (s *session) Delete(url string, kwargs py.Dict) *Response {
	return s.request("delete", url, kwargs)
}

func (s *session) send(request *PreparedRequest, kwargs py.Dict) *Response {
	kwargs.SetDefault("stream", s.stream)
	kwargs.SetDefault("verify", s.verify)
	kwargs.SetDefault("cert", s.cert)
	kwargs.SetDefault("proxies", s.proxies)

	//allowRedirects := kwargs.Pop("allowRedirects", true)
	//stream := kwargs["stream"]
	hooks := request.hooks

	adapter := s.getAdapter(request.url)

	start := time.Now()

	r := adapter.send(request, kwargs)

	elapsed := time.Now().Sub(start)
	r.elapsed = time.Duration(elapsed)

	r = dispatchHook("response", hooks, r, kwargs)

	if len(r.history) > 0 {
		//for _, resp := range r.History {
		//extractCookiesToJar(s.cookies, resp.Request, resp.Raw)
		//}
	}
	//extractCookiesToJar(s.cookies, resp.Request, resp.Raw)

	//gen := s.resolveRedirects(r, request, kwargs)

	return r
}

func (s *session) mergeEnvironmentSettings(Url string, proxies py.Dict, stream bool, verify string, cert []string) py.Dict {
	//fixme ???

	proxies = mergeSetting(proxies, s.proxies).(py.Dict)
	stream = mergeSetting(stream, s.stream).(bool)
	verify = mergeSetting(verify, s.verify).(string)
	cert = mergeSetting(cert, s.cert).([]string)

	return py.Dict{
		"verify":  verify,
		"proxies": proxies,
		"stream":  stream,
		"cert":    cert,
	}
}

func (s *session) getAdapter(url string) HttpAdapter {
	for prefix, adapter := range s.adapters {
		if strings.HasPrefix(strings.ToLower(url), strings.ToLower(prefix)) {
			return adapter
		}
	}
	panic(fmt.Sprintf("No connection adapters were found for %s", url))
	return nil
}

func (s *session) close() {

}

func (s *session) mount(prefix string, adapter HttpAdapter) {
	adapter.init(nil)
	s.adapters[prefix] = adapter
	keysToMove := make([]string, 0)
	for k := range s.adapters {
		if len(k) < len(prefix) {
			keysToMove = append(keysToMove, k)
		}
	}

	for _, key := range keysToMove {
		s.adapters[key] = s.adapters.pop(key)
	}
}

func Session() *session {
	s := &session{}
	s.init()
	return s
}
