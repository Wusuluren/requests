package requests

import (
	"fmt"
	"github.com/wusuluren/requests/py"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
)

func defaultUserAgent(name string) string {
	if name == "" {
		name = "golang-requests"
	}
	return fmt.Sprintf("%s/%s", name, __version__)
}

func defaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      defaultUserAgent(""),
		"Accept-Encoding": strings.Join([]string{"gzip", "deflate"}, ""),
		"Accept":          "*/*",
		"Connection":      "keep-alive",
	}
}

func InStrSlice(value string, slice []string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func InStrStrMap(value string, dict map[string]string) bool {
	for _, v := range dict {
		if v == value {
			return true
		}
	}
	return false
}

func defaultValue(value, defaultValue interface{}) interface{} {
	if value != defaultValue {
		return value
	}
	return defaultValue
}

func selectProxy(Url string, proxies py.Dict) string {
	if proxies == nil {
		proxies = make(py.Dict)
	}
	urlparts, _ := url.Parse(Url)
	if urlparts.Hostname() == "" {
		return proxies.Get(urlparts.Scheme, proxies["all"]).(string)
	}

	proxyKeys := []string{
		urlparts.Scheme + "://" + urlparts.Host,
		urlparts.Scheme,
		"all://" + urlparts.Host,
		"all",
	}
	proxy := ""
	for _, proxyKey := range proxyKeys {
		if proxies.Have(proxyKey) {
			proxy = proxies[proxyKey].(string)
			break
		}
	}
	return proxy
}

func prependSchemeIfNeeded(rawurl string, newScheme string) string {
	return rawurl
}

func parseHeaderLinks(value string) []map[string]string {
	links := make([]map[string]string, 0)

	replaceChars := `'"`

	value = strings.Trim(value, replaceChars)
	if value == "" {
		return links
	}

	re := regexp.MustCompile(`, *<`)
	for _, val := range re.Split(value, -1) {
		var Url, params string
		splited := strings.SplitN(val, ";", 1)
		Url = splited[0]
		if len(splited) == 2 {
			params = splited[1]
		}

		link := map[string]string{
			"url": strings.Trim(Url, `<> '"`),
		}

		for _, param := range strings.Split(params, ";") {
			var key, value string
			splited := strings.Split(param, "=")
			if len(splited) != 2 {
				continue
			}
			key = splited[0]
			value = splited[1]

			link[strings.Trim(key, replaceChars)] = strings.Trim(value, replaceChars)
		}

		links = append(links, link)
	}

	return links
}

func parseContentTypeHeader(header string) (string, map[string]string) {
	tokens := strings.Split(header, ";")
	contentType, params := strings.Trim(tokens[0], " "), tokens[1:]
	paramsDict := make(map[string]string)
	itemsToStrip := `"' `

	for _, param := range params {
		param = strings.Trim(param, " ")
		if param != "" {
			key, value := param, ""
			indexOfEquals := strings.Index(param, "=")
			if indexOfEquals != -1 {
				key = strings.Trim(param[:indexOfEquals], itemsToStrip)
				value = strings.Trim(param[indexOfEquals+1:], itemsToStrip)
			}
			paramsDict[key] = value
		}
	}
	return contentType, paramsDict
}

func getEncodingFromHeaders(headers map[string]string) string {
	var contentType string
	if val, ok := headers["content-type"]; ok {
		contentType = val
	}

	if contentType == "" {
		return ""
	}

	contentType, params := parseContentTypeHeader(contentType)

	if val, ok := params["charset"]; ok {
		return strings.Trim(val, `'""'`)
	}

	if _, ok := params["text"]; ok {
		return "ISO-8859-1"
	}
	return ""
}

func checkDict(dict py.Dict) py.Dict {
	if dict == nil {
		dict = py.Dict{}
	}
	return dict
}

func strStartsWith(str, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(str), prefix)
}

func superLen(o io.Reader) int {
	if o == ioutil.NopCloser(nil) {
		return 0
	}
	body, err := ioutil.ReadAll(o)
	if err != nil {
		return 0
	}
	return len(body)
}

func getAuthFromUrl(Url string) []string {
	parsed, err := url.Parse(Url)
	if err != nil {
		panic(err)
	}

	auth := make([]string, 0)
	auth = append(auth, parsed.User.Username())
	passwd, _ := parsed.User.Password()
	auth = append(auth, passwd)

	return auth
}
