package requests

import (
	"github.com/wusuluren/requests/py"
)

func request(method, url string, params py.Dict) *Response {
	s := &session{}
	s.init()
	defer s.close()
	resp := s.request(method, url, params)
	return resp
}

func Get(url string, params py.Dict) *Response {
	params = checkDict(params)
	params.SetDefault("allowRedirects", true)
	return request("get", url, params)
}

func Option(url string, params py.Dict) *Response {
	params = checkDict(params)
	params.SetDefault("allowRedirects", true)
	return request("option", url, params)
}

func Head(url string, params py.Dict) *Response {
	params = checkDict(params)
	params.SetDefault("allowRedirects", false)
	return request("head", url, params)
}

func Post(url string, params py.Dict) *Response {
	params = checkDict(params)
	return request("post", url, params)
}

func Put(url string, params py.Dict) *Response {
	params = checkDict(params)
	return request("put", url, params)
}

func Patch(url string, params py.Dict) *Response {
	params = checkDict(params)
	return request("patch", url, params)
}

func Delete(url string, params py.Dict) *Response {
	params = checkDict(params)
	return request("delete", url, params)
}
