package xconv

import (
	"github.com/duanhf2012/origin/v2/util/xreflect"
	"reflect"
)

func Anys(any interface{}) []interface{} {
	if any == nil {
		return nil
	}

	switch rk, rv := xreflect.Value(any); rk {
	case reflect.Slice, reflect.Array:
		count := rv.Len()
		slice := make([]interface{}, count)
		for i := 0; i < count; i++ {
			slice[i] = rv.Index(i).Interface()
		}
		return slice
	default:
		return nil
	}
}
