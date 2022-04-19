package datastore

import (
	"database/sql"
	"testing"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// todo: finish unit tests!

// Test_whereObject test the SQL where selector
func Test_whereSlice(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		query := whereSlice(MySQL, "xpub_in_ids", "id_1")
		expected := `JSON_CONTAINS(xpub_in_ids, CAST('["id_1"]' AS JSON))`
		assert.Equal(t, expected, query)
	})

	t.Run("Postgres", func(t *testing.T) {
		query := whereSlice(PostgreSQL, "xpub_in_ids", "id_1")
		expected := `xpub_in_ids::jsonb @> '["id_1"]'`
		assert.Equal(t, expected, query)
	})

	t.Run("SQLite", func(t *testing.T) {
		query := whereSlice(SQLite, "xpub_in_ids", "id_1")
		expected := `EXISTS (SELECT 1 FROM json_each(xpub_in_ids) WHERE value = "id_1")`
		assert.Equal(t, expected, query)
	})
}

// Test_processConditions test the SQL where selectors
func Test_processConditions(t *testing.T) {
	conditions := map[string]interface{}{
		"monitor": map[string]interface{}{
			"$gt": utils.NullTime{NullTime: sql.NullTime{
				Valid: true,
				Time:  time.Date(2022, 4, 4, 15, 12, 37, 651387237, time.UTC),
			}},
		},
	}

	t.Run("MySQL", func(t *testing.T) {
		tx := &mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		varNum := 0
		_ = processConditions(tx, conditions, MySQL, &varNum, nil)
		assert.Equal(t, "monitor > @var0", tx.WhereClauses[0])
		assert.Equal(t, "2022-04-04 15:12:37", tx.Vars["var0"])
	})

	t.Run("Postgres", func(t *testing.T) {
		tx := &mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		varNum := 0
		_ = processConditions(tx, conditions, PostgreSQL, &varNum, nil)
		assert.Equal(t, "monitor > @var0", tx.WhereClauses[0])
		assert.Equal(t, "2022-04-04T15:12:37Z", tx.Vars["var0"])
	})

	t.Run("SQLite", func(t *testing.T) {
		tx := &mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		varNum := 0
		_ = processConditions(tx, conditions, SQLite, &varNum, nil)
		assert.Equal(t, "monitor > @var0", tx.WhereClauses[0])
		assert.Equal(t, "2022-04-04T15:12:37.651Z", tx.Vars["var0"])
	})
}

