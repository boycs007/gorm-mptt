package mptt

import (
	"context"
	"gorm.io/gorm/schema"
	"reflect"
)

// KeyField mptt has id parent_id tree_id left right level fields
// id parent_id tree_id only support positive integer or string
// left right level only support positive integer
type KeyField struct {
	*schema.Field
	Attr string
}

// setFieldValue ignore error
func setFieldValue(n interface{}, field KeyField, value interface{}) {
	ctx := context.Background()
	_ = field.Set(ctx, reflect.ValueOf(n), value)
}

func getIntFieldValue(n interface{}, field KeyField) int {
	ctx := context.Background()
	v, _ := field.ValueOf(ctx, reflect.ValueOf(n))
	switch data := v.(type) {
	case int64:
		return int(data)
	case int:
		return data
	case int8:
		return int(data)
	case int16:
		return int(data)
	case int32:
		return int(data)
	case uint:
		return int(data)
	case uint8:
		return int(data)
	case uint16:
		return int(data)
	case uint32:
		return int(data)
	case uint64:
		return int(data)
	case float32:
		return int(data)
	case float64:
		return int(data)
	default:
		return 0
	}
}

func getFieldValue(n interface{}, field KeyField) interface{} {
	ctx := context.Background()
	v, _ := field.ValueOf(ctx, reflect.ValueOf(n))
	return v
}

type KeyFields struct {
	ID     KeyField
	Parent KeyField
	Tree   KeyField
	Left   KeyField
	Right  KeyField
	Level  KeyField
}
