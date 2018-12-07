package crdt

func BinarySearch(n int, f func(int) int) (int, bool) {
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1)
		r := f(h)
		switch {
		case r < 0:
			i = h + 1
		case r > 0:
			j = h - 1
		default:
			return h, true
		}
	}
	return i, false
}
