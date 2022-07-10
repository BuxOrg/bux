package bux

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/mrz1836/go-datastore"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ValueTypeString is the value type "string"
const ValueTypeString = "string"

// IDs are string ids saved as an array
type IDs []string

// GormDataType type in gorm
func (i IDs) GormDataType() string {
	return gormTypeText
}

// Scan scan value into JSON, implements sql.Scanner interface
func (i *IDs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	xType := fmt.Sprintf("%T", value)
	var byteValue []byte
	if xType == ValueTypeString {
		byteValue = []byte(value.(string))
	} else {
		byteValue = value.([]byte)
	}

	return json.Unmarshal(byteValue, &i)
}

// Value return json value, implement driver.Valuer interface
func (i IDs) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	marshal, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}

// MarshalIDs will unmarshal the custom type
func MarshalIDs(i IDs) graphql.Marshaler {
	if i == nil {
		return graphql.Null
	}
	return graphql.MarshalAny(i)
}

// GormDBDataType the gorm data type for metadata
func (IDs) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == datastore.Postgres {
		return datastore.JSONB
	}
	return datastore.JSON
}

// UnmarshalIDs will marshal the custom type
func UnmarshalIDs(v interface{}) (IDs, error) {
	if v == nil {
		return nil, nil
	}

	// Try to unmarshal
	ids, err := graphql.UnmarshalAny(v)
	if err != nil {
		return nil, err
	}

	// Cast interface back to IDs (safely)
	if idCast, ok := ids.(IDs); ok {
		return idCast, nil
	}

	return nil, ErrCannotConvertToIDs
}
