package wfe

import (
	"fmt"
)

func parseInt(s string, d int) (int, error) {
	var v int
	if s != "" {
		if n, err := fmt.Sscanf(s, "%d", &v); err != nil || n != 1 {
			return 0, fmt.Errorf("invalid int value")
		}
	} else {
		v = d
	}

	return v, nil
}

func StringResult(r interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if x, ok := r.(string); ok {
		return x, err
	}
	return "", fmt.Errorf("not string")
}

func StringListResult(r interface{}, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}

	if x, ok := r.([]string); ok {
		return x, err
	}
	return nil, fmt.Errorf("not string list")
}

func IntResult(r interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}

	if x, ok := r.(int); ok {
		return x, err
	}

	return 0, fmt.Errorf("not int")
}
