package requests

import (
	"net/http"
)

// AdapterDict
type AdapterDict map[string]HttpAdapter

func (a AdapterDict) pop(key string) HttpAdapter {
	value := a[key]
	delete(a, key)
	return value
}

// CookieJar
type CookieJar map[string]*http.Cookie

func (c CookieJar) SetCookie(cookie *http.Cookie) {
	c[cookie.Name] = cookie
}