// Test_whereObject test the SQL where selector
func Test_whereObject(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		metadata := map[string]interface{}{
			"test_key": "test-value",
		}
		query := whereObject(MySQL, "metadata", metadata)
		expected := "JSON_EXTRACT(metadata, '$.test_key') = \"test-value\""
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key": "test-'value'",
		}
		query = whereObject(MySQL, "metadata", metadata)
		expected = "JSON_EXTRACT(metadata, '$.test_key') = \"test-\\'value\\'\""
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key1": "test-value",
			"test_key2": "test-value2",
		}
		query = whereObject(MySQL, "metadata", metadata)

		assert.Contains(t, []string{
			"(JSON_EXTRACT(metadata, '$.test_key1') = \"test-value\" AND JSON_EXTRACT(metadata, '$.test_key2') = \"test-value2\")",
			"(JSON_EXTRACT(metadata, '$.test_key2') = \"test-value2\" AND JSON_EXTRACT(metadata, '$.test_key1') = \"test-value\")",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)

		xpubMetadata := map[string]interface{}{
			"testXpubId": map[string]interface{}{
				"test_key1": "test-value",
				"test_key2": "test-value2",
			},
		}
		query = whereObject(MySQL, "xpub_metadata", xpubMetadata)

		assert.Contains(t, []string{
			"(JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key1') = \"test-value\" AND JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key2') = \"test-value2\")",
			"(JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key2') = \"test-value2\" AND JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key1') = \"test-value\")",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)
	})

	t.Run("Postgres", func(t *testing.T) {
		metadata := map[string]interface{}{
			"test_key": "test-value",
		}
		query := whereObject(PostgreSQL, "metadata", metadata)
		expected := "metadata::jsonb @> '{\"test_key\":\"test-value\"}'::jsonb"
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key": "test-'value'",
		}
		query = whereObject(PostgreSQL, "metadata", metadata)
		expected = "metadata::jsonb @> '{\"test_key\":\"test-\\'value\\'\"}'::jsonb"
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key1": "test-value",
			"test_key2": "test-value2",
		}
		query = whereObject(PostgreSQL, "metadata", metadata)

		assert.Contains(t, []string{
			"(metadata::jsonb @> '{\"test_key1\":\"test-value\"}'::jsonb AND metadata::jsonb @> '{\"test_key2\":\"test-value2\"}'::jsonb)",
			"(metadata::jsonb @> '{\"test_key2\":\"test-value2\"}'::jsonb AND metadata::jsonb @> '{\"test_key1\":\"test-value\"}'::jsonb)",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)

		xpubMetadata := map[string]interface{}{
			"testXpubId": map[string]interface{}{
				"test_key1": "test-value",
				"test_key2": "test-value2",
			},
		}
		query = whereObject(PostgreSQL, "xpub_metadata", xpubMetadata)
		assert.Contains(t, []string{
			"xpub_metadata::jsonb @> '{\"testXpubId\":{\"test_key1\":\"test-value\",\"test_key2\":\"test-value2\"}}'::jsonb",
			"xpub_metadata::jsonb @> '{\"testXpubId\":{\"test_key2\":\"test-value2\",\"test_key1\":\"test-value\"}}'::jsonb",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)
	})

	t.Run("SQLite", func(t *testing.T) {
		metadata := map[string]interface{}{
			"test_key": "test-value",
		}
		query := whereObject(SQLite, "metadata", metadata)
		expected := "JSON_EXTRACT(metadata, '$.test_key') = \"test-value\""
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key": "test-'value'",
		}
		query = whereObject(SQLite, "metadata", metadata)
		expected = "JSON_EXTRACT(metadata, '$.test_key') = \"test-\\'value\\'\""
		assert.Equal(t, expected, query)

		metadata = map[string]interface{}{
			"test_key1": "test-value",
			"test_key2": "test-value2",
		}
		query = whereObject(SQLite, "metadata", metadata)
		assert.Contains(t, []string{
			"(JSON_EXTRACT(metadata, '$.test_key1') = \"test-value\" AND JSON_EXTRACT(metadata, '$.test_key2') = \"test-value2\")",
			"(JSON_EXTRACT(metadata, '$.test_key2') = \"test-value2\" AND JSON_EXTRACT(metadata, '$.test_key1') = \"test-value\")",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)

		xpubMetadata := map[string]interface{}{
			"testXpubId": map[string]interface{}{
				"test_key1": "test-value",
				"test_key2": "test-value2",
			},
		}
		query = whereObject(SQLite, "xpub_metadata", xpubMetadata)
		assert.Contains(t, []string{
			"(JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key1') = \"test-value\" AND JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key2') = \"test-value2\")",
			"(JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key2') = \"test-value2\" AND JSON_EXTRACT(xpub_metadata, '$.testXpubId.test_key1') = \"test-value\")",
		}, query)
		// NOTE: the order of the items can change, hence the query order can change
		// assert.Equal(t, expected, query)
	})
}

// mockSQLCtx is used to mock the SQL
type mockSQLCtx struct {
	WhereClauses []interface{}
	Vars         map[string]interface{}
}

func (f *mockSQLCtx) Where(query interface{}, args ...interface{}) {
	f.WhereClauses = append(f.WhereClauses, query)
	if len(args) > 0 {
		for _, variables := range args {
			for key, value := range variables.(map[string]interface{}) {
				f.Vars[key] = value
			}
		}
	}
}

func (f *mockSQLCtx) getGormTx() *gorm.DB {
	return nil
}

