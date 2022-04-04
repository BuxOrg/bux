package utils

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModelName = "test_model"
	testTableName = "test_model"
)

type testModel struct {
	Field string `json:"field"`
}

// GetModelName will return a model name
func (t *testModel) GetModelName() string {
	return testModelName
}

// GetModelTableName will return a table name
func (t *testModel) GetModelTableName() string {
	return testTableName
}

type badModel struct {
	Field string `json:"field"`
}

// TestGetModelStringAttribute will test the method GetModelStringAttribute()
func TestGetModelStringAttribute(t *testing.T) {
	t.Parallel()

	type TestModel struct {
		StringField string `json:"string_field"`
		ID          string `json:"id"`
	}

	t.Run("valid string attribute", func(t *testing.T) {
		m := &TestModel{
			StringField: "test",
			ID:          "12345678",
		}
		field1 := GetModelStringAttribute(m, "StringField")
		id := GetModelStringAttribute(m, "ID")
		assert.Equal(t, "test", *field1)
		assert.Equal(t, "12345678", *id)
	})

	t.Run("nil input", func(t *testing.T) {
		id := GetModelStringAttribute(nil, "ID")
		assert.Nil(t, id)
	})

	t.Run("invalid type", func(t *testing.T) {
		id := GetModelStringAttribute("invalid-type", "ID")
		assert.Nil(t, id)
	})
}

// TestGetModelBoolAttribute will test the method GetModelBoolAttribute()
func TestGetModelBoolAttribute(t *testing.T) {
	t.Parallel()

	type TestModel struct {
		BoolField bool   `json:"bool_field"`
		ID        string `json:"id"`
	}

	t.Run("valid bool attribute", func(t *testing.T) {
		m := &TestModel{
			BoolField: true,
		}
		field1 := GetModelBoolAttribute(m, "BoolField")
		assert.Equal(t, true, *field1)
	})

	t.Run("nil input", func(t *testing.T) {
		val := GetModelBoolAttribute(nil, "BoolField")
		assert.Nil(t, val)
	})

	t.Run("invalid type", func(t *testing.T) {
		assert.Panics(t, func() {
			val := GetModelBoolAttribute("invalid-type", "BoolField")
			assert.Nil(t, val)
		})
	})
}

// TestGetModelUnset will test the method GetModelUnset()
func TestGetModelUnset(t *testing.T) {
	t.Parallel()

	type TestModel struct {
		NullableTime   NullTime   `json:"nullable_time" bson:"nullable_time"`
		NullableString NullString `json:"nullable_string" bson:"nullable_string"`
		Internal       string     `json:"-" bson:"-"`
	}

	t.Run("basic test", func(t *testing.T) {
		ty := make(map[string]bool)
		m := &TestModel{
			NullableTime: NullTime{sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			}},
			NullableString: NullString{sql.NullString{
				String: "",
				Valid:  false,
			}},
			Internal: "test",
		}
		un := GetModelUnset(m)
		assert.IsType(t, ty, un)
		assert.Equal(t, true, un["nullable_time"])
		assert.Equal(t, true, un["nullable_string"])
		assert.Equal(t, false, un["internal"])
	})

	// todo: finish this test with better model tests
}

// TestIsModelSlice will test the method IsModelSlice()
func TestIsModelSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid slices", func(t *testing.T) {
		s := []string{"test"}

		assert.Equal(t, true, IsModelSlice(s))

		i := []int{1}
		assert.Equal(t, true, IsModelSlice(i))

		in := []interface{}{"test"}
		assert.Equal(t, true, IsModelSlice(in))

		ptr := []string{"test"}

		assert.Equal(t, true, IsModelSlice(&ptr))

	})

	t.Run("not a slice", func(t *testing.T) {
		s := "string"
		assert.Equal(t, false, IsModelSlice(s))

		i := 1
		assert.Equal(t, false, IsModelSlice(i))

	})
}

// TestGetModelName will test the method GetModelName()
func TestGetModelName(t *testing.T) {
	t.Parallel()

	t.Run("model is nil", func(t *testing.T) {
		name := GetModelName(nil)
		require.Nil(t, name)
	})

	t.Run("model is set - pointer", func(t *testing.T) {
		tm := &testModel{Field: testModelName}
		name := GetModelName(tm)
		assert.Equal(t, testModelName, *name)
	})

	t.Run("model is set - value", func(t *testing.T) {
		tm := testModel{Field: testModelName}
		name := GetModelName(tm)
		assert.Equal(t, testModelName, *name)
	})

	t.Run("model does not have method - pointer", func(t *testing.T) {
		tm := &badModel{}
		name := GetModelName(tm)
		assert.Nil(t, name)
	})

	t.Run("model does not have method - value", func(t *testing.T) {
		tm := badModel{}
		name := GetModelName(tm)
		assert.Nil(t, name)
	})
}

// TestGetModelTableName will test the method GetModelTableName()
func TestGetModelTableName(t *testing.T) {
	t.Parallel()

	t.Run("model is nil", func(t *testing.T) {
		name := GetModelTableName(nil)
		require.Nil(t, name)
	})

	t.Run("model is set - pointer", func(t *testing.T) {
		tm := &testModel{Field: testTableName}
		name := GetModelTableName(tm)
		assert.Equal(t, testTableName, *name)
	})

	t.Run("model is set - value", func(t *testing.T) {
		tm := testModel{Field: testTableName}
		name := GetModelTableName(tm)
		assert.Equal(t, testTableName, *name)
	})

	t.Run("model does not have method - pointer", func(t *testing.T) {
		tm := &badModel{}
		name := GetModelTableName(tm)
		assert.Nil(t, name)
	})

	t.Run("model does not have method - value", func(t *testing.T) {
		tm := badModel{}
		name := GetModelTableName(tm)
		assert.Nil(t, name)
	})
}

// TestGetModelType will test the method GetModelType()
func TestGetModelType(t *testing.T) {
	t.Parallel()

	type modelExample struct {
		Field string `json:"field"`
	}

	t.Run("panic - nil model", func(t *testing.T) {
		assert.Panics(t, func() {
			modelType := GetModelType(nil)
			require.NotNil(t, modelType)
		})
	})

	t.Run("default type", func(t *testing.T) {
		m := new(modelExample)
		modelType := GetModelType(m)
		assert.NotNil(t, modelType)
	})

	// todo: finish tests!
}
