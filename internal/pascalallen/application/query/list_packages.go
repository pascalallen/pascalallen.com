package query

import "reflect"

type ListPackages struct{}

func (q *ListPackages) QueryName() string {
	return reflect.TypeOf(q).Elem().Name()
}
