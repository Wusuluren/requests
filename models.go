package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wusuluren/requests/py"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var redirectStati = []int{
	http.StatusMovedPermanently,  //301
	http.StatusFound,             //302
	http.StatusSeeOther,          //303
	http.StatusTemporaryRedirect, //307
	http.StatusPermanentRedirect, //308
}

const (
	defaultRedirectLimit = 30
	contentChunkSize     = 10 * 1024
	IterChunkSize        = 512
)

// Request
type Request struct {
	*http.Request

	method         string
	url            string
	params         py.Dict
	data           io.ReadCloser
	json           interface{}
	headers        map[string]string
	cookies        CookieJar
	files          map[string]string
	auth           []string
	timeout        time.Duration //float64
	allowRedirects bool
	proxies        map[string]string
	verify         string
	stream         bool
	cert           []string
	hooks          map[string][]hookFunc
}

func (r *Request) init() {

}
func (r *Request) String() {

}

func (r *Request) prepare(dftargs py.Dict) *PreparedRequest {
	p := &PreparedRequest{Request: *r}
	p.init()
	p.prepare(dftargs)
	return p
}

// PreparedRequest
type PreparedRequest struct {
	Request
	_cookies CookieJar
}

func (r *PreparedRequest) init() {
	r.hooks = defaultHooks()
}

func (r *PreparedRequest) prepare(dftargs py.Dict) {
	method := dftargs.Get("method", "").(string)
	Url := dftargs.Get("url", "").(string)
	headers := dftargs.Get("headers", make(map[string]string)).(map[string]string)
	files := dftargs.Get("files", make(map[string]string)).(map[string]string)
	data := dftargs.Get("data", io.ReadCloser(nil)).(io.ReadCloser)
	params := dftargs.Get("params", py.Dict{}).(py.Dict)
	auth := dftargs.Get("auth", make([]string, 0)).([]string)
	cookies := dftargs.Get("cookies", CookieJar{}).(CookieJar)
	hooks := dftargs.Get("hooks", make(map[string][]hookFunc)).(map[string][]hookFunc)
	json := dftargs.Get("json", nil)

	r.prepareMethod(method)
	r.prepareUrl(Url, params)
	r.prepareHeaders(headers)
	r.prepareCookies(cookies)
	r.prepareBody(data, files, json)
	r.prepareAuth(auth, Url)

	r.prepareHooks(hooks)
}

func (r *PreparedRequest) String() string {
	return fmt.Sprintf("<PreparedRequest [%s]>", r.method)
}
func (r *PreparedRequest) copy() {
	p := &PreparedRequest{}
	p.method = r.method
	p.URL = r.URL
	if r.headers == nil {
		p.headers = nil
	} else {
		p.headers = r.headers
	}
}
func (r *PreparedRequest) prepareMethod(method string) {
	r.method = method
	r.method = strings.ToUpper(r.method)
}
func (r *PreparedRequest) prepareUrl(Url string, params py.Dict) {
	Url = strings.TrimSpace(Url)

	if strings.Contains(Url, ":") && !strStartsWith(Url, "http") {
		r.url = Url
		return
	}

	URL, err := url.Parse(Url)
	if err != nil {
		panic(err)
	}

	if URL.Scheme == "" {
		panic(fmt.Sprintf("Invalid URL %s: No schema supplied. Perhaps you meant http://%s?", Url, Url))
	}

	if URL.Host == "" {
		panic(fmt.Sprintf("Invalid URL %s: No host supplied", Url))
	}

	if URL.Path == "" {
		URL.Path = "/"
	}

	r.URL = URL
}
func (r *PreparedRequest) prepareHeaders(headers map[string]string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	for key, val := range headers {
		r.headers[key] = val
	}
}
func (r *PreparedRequest) prepareBody(data io.ReadCloser, files map[string]string, Json interface{}) {
	var err error
	var body []byte
	var contentType string

	if data == ioutil.NopCloser(nil) && Json != nil {
		contentType = "application/json"
		body, err = json.Marshal(Json)
		if err != nil {
			panic(err)
		}
	}

	checkReader := func(reader io.ReadCloser) (bool, []byte) {
		if reader == ioutil.NopCloser(nil) {
			return false, make([]byte, 0)
		}
		body, err = ioutil.ReadAll(reader)
		if err != nil {
			return false, make([]byte, 0)
		}
		return true, body
	}

	isStream, _ := checkReader(data)

	length := superLen(data)

	var w *multipart.Writer
	var multipart bytes.Buffer
	if isStream {
		if len(files) > 0 {
			panic("Streamed bodies and files are mutually exclusive.")
		}

		if length != 0 {
			r.headers["Content-Length"] = strconv.Itoa(length)
		} else {
			r.headers["Transfer-Encoding"] = "chunked"
		}
	} else {
		if len(files) > 0 {
			w, multipart, err = r.createMultiPart(files)
			if err != nil {
				panic(err)
			}
		} else {
			if data != ioutil.NopCloser(nil) {
				body, err = ioutil.ReadAll(data)
				if err != nil {
					panic(err)
				}
				contentType = "application/x-www-form-urlencoded"
			}
		}

		r.prepareContentLength(body)

		if contentType != "" {
			if _, ok := r.headers["Content-Type"]; !ok {
				r.headers["Content-Type"] = contentType
			}
		}
	}

	if len(files) > 0 {
		r.Request.Request, err = http.NewRequest(r.method, r.url, &multipart)
		if err != nil {
			panic(err)
		}
		r.Request.Request.Header.Set("Content-Type", w.FormDataContentType())
		w.Close()
	} else {
		r.Request.Request, err = http.NewRequest(r.method, r.url, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}
	}
	//b, err :=ioutil.ReadAll(r.Request.Request.Body)
	//fmt.Println("222 ", err,  string(b), len(b))
}

