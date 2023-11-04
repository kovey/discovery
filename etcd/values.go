package etcd

type Values map[string]string

func (v Values) Add(key, value string) {
	v[key] = value
}

func (v Values) Get(key string) string {
	if val, ok := v[key]; ok {
		return val
	}

	return ""
}
