package tools

func EnvToPairs(envmap map[string]string) []string {
	environ := make([]string, len(envmap))
	for key, val := range envmap {
		environ = append(environ, key+"="+val)
	}
	return environ
}