func (r *PreparedRequest) createMultiPart(files map[string]string) (w *multipart.Writer, b bytes.Buffer, err error) {
	var fw io.Writer
	w = multipart.NewWriter(&b)
	for fieldname, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			return w, b, err
		}
		fw, err = w.CreateFormFile(fieldname, filename)
		if err != nil {
			return w, b, err
		}
		_, err = io.Copy(fw, f)
		if err != nil {
			return w, b, err
		}
		fw, err = w.CreateFormField(filename)
		if err != nil {
			return w, b, err
		}
		f.Close()
		break
	}
	return w, b, err
}

func (r *PreparedRequest) prepareContentLength(body []byte) {
	if body != nil {
		length := len(body)
		if length > 0 {
			r.headers["Content-Length"] = strconv.Itoa(length)
		}
	} else {
		var b1, b2 bool
		if !InStrSlice(r.method, []string{"GET", "HEAD"}) {
			b1 = true
		}
		if _, ok := r.headers["Content-Length"]; !ok {
			b2 = true
		}
		if b1 && b2 {
			r.headers["Content-Length"] = "0"
		}
	}
}
func (r *PreparedRequest) prepareAuth(auth []string, Url string) {
	//if len(auth) == 0 {
	//	urlAuth := getAuthFromUrl(r.url)
	//	auth = urlAuth
	//}
	//
	//if len(auth) > 0 {
	//	httpAuth := &HTTPBasicAuth{}
	//	if len(auth) == 2 {
	//		httpAuth.init(auth)
	//	}
	//
	//	r = httpAuth(r)
	//	r.update()
	//	body, err := ioutil.ReadAll(r.Body)
	//	if err != nil {
	//		panic(err)
	//	}
	//	r.prepareContentLength(body)
	//}
}
func (r *PreparedRequest) prepareCookies(cookies CookieJar) {
	//if cookies != nil {
	//	r._cookies = cookies
	//} else {
	//	r._cookies = cookiejarFromDict(cookies)
	//}
	//
	//cookieHeader := getCookieHeader(r._cookies, r)
	//if cookieHeader != nil {
	//	r.headers["Cookie"] = cookieHeader
	//}
}
func (r *PreparedRequest) prepareHooks(hooks map[string][]hookFunc) {

}

// Response
type Response struct {
	*http.Response

	statusCode int
	headers    map[string]string
	raw        []byte
	url        string
	encoding   string
	history    []*Response
	cookies    CookieJar
	elapsed    time.Duration
	request    *Request
}

func (r *Response) init(httpResp *http.Response) {
	r.Response = httpResp
	r.cookies = CookieJar{}
	r.headers = make(map[string]string)
	r.raw = make([]byte, 0)
	r.history = make([]*Response, 0)

	r.headers = make(map[string]string)
	//for key, vals := range httpResp.Header {
	//	r.headers[key] = vals[0]
	//}
}

func (r *Response) Ok() bool {
	if r.raiseForStatus() != "" {
		return false
	}
	return true
}

func (r *Response) IsRedirect() bool {
	var b1, b2 bool
	if _, ok := r.headers["location"]; ok {
		b1 = true
	}
	for _, code := range redirectStati {
		if r.statusCode == code {
			b2 = true
			break
		}
	}
	return b1 && b2
}

func (r *Response) Content() []byte {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	return raw
}

func (r *Response) Text() string {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func (r *Response) Json(v interface{}) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(raw, v)
}

func (r *Response) Links() map[string]map[string]string {
	header := ""
	if _, ok := r.headers["link"]; ok {
		header = r.headers["link"]
	}

	l := make(map[string]map[string]string)
	if header != "" {
		links := parseHeaderLinks(header)

		for _, link := range links {
			var key string
			if val, ok := link["rel"]; ok {
				key = val
			} else if val, ok := link["url"]; ok {
				key = val
			}
			l[key] = link
		}
	}

	return l
}

func (r *Response) raiseForStatus() string {
	httpErrorMsg := ""
	reason := "" //fixme ???
	if r.statusCode >= 400 && r.statusCode < 500 {
		httpErrorMsg = fmt.Sprintf("%s Client Error: %s for url: %s", r.statusCode, reason, r.url)
	} else if r.statusCode >= 500 && r.statusCode < 600 {
		httpErrorMsg = fmt.Sprintf("%s Server Error: %s for url: %s", r.statusCode, reason, r.url)
	}
	return httpErrorMsg
}

func (r *Response) Close() {
	r.Response.Body.Close()
}
