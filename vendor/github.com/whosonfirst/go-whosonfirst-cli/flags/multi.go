package flags

import (
	"strconv"
	"strings"
)

type MultiString []string

func (m *MultiString) String() string {
	return strings.Join(*m, "\n")
}

func (m *MultiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func (m *MultiString) Contains(value string) bool {

	for _, test := range *m {

		if test == value {
			return true
		}
	}

	return false
}

type MultiInt64 []int64

func (m *MultiInt64) String() string {

	str_values := make([]string, len(*m))

	for i, v := range *m {
		str_values[i] = strconv.FormatInt(v, 10)
	}

	return strings.Join(str_values, "\n")
}

func (m *MultiInt64) Set(str_value string) error {

	value, err := strconv.ParseInt(str_value, 10, 64)

	if err != nil {
		return err
	}

	*m = append(*m, value)
	return nil
}

func (m *MultiInt64) Contains(value int64) bool {

	for _, test := range *m {

		if test == value {
			return true
		}
	}

	return false
}
