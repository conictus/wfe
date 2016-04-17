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

func InterfaceResult(vs []interface{}, err error) (interface{}, error) {
	if err != nil {
		return "", err
	}

	if len(vs) != 2 {
		return "", fmt.Errorf("expecting 2 argument, got %d", len(vs))
	}

	o, e := vs[0], vs[1]
	if e != nil {
		if x, ok := e.(error); ok {
			return o, x
		} else {
			return o, fmt.Errorf("not error")
		}
	}

	return o, nil
}

func StringResult(vs []interface{}, err error) (string, error) {
	o, e := InterfaceResult(vs, err)

	if e != nil {
		return "", e
	}

	if x, ok := o.(string); ok {
		return x, e
	} else {
		return "", fmt.Errorf("not string")
	}
}

func IntResult(vs []interface{}, err error) (int, error) {
	o, e := InterfaceResult(vs, err)

	if e != nil {
		return 0, e
	}

	if x, ok := o.(int); ok {
		return x, e
	} else {
		return 0, fmt.Errorf("not int")
	}
}
