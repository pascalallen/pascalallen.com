package query

import "reflect"

type ListRepositories struct{}

func (q *ListRepositories) QueryName() string {
	return reflect.TypeOf(q).Elem().Name()
}
