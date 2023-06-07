package _type

import (
	"context"
	"fmt"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm/schema"
	"reflect"
)

type GormUlid ulid.ULID

func (id *GormUlid) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	switch value := dbValue.(type) {
	case GormUlid:
		*id = value
	case string:
		*id = GormUlid(ulid.MustParse(value))
	default:
		return fmt.Errorf("unsupported data while parsing GormUlidType: %s", dbValue)
	}
	return nil
}

func (id GormUlid) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return ulid.ULID(id).String(), nil
}
