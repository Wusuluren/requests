package requests

import (
	"net/http"
	"net/url"
	"time"
)

// mockRequest
type mockRequest struct {
	r          *Request
	newHeaders map[string]string
	type_      string
}

func (m *mockRequest) init(request *Request) {
	m.r = request
	m.newHeaders = make(map[string]string)
	url, _ := url.Parse(m.r.url)
	m.type_ = url.Scheme
}

func (m *mockRequest) getType() string {
	return m.type_
}

func (m *mockRequest) getHost() string {
	url, _ := url.Parse(m.r.url)
	return url.Host
}

func (m *mockRequest) getOriginReqHost() string {
	return m.getHost()
}

func (m *mockRequest) getFullUrl() string {
	if _, ok := m.r.headers["Host"]; !ok {
		return m.r.url
	}
	host := m.r.headers["Host"]
	parsed, _ := url.Parse(m.r.url)
	parsed.Host = host
	return parsed.String()
}

func (m *mockRequest) isUnverifiable() bool {
	return true
}

func (m *mockRequest) hasHeader(name string) bool {
	has := InStrStrMap(name, m.r.headers)
	has2 := InStrStrMap(name, m.newHeaders)
	return has || has2
}

func (m *mockRequest) getHeader(key, dft string) string {
	if val, ok := m.r.headers[key]; ok {
		return val
	}
	if val, ok := m.newHeaders[key]; ok {
		return val
	}
	return dft
}

func (m *mockRequest) addHeader(key, val string) {
	panic("Cookie headers should be added with add_unredirected_header()")
}

func (m *mockRequest) addUnredirectedHeader(name, value string) {
	m.newHeaders[name] = value
}

func (m *mockRequest) getNewHeaders() map[string]string {
	return m.newHeaders
}

func (m *mockRequest) unverifiable() bool {
	return m.isUnverifiable()
}
func (m *mockRequest) originReqHost() string {
	return m.getOriginReqHost()
}
func (m *mockRequest) host() string {
	return m.getHost()
}

// mockResponse
type mockResponse struct {
	headers map[string]string
}

func (m *mockResponse) init(headers map[string]string) {
	m.headers = headers
}

func (m *mockResponse) info() map[string]string {
	return m.headers
}

func (m *mockResponse) getheaders(name string) string {
	if val, ok := m.headers[name]; ok {
		return val
	}
	return ""
}

func extractCookiesToJar(jar CookieJar, request *PreparedRequest, response *http.Response) {
	//req := mockRequest{}
	//req.init(request)
	//res := mockResponse{}
	//res.init(response.originalResponse.msg)
}

func createCookie(name, value string) *http.Cookie {
	result := &http.Cookie{
		Name:    name,
		Value:   value,
		Path:    "/",
		Domain:  "",
		Secure:  false,
		Expires: time.Time{},
	}

	return result
}

func cookiejarFromDict(cookieDict map[string]string, cookiejar CookieJar, overwrite bool) CookieJar {
	if cookiejar == nil {
		cookiejar = make(CookieJar, 0)
	}

	if len(cookieDict) > 0 {
		namesFromJar := make([]string, 0)
		for _, cookie := range cookiejar {
			namesFromJar = append(namesFromJar, cookie.Name)
		}
		for name := range cookieDict {
			if overwrite || !InStrSlice(name, namesFromJar) {
				cookiejar.SetCookie(createCookie(name, cookieDict[name]))
			}
		}
	}

	return cookiejar
}

func getCookieHeader(jar CookieJar, request *Request) {
	//r := mockRequest{r:request}
	//jar.
}
