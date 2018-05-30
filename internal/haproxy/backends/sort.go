package backends

import (
	"sort"
	"strings"
)

type sortBackends []Backend

func (s sortBackends) Len() int {
	return len(s)
}

func (s sortBackends) Less(i, j int) bool {
	return order(s[j]) < order(s[i])
}

func (s sortBackends) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sorted list of backends represented as a slice.
func (b Backends) Sorted() []Backend {
	var toSort sortBackends

	for _, backend := range b {
		toSort = append(toSort, backend)
	}

	sort.Sort(toSort)

	return toSort
}

func order(bck Backend) int {
	// If this is a root path, it should be at the bottom of the list.
	if bck.Path == "/" {
		return 0
	}

	return len(strings.Split(bck.Path, "/"))
}
