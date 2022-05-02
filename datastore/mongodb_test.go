package datastore

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: finish unit tests!

// TestClient_saveWithMongo will test the method saveWithMongo()
func TestClient_saveWithMongo(t *testing.T) {
	// finish test
}

// TestClient_incrementWithMongo will test the method incrementWithMongo()
func TestClient_incrementWithMongo(t *testing.T) {
	// finish test
}

// TestClient_getWithMongo will test the method getWithMongo()
func TestClient_getWithMongo(t *testing.T) {
	// finish test
}

// TestClient_setPrefix will test the method setPrefix()
func TestClient_setPrefix(t *testing.T) {
	// finish test
}

type mockModel struct {
	ID string `json:"id"`
}

type testStruct struct {
	ID              string `json:"id" toml:"id" yaml:"hash" bson:"_id"`
	CurrentBalance  uint64 `json:"current_balance" toml:"current_balance" yaml:"current_balance" bson:"current_balance"`
	NextInternalNum uint32 `json:"next_internal_num" toml:"next_internal_num" yaml:"next_internal_num" bson:"next_internal_num"`
	NextExternalNum uint32 `json:"next_external_num" toml:"next_external_num" yaml:"next_external_num" bson:"next_external_num"`
}

// TestClient_getFieldNames will test the method getFieldNames()
func TestClient_getFieldNames(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		fields := getFieldNames(nil)
		assert.Len(t, fields, 0)
		assert.Equal(t, []string{}, fields)
	})

	t.Run("normal struct", func(t *testing.T) {
		s := testStruct{}
		fields := getFieldNames(s)
		assert.Len(t, fields, 4)
		assert.Equal(t, []string{"_id", "current_balance", "next_internal_num", "next_external_num"}, fields)
	})

	t.Run("pointer struct", func(t *testing.T) {
		s := &testStruct{}
		fields := getFieldNames(s)
		assert.Len(t, fields, 4)
		assert.Equal(t, []string{"_id", "current_balance", "next_internal_num", "next_external_num"}, fields)
	})

	t.Run("slice of structs", func(t *testing.T) {
		s := []testStruct{}
		fields := getFieldNames(s)
		assert.Len(t, fields, 4)
		assert.Equal(t, []string{"_id", "current_balance", "next_internal_num", "next_external_num"}, fields)
	})

	t.Run("pointer of slice of structs", func(t *testing.T) {
		s := &[]testStruct{}
		fields := getFieldNames(s)
		assert.Len(t, fields, 4)
		assert.Equal(t, []string{"_id", "current_balance", "next_internal_num", "next_external_num"}, fields)
	})

	t.Run("pointer of slice of pointers of structs", func(t *testing.T) {
		s := &[]*testStruct{}
		fields := getFieldNames(s)
		assert.Len(t, fields, 4)
		assert.Equal(t, []string{"_id", "current_balance", "next_internal_num", "next_external_num"}, fields)
	})
}

