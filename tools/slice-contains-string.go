package tools

// SliceContainsString tells whether a contains x.
func SliceContainsString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
