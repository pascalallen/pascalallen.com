package command

import "reflect"

type RecordVisit struct {
	Language  string
	UserAgent string
	ClientIP  string
}

func (c *RecordVisit) CommandName() string {
	return reflect.TypeOf(c).Elem().Name()
}
