package slices

func AllZeros(s []byte) bool {
	for i := range s {
		if s[i] != 0 {
			return false
		}
	}
	return true
}
