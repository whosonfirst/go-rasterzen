package cache

import (
	"fmt"
)

type CacheMiss struct {
	error string
}

func (m CacheMiss) Error() string {
	return fmt.Sprintf("CACHE MISS %s", m.error)
}

func IsCacheMiss(e error) bool {

	switch e.(type) {
	case *CacheMiss:
		return true
	case CacheMiss:
		return true
	default:
		// pass
	}

	return false
}

type CacheMissMulti struct {
	error string
}

func (m CacheMissMulti) Error() string {

	return fmt.Sprintf("ONE OR MORE MULTI CACHE MISSES %s", m.error)
}

func IsCacheMissMulti(e error) bool {

	switch e.(type) {
	case *CacheMissMulti:
		return true
	case CacheMissMulti:
		return true
	default:
		// pass
	}

	return false
}
