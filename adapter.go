package requests

import (
	"github.com/wusuluren/requests/py"
	"github.com/wusuluren/requests/urllib3"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultPoolBlock   = false
	defaultPoolSize    = 10
	defaultRetries     = 0
	defaultPoolTimeout = 0
)

type BaseAdapter interface {
	init()
	send()
	close()
}

type BaseAdapterImpl struct {
}

func (a *BaseAdapterImpl) init() {

}

func (a *BaseAdapterImpl) send() {

}

func (a *BaseAdapterImpl) close() {

}

type HttpAdapter interface {
	init(dftargs py.Dict)
	send(request *PreparedRequest, kwargs py.Dict) *Response
	close()
}

type HttpAdapterImpl struct {
	maxRetries      int
	config          map[string]string
	poolConnections int
	poolMaxSize     int
	poolBlock       bool

	proxyManager map[string]*urllib3.PoolManager
	poolManager  *urllib3.PoolManager
}

func (a *HttpAdapterImpl) init(dftargs py.Dict) {
	poolConnections := dftargs.Get("poolConnections", defaultPoolSize).(int)
	poolMaxSize := dftargs.Get("poolMaxSize", defaultPoolSize).(int)
	maxRetries := dftargs.Get("maxRetries", defaultRetries).(int)
	poolBlock := dftargs.Get("poolBlock", defaultPoolBlock).(bool)

	if maxRetries == defaultRetries {
		//a.maxRetries = Retry(0, false) //Fixme
	} else {
		a.maxRetries = 0 //fixme
	}
	a.config = make(map[string]string)
	a.proxyManager = make(map[string]*urllib3.PoolManager)

	a.poolConnections = poolConnections
	a.poolMaxSize = poolMaxSize
	a.poolBlock = poolBlock

	a.initPoolManager(poolConnections, poolMaxSize, poolBlock)
}

func (a *HttpAdapterImpl) initPoolManager(poolConnections int, poolMaxSize int, poolBlock bool) {
	poolBlock = defaultValue(poolBlock, defaultPoolBlock).(bool)

	a.poolConnections = poolConnections
	a.poolMaxSize = poolMaxSize
	a.poolBlock = poolBlock

	a.poolManager = &urllib3.PoolManager{}
	a.poolManager.Init()
}

func (a *HttpAdapterImpl) proxyManagerFor(proxy string, proxyKwargs py.Dict) *urllib3.PoolManager {
	var manager *urllib3.PoolManager
	if _, ok := a.proxyManager[proxy]; ok {
		manager = a.proxyManager[proxy]
	} else if strings.HasPrefix(strings.ToLower(proxy), "socks") {
		panic("not implement")
	} else {
		panic("not implement")
	}
	return manager
}

func (a *HttpAdapterImpl) certVerify(conn *urllib3.Conn, Url string, verify string, cert []string) {
	if strings.HasPrefix(strings.ToLower(Url), "https") && verify != "" {

	} else {

	}

	if len(cert) > 0 {

	}
}

func (a *HttpAdapterImpl) buildResponse(req *PreparedRequest, resp *http.Response) *Response {
	response := &Response{}
	response.init(resp)

	response.encoding = getEncodingFromHeaders(response.headers)

	extractCookiesToJar(response.cookies, req, resp)

	response.request = &req.Request
	return response
}

func (a *HttpAdapterImpl) getConnection(Url string, dftargs py.Dict) *urllib3.Conn {
	proxies := dftargs.Get("proxies", make(py.Dict)).(py.Dict)

	var conn *urllib3.Conn
	proxy := selectProxy(Url, proxies)
	if proxy != "" {
		proxy = prependSchemeIfNeeded(proxy, "http")
		proxyUrl, _ := url.Parse(proxy)
		if proxyUrl.Host == "" {
			panic("Please check proxy URL. It is malformed and could be missing the host.")
		}
		proxyManager := a.proxyManagerFor(proxy, nil)
		conn = proxyManager.ConnectionFromUrl(Url)
	} else {
		parsed, _ := url.Parse(Url)
		Url = parsed.RequestURI() //fixme:???
		conn = a.poolManager.ConnectionFromUrl(Url)
	}

	return conn
}

func (a *HttpAdapterImpl) close() {

}

func (a *HttpAdapterImpl) requestUrl(request *PreparedRequest, proxies py.Dict) string {
	proxy := selectProxy(request.url, proxies)
	parsedUrl, _ := url.Parse(request.url)
	scheme := parsedUrl.Scheme

	isProxiedHttpRequest := proxy != "" && scheme != "https"
	usingSocksProxy := false
	if proxy != "" {
		parsedUrl, _ = url.Parse(proxy)
		proxyScheme := strings.ToLower(parsedUrl.Scheme)
		usingSocksProxy = strings.HasPrefix(proxyScheme, "socks")
	}

	//Url := request.URL.Path
	//Url := parsedUrl.Path
	Url := request.url //fixme ???
	if isProxiedHttpRequest && !usingSocksProxy {
		Url = request.url //fixme ???
	}

	//println("!!!", Url)
	return Url
}

func (a *HttpAdapterImpl) addHeaders(request *PreparedRequest, dftargs py.Dict) {

}

func (a *HttpAdapterImpl) proxyHeaders() {

}

func (a *HttpAdapterImpl) send(request *PreparedRequest, dftargs py.Dict) *Response {
	stream := dftargs.Get("stream", false).(bool)
	timeout := dftargs.Get("timeout", time.Duration(0)).(time.Duration)
	verify := dftargs.Get("verify", "").(string)
	cert := dftargs.Get("cert", make([]string, 0)).([]string)
	proxies := dftargs.Get("proxies", py.Dict{}).(py.Dict)

	conn := a.getConnection(request.url, proxies)

	a.certVerify(conn, request.url, verify, cert)
	Url := a.requestUrl(request, proxies)
	_ = Url
	a.addHeaders(request, py.Dict{
		"stream":  stream,
		"timeout": timeout,
		"verify":  verify,
		"cert":    cert,
		"proxies": proxies,
	})

	var b1, b2 bool
	if request.Body != ioutil.NopCloser(nil) {
		b1 = true
	}
	if _, ok := request.headers["Content-Length"]; ok {
		b2 = true
	}
	chunked := !(b1 || b2)

	var resp *http.Response
	var err error
	_ = chunked
	if !chunked {
		//if true {
		resp, err = conn.Urlopen(request.Request.Request)
	} else {
		//if conn.proxyPool {
		//	conn = conn.proxyPool
		//}

		//lowConn := conn.GetConn()
		panic("unimplement")
	}
	if err != nil {
		panic(err)
	}

	return a.buildResponse(request, resp)
}
