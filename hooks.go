package requests

import "github.com/wusuluren/requests/py"

type hookFunc func(*Response, py.Dict) *Response

var HOOKS = []string{"response"}

func defaultHooks() map[string][]hookFunc {
	hooksDict := make(map[string][]hookFunc)
	for _, event := range HOOKS {
		hooksDict[event] = make([]hookFunc, 0)
	}
	return hooksDict
}

func dispatchHook(key string, hooks map[string][]hookFunc, hookData *Response, kwargs py.Dict) *Response {
	if hooks == nil {
		hooks = make(map[string][]hookFunc)
	}
	if hookValue, ok := hooks[key]; ok {
		hooks := hookValue
		for _, hook := range hooks {
			_hookData := hook(hookData, kwargs)
			if _hookData != nil {
				hookData = _hookData
			}
		}
	}
	return hookData
}