// TestBuxWhere will test the method BuxWhere()
func TestBuxWhere(t *testing.T) {
	t.Run("SQLite empty select", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{}
		_ = BuxWhere(&tx, conditions, SQLite)
		assert.Equal(t, []interface{}{}, tx.WhereClauses)
	})

	t.Run("SQLite simple select", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"ID": "testID",
		}
		_ = BuxWhere(&tx, conditions, SQLite)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, "ID = @var0", tx.WhereClauses[0])
		assert.Equal(t, "testID", tx.Vars["var0"])
	})

	t.Run("SQLite $or", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"xpub_in_ids": "xPubID",
			}, {
				"xpub_out_ids": "xPubID",
			}},
		}
		_ = BuxWhere(&tx, conditions, SQLite)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( (EXISTS (SELECT 1 FROM json_each(xpub_in_ids) WHERE value = \"xPubID\")) OR (EXISTS (SELECT 1 FROM json_each(xpub_out_ids) WHERE value = \"xPubID\")) ) ", tx.WhereClauses[0])
	})

	t.Run("MySQL $or", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"xpub_in_ids": "xPubID",
			}, {
				"xpub_out_ids": "xPubID",
			}},
		}
		_ = BuxWhere(&tx, conditions, MySQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( (JSON_CONTAINS(xpub_in_ids, CAST('[\"xPubID\"]' AS JSON))) OR (JSON_CONTAINS(xpub_out_ids, CAST('[\"xPubID\"]' AS JSON))) ) ", tx.WhereClauses[0])
	})

	t.Run("PostgreSQL $or", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"xpub_in_ids": "xPubID",
			}, {
				"xpub_out_ids": "xPubID",
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( (xpub_in_ids::jsonb @> '[\"xPubID\"]') OR (xpub_out_ids::jsonb @> '[\"xPubID\"]') ) ", tx.WhereClauses[0])
	})

	t.Run("SQLite metadata", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"metadata": map[string]interface{}{
				"xpub_id": "xPubID",
			},
		}
		_ = BuxWhere(&tx, conditions, SQLite)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, "JSON_EXTRACT(metadata, '$.xpub_id') = \"xPubID\"", tx.WhereClauses[0])
	})

	t.Run("MySQL metadata", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"metadata": map[string]interface{}{
				"xpub_id": "xPubID",
			},
		}
		_ = BuxWhere(&tx, conditions, MySQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, "JSON_EXTRACT(metadata, '$.xpub_id') = \"xPubID\"", tx.WhereClauses[0])
	})

	t.Run("PostgreSQL metadata", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"metadata": map[string]interface{}{
				"xpub_id": "xPubID",
			},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, "metadata::jsonb @> '{\"xpub_id\":\"xPubID\"}'::jsonb", tx.WhereClauses[0])
	})

	t.Run("SQLite $and", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"reference_id": "reference",
			}, {
				"chain": 12,
			}, {
				"$or": []map[string]interface{}{{
					"xpub_in_ids": "xPubID",
				}, {
					"xpub_out_ids": "xPubID",
				}},
			}},
		}
		_ = BuxWhere(&tx, conditions, SQLite)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( reference_id = @var0 AND chain = @var1 AND  ( (EXISTS (SELECT 1 FROM json_each(xpub_in_ids) WHERE value = \"xPubID\")) OR (EXISTS (SELECT 1 FROM json_each(xpub_out_ids) WHERE value = \"xPubID\")) )  ) ", tx.WhereClauses[0])
		assert.Equal(t, "reference", tx.Vars["var0"])
		assert.Equal(t, 12, tx.Vars["var1"])
	})

	t.Run("MySQL $and", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"reference_id": "reference",
			}, {
				"chain": 12,
			}, {
				"$or": []map[string]interface{}{{
					"xpub_in_ids": "xPubID",
				}, {
					"xpub_out_ids": "xPubID",
				}},
			}},
		}
		_ = BuxWhere(&tx, conditions, MySQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( reference_id = @var0 AND chain = @var1 AND  ( (JSON_CONTAINS(xpub_in_ids, CAST('[\"xPubID\"]' AS JSON))) OR (JSON_CONTAINS(xpub_out_ids, CAST('[\"xPubID\"]' AS JSON))) )  ) ", tx.WhereClauses[0])
		assert.Equal(t, "reference", tx.Vars["var0"])
		assert.Equal(t, 12, tx.Vars["var1"])
	})

	t.Run("PostgreSQL $and", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"reference_id": "reference",
			}, {
				"chain": 12,
			}, {
				"$or": []map[string]interface{}{{
					"xpub_in_ids": "xPubID",
				}, {
					"xpub_out_ids": "xPubID",
				}},
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL)
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( reference_id = @var0 AND chain = @var1 AND  ( (xpub_in_ids::jsonb @> '[\"xPubID\"]') OR (xpub_out_ids::jsonb @> '[\"xPubID\"]') )  ) ", tx.WhereClauses[0])
		assert.Equal(t, "reference", tx.Vars["var0"])
		assert.Equal(t, 12, tx.Vars["var1"])
	})

	t.Run("Where $gt", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"fee": map[string]interface{}{
				"$gt": 502,
			},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL) // all the same
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, "fee > @var0", tx.WhereClauses[0])
		assert.Equal(t, 502, tx.Vars["var0"])
	})

	t.Run("Where $gt $lt", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lt": 503,
				},
			}, {
				"fee": map[string]interface{}{
					"$gt": 203,
				},
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL) // all the same
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( fee < @var0 AND fee > @var1 ) ", tx.WhereClauses[0])
		assert.Equal(t, 503, tx.Vars["var0"])
		assert.Equal(t, 203, tx.Vars["var1"])
	})

	t.Run("Where $gte $lte", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"fee": map[string]interface{}{
					"$lte": 203,
				},
			}, {
				"fee": map[string]interface{}{
					"$gte": 1203,
				},
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL) // all the same
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( (fee <= @var0) OR (fee >= @var1) ) ", tx.WhereClauses[0])
		assert.Equal(t, 203, tx.Vars["var0"])
		assert.Equal(t, 1203, tx.Vars["var1"])
	})

	t.Run("Where $or $and $or $gte $lte", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$or": []map[string]interface{}{{
				"$and": []map[string]interface{}{{
					"fee": map[string]interface{}{
						"$lte": 203,
					},
				}, {
					"$or": []map[string]interface{}{{
						"fee": map[string]interface{}{
							"$gte": 1203,
						},
					}, {
						"value": map[string]interface{}{
							"$gte": 2203,
						},
					}},
				}},
			}, {
				"$and": []map[string]interface{}{{
					"fee": map[string]interface{}{
						"$gte": 3203,
					},
				}, {
					"value": map[string]interface{}{
						"$gte": 4203,
					},
				}},
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL) // all the same
		assert.Len(t, tx.WhereClauses, 1)
		assert.Equal(t, " ( ( ( fee <= @var0 AND  ( (fee >= @var1) OR (value >= @var2) )  ) ) OR ( ( fee >= @var3 AND value >= @var4 ) ) ) ", tx.WhereClauses[0])
		assert.Equal(t, 203, tx.Vars["var0"])
		assert.Equal(t, 1203, tx.Vars["var1"])
		assert.Equal(t, 2203, tx.Vars["var2"])
		assert.Equal(t, 3203, tx.Vars["var3"])
		assert.Equal(t, 4203, tx.Vars["var4"])
	})

	t.Run("Where $and $or $or $gte $lte", func(t *testing.T) {
		tx := mockSQLCtx{
			WhereClauses: make([]interface{}, 0),
			Vars:         make(map[string]interface{}),
		}
		conditions := map[string]interface{}{
			"$and": []map[string]interface{}{{
				"$and": []map[string]interface{}{{
					"fee": map[string]interface{}{
						"$lte": 203,
						"$gte": 103,
					},
				}, {
					"$or": []map[string]interface{}{{
						"fee": map[string]interface{}{
							"$gte": 1203,
						},
					}, {
						"value": map[string]interface{}{
							"$gte": 2203,
						},
					}},
				}},
			}, {
				"$or": []map[string]interface{}{{
					"fee": map[string]interface{}{
						"$gte": 3203,
					},
				}, {
					"value": map[string]interface{}{
						"$gte": 4203,
					},
				}},
			}},
		}
		_ = BuxWhere(&tx, conditions, PostgreSQL) // all the same
		assert.Len(t, tx.WhereClauses, 1)
		assert.Contains(t, []string{
			" (  ( fee <= @var0 AND fee >= @var1 AND  ( (fee >= @var2) OR (value >= @var3) )  )  AND  ( (fee >= @var4) OR (value >= @var5) )  ) ",
			" (  ( fee >= @var0 AND fee <= @var1 AND  ( (fee >= @var2) OR (value >= @var3) )  )  AND  ( (fee >= @var4) OR (value >= @var5) )  ) ",
		}, tx.WhereClauses[0])

		// assert.Equal(t, " (  ( fee <= @var0 AND fee >= @var1 AND  ( (fee >= @var2) OR (value >= @var3) )  )  AND  ( (fee >= @var4) OR (value >= @var5) )  ) ", tx.WhereClauses[0])

		assert.Contains(t, []int{203, 103}, tx.Vars["var0"])
		assert.Contains(t, []int{203, 103}, tx.Vars["var1"])
		// assert.Equal(t, 203, tx.Vars["var0"])
		// assert.Equal(t, 103, tx.Vars["var1"])
		assert.Equal(t, 1203, tx.Vars["var2"])
		assert.Equal(t, 2203, tx.Vars["var3"])
		assert.Equal(t, 3203, tx.Vars["var4"])
		assert.Equal(t, 4203, tx.Vars["var5"])
	})
}

// Test_escapeDBString will test the method escapeDBString()
func Test_escapeDBString(t *testing.T) {
	// finish test
}
