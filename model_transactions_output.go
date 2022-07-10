package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/mrz1836/go-datastore"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// XpubOutputValue Xpub specific output value of the transaction
type XpubOutputValue map[string]int64

// Scan scan value into Json, implements sql.Scanner interface
func (x *XpubOutputValue) Scan(value interface{}) error {
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
	if bytes.Equal(byteValue, []byte("")) || bytes.Equal(byteValue, []byte("\"\"")) {
		return nil
	}

	return json.Unmarshal(byteValue, &x)
}

// Value return json value, implement driver.Valuer interface
func (x XpubOutputValue) Value() (driver.Value, error) {
	if x == nil {
		return nil, nil
	}
	marshal, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}

// GormDBDataType the gorm data type for metadata
func (XpubOutputValue) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == datastore.Postgres {
		return datastore.JSONB
	}
	return datastore.JSON
}
