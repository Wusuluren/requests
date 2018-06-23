package py

// Python-like dictionary
type Dict map[string]interface{}

func (d Dict) Have(key string) bool {
	if _, ok := d[key]; ok {
		return true
	}
	return false
}

func (d Dict) safeSet(key string, value interface{}) {
	if !d.Have(key) {
		d.Set(key, value)
	}
}

func (d Dict) Set(k string, v interface{}) {
	d[k] = v
}

func (d Dict) SetDefault(k string, v interface{}) {
	d.safeSet(k, v)
}

func (d Dict) SetMany(dict Dict) {
	for key, value := range dict {
		d.Set(key, value)
	}
}

func (d Dict) SetDefaultMany(dict Dict) {
	for key, value := range dict {
		d.safeSet(key, value)
	}
}

func (d Dict) Get(k string, defaultValue interface{}) interface{} {
	if v, ok := d[k]; ok {
		return v
	}
	return defaultValue
}

func (d Dict) Pop(k string, defaultValue interface{}) interface{} {
	if v, ok := d[k]; ok {
		delete(d, k)
		return v
	}
	return defaultValue
}

func (d Dict) Update(dict Dict) {
	for key, value := range dict {
		d.Set(key, value)
	}
}