// TestClient_getMongoQueryConditions will test the method getMongoQueryConditions()
func TestClient_getMongoQueryConditions(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		condition := map[string]interface{}{}
		queryConditions := getMongoQueryConditions(Transaction{}, condition)
		assert.Equal(t, map[string]interface{}{}, queryConditions)
	})

	t.Run("simple", func(t *testing.T) {
		condition := map[string]interface{}{
			"test-key": "test-value",
		}
		queryConditions := getMongoQueryConditions(Transaction{}, condition)
		assert.Equal(t, map[string]interface{}{"test-key": "test-value"}, queryConditions)
	})

	t.Run("ID", func(t *testing.T) {
		condition := map[string]interface{}{}
		queryConditions := getMongoQueryConditions(mockModel{
			ID: "identifier",
		}, condition)
		assert.Equal(t, map[string]interface{}{"_id": "identifier"}, queryConditions)
	})

	t.Run("$or ID", func(t *testing.T) {
		condition := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"id": "test-key",
			}},
		}
		queryConditions := getMongoQueryConditions(nil, condition)
		expected := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"_id": "test-key",
			}},
		}
		assert.Equal(t, expected, queryConditions)
	})

	t.Run("$and $or ID", func(t *testing.T) {
		condition := map[string]interface{}{
			"metadata": map[string]interface{}{},
			"$and": []map[string]interface{}{{
				"$or": []map[string]interface{}{{
					"id": "test-key",
				}},
			}},
		}
		queryConditions := getMongoQueryConditions(nil, condition)
		expected := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"$or": []map[string]interface{}{{
					"_id": "test-key",
				}},
			}},
		}
		assert.Equal(t, expected, queryConditions)
	})

	t.Run("embedded ID", func(t *testing.T) {
		condition := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"metadata": map[string]interface{}{
					"test-key": "test-value",
				},
			}, {
				"id": "identifier",
			}},
		}
		expected := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"$and": []map[string]interface{}{{
					"metadata.k": "test-key", "metadata.v": "test-value",
				}},
			}, {
				"_id": "identifier",
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		assert.Equal(t, expected, queryConditions)
	})

	t.Run("metadata", func(t *testing.T) {
		condition := map[string]interface{}{
			"metadata": map[string]interface{}{
				"test-key": "test-value",
			},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"metadata.k": "test-key",
				"metadata.v": "test-value",
			}},
		}
		assert.Equal(t, expected, queryConditions)
	})

	t.Run("metadata 2", func(t *testing.T) {
		condition := map[string]interface{}{
			"metadata": map[string]interface{}{
				"test-key":  "test-value",
				"test-key2": "test-value2",
			},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"metadata.k": "test-key",
			"metadata.v": "test-value",
		}, {
			"metadata.k": "test-key2",
			"metadata.v": "test-value2",
		}}
		assert.Len(t, queryConditions["$and"], 2)
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
	})

	t.Run("metadata 3", func(t *testing.T) {
		condition := map[string]interface{}{
			"metadata": map[string]interface{}{
				"test-key":  "test-value",
				"test-key2": "test-value2",
			},
			"$and": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lt": 98,
				},
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"metadata.k": "test-key",
			"metadata.v": "test-value",
		}, {
			"metadata.k": "test-key2",
			"metadata.v": "test-value2",
		}, {
			"fee": map[string]interface{}{
				"$lt": float64(98),
			},
		}}
		assert.Len(t, queryConditions["$and"], 3)
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[2])
	})

	t.Run("metadata $or", func(t *testing.T) {
		condition := map[string]interface{}{
			"metadata": map[string]interface{}{
				"test-key":  "test-value",
				"test-key2": "test-value2",
			},
			"$or": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lt": 98,
				},
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"metadata.k": "test-key",
			"metadata.v": "test-value",
		}, {
			"metadata.k": "test-key2",
			"metadata.v": "test-value2",
		}}
		expectedOr := []map[string]interface{}{{
			"fee": map[string]interface{}{
				"$lt": float64(98),
			},
		}}
		assert.Len(t, queryConditions["$and"], 2)
		assert.Len(t, queryConditions["$or"], 1)
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
		assert.Contains(t, expectedOr, queryConditions["$or"].([]map[string]interface{})[0])
	})

	t.Run("xpub_metdata", func(t *testing.T) {
		condition := map[string]interface{}{
			"xpub_metadata": map[string]interface{}{
				"xpubID": map[string]interface{}{
					"test-key": "test-value",
				},
			},
			"$and": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lt": 98,
				},
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"xpub_metadata.x": "xpubID",
			"xpub_metadata.k": "test-key",
			"xpub_metadata.v": "test-value",
		}, {
			"fee": map[string]interface{}{
				"$lt": float64(98),
			},
		}}
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
	})

	t.Run("xpub_metdata 2", func(t *testing.T) {
		condition := map[string]interface{}{
			"xpub_metadata": map[string]interface{}{
				"xpubID": map[string]interface{}{
					"test-key":  "test-value",
					"test-key2": "test-value2",
				},
			},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"xpub_metadata.x": "xpubID",
			"xpub_metadata.k": "test-key",
			"xpub_metadata.v": "test-value",
		}, {
			"xpub_metadata.x": "xpubID",
			"xpub_metadata.k": "test-key2",
			"xpub_metadata.v": "test-value2",
		}}
		assert.Len(t, queryConditions["$and"], 2)
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
	})

	t.Run("json interface", func(t *testing.T) {
		condition := map[string]interface{}{
			"xpub_metadata": map[string]interface{}{
				"xpubID": map[string]interface{}{
					"test-key":  "test-value",
					"test-key2": "test-value2",
				},
			},
		}
		c, err := json.Marshal(condition)
		require.NoError(t, err)

		var cc interface{}
		err = json.Unmarshal(c, &cc)
		require.NoError(t, err)
		queryConditions := getMongoQueryConditions(mockModel{}, cc.(map[string]interface{}))
		expected := []map[string]interface{}{{
			"xpub_metadata.x": "xpubID",
			"xpub_metadata.k": "test-key",
			"xpub_metadata.v": "test-value",
		}, {
			"xpub_metadata.x": "xpubID",
			"xpub_metadata.k": "test-key2",
			"xpub_metadata.v": "test-value2",
		}}
		assert.Len(t, queryConditions["$and"], 2)
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
	})

	t.Run("xpub_metdata 3", func(t *testing.T) {
		condition := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"xpub_in_ids": "xpub_id",
			}, {
				"xpub_out_ids": "xpub_id",
			}},
			"$and": []map[string]interface{}{{
				"$or": []map[string]interface{}{{
					"metadata": map[string]interface{}{"test-key": "test-value"},
				}, {
					"xpub_metadata": map[string]interface{}{
						"xpub_id": map[string]interface{}{"test-key": "test-value"},
					},
				}},
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		// {"$and":[{"$or":[{"$and":[{"metadata.k":"test-key","metadata.v":"test-value"}]},{"$and":[{"xpub_metadata.k":"test-key","xpub_metadata.v":"test-value"}],"xpub_metadata.x":"xpub_id"}]}],"$or":[{"xpub_in_ids":"xpub_id"},{"xpub_out_ids":"xpub_id"}]}
		assert.Len(t, queryConditions["$and"], 1)
		assert.Len(t, queryConditions["$or"], 2)

		expectedXpubID := []map[string]interface{}{{
			"xpub_in_ids": "xpub_id",
		}, {
			"xpub_out_ids": "xpub_id",
		}}
		assert.Contains(t, expectedXpubID, queryConditions["$or"].([]map[string]interface{})[0])
		assert.Contains(t, expectedXpubID, queryConditions["$or"].([]map[string]interface{})[1])

		expected0 := map[string]interface{}{
			"metadata.k": "test-key",
			"metadata.v": "test-value",
		}
		expected1 := map[string]interface{}{
			"xpub_metadata.x": "xpub_id",
			"xpub_metadata.k": "test-key",
			"xpub_metadata.v": "test-value",
		}
		or := (queryConditions["$and"].([]map[string]interface{})[0])["$or"]
		or0 := or.([]map[string]interface{})[0]
		or1 := or.([]map[string]interface{})[1]
		assert.Equal(t, expected0, or0["$and"].([]map[string]interface{})[0])
		assert.Equal(t, expected1, or1["$and"].([]map[string]interface{})[0])
	})

	t.Run("xpub_output_value", func(t *testing.T) {
		condition := map[string]interface{}{
			"xpub_output_value": map[string]interface{}{
				"xpubID": map[string]interface{}{
					"$gt": 0,
				},
			},
			"$and": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lt": 98,
				},
			}},
		}
		queryConditions := getMongoQueryConditions(mockModel{}, condition)
		expected := []map[string]interface{}{{
			"xpub_output_value.xpubID": map[string]interface{}{
				"$gt": float64(0),
			},
		}, {
			"fee": map[string]interface{}{
				"$lt": float64(98),
			},
		}}
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[0])
		assert.Contains(t, expected, queryConditions["$and"].([]map[string]interface{})[1])
	})
}

// TestClient_openMongoDatabase will test the method openMongoDatabase()
func TestClient_openMongoDatabase(t *testing.T) {
	// finish test
}

// TestClient_getMongoIndexes will test the method getMongoIndexes()
func TestClient_getMongoIndexes(t *testing.T) {
	// finish test
}
