package datastore

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/BuxOrg/bux/utils"
	"gorm.io/gorm"
)

// todo: these are specific fields, inject these in the function params?
var arrayFields = []string{"xpub_in_ids", "xpub_out_ids"}
var objectFields = []string{"metadata", "xpub_metadata", "xpub_output_value"}

type buxWhereInterface interface {
	Where(query interface{}, args ...interface{})
	getGormTx() *gorm.DB
}

// BuxWhere add conditions
func BuxWhere(tx buxWhereInterface, conditions map[string]interface{}, engine Engine) interface{} {

	varNum := 0
	processConditions(tx, conditions, engine, &varNum, nil)

	return tx.getGormTx()
}

type txAccumulator struct {
	buxWhereInterface
	WhereClauses []string
	Vars         map[string]interface{}
}

func (tx *txAccumulator) Where(query interface{}, args ...interface{}) {
	tx.WhereClauses = append(tx.WhereClauses, query.(string))

	if len(args) > 0 {
		for _, variables := range args {
			for key, value := range variables.(map[string]interface{}) {
				tx.Vars[key] = value
			}
		}
	}
}

func (tx *txAccumulator) getGormTx() *gorm.DB {
	return nil
}

func processConditions(tx buxWhereInterface, conditions map[string]interface{}, engine Engine, varNum *int, parentKey *string) map[string]interface{} {

	for key, condition := range conditions {
		if key == "$and" {
			processWhereAnd(tx, condition, engine, varNum)
		} else if key == "$or" {
			processWhereOr(tx, conditions["$or"], engine, varNum)
		} else if key == "$gt" {
			varName := "var" + strconv.Itoa(*varNum)
			tx.Where(*parentKey+" > @"+varName, map[string]interface{}{varName: condition})
			*varNum++
		} else if key == "$lt" {
			varName := "var" + strconv.Itoa(*varNum)
			tx.Where(*parentKey+" < @"+varName, map[string]interface{}{varName: condition})
			*varNum++
		} else if key == "$gte" {
			varName := "var" + strconv.Itoa(*varNum)
			tx.Where(*parentKey+" >= @"+varName, map[string]interface{}{varName: condition})
			*varNum++
		} else if key == "$lte" {
			varName := "var" + strconv.Itoa(*varNum)
			tx.Where(*parentKey+" <= @"+varName, map[string]interface{}{varName: condition})
			*varNum++
		} else if utils.StringInSlice(key, arrayFields) {
			tx.Where(whereSlice(engine, key, condition))
		} else if utils.StringInSlice(key, objectFields) {
			tx.Where(whereObject(engine, key, condition))
		} else {
			if condition == nil {
				tx.Where(key + " IS NULL")
			} else {
				v := reflect.ValueOf(condition)
				switch v.Kind() { //nolint:exhaustive // not all cases are needed
				case reflect.Map:
					if _, ok := condition.(map[string]interface{}); ok {
						processConditions(tx, condition.(map[string]interface{}), engine, varNum, &key)
					} else {
						c, _ := json.Marshal(condition)
						var cc map[string]interface{}
						_ = json.Unmarshal(c, &cc)
						processConditions(tx, cc, engine, varNum, &key)
					}
				default:
					varName := "var" + strconv.Itoa(*varNum)
					tx.Where(key+" = @"+varName, map[string]interface{}{varName: condition})
					*varNum++
				}
			}
		}
	}

	return conditions
}

func processWhereAnd(tx buxWhereInterface, condition interface{}, engine Engine, varNum *int) {
	accumulator := &txAccumulator{
		WhereClauses: make([]string, 0),
		Vars:         make(map[string]interface{}),
	}
	for _, c := range condition.([]map[string]interface{}) {
		processConditions(accumulator, c, engine, varNum, nil)
	}

	if len(accumulator.Vars) > 0 {
		tx.Where(" ( "+strings.Join(accumulator.WhereClauses, " AND ")+" ) ", accumulator.Vars)
	} else {
		tx.Where(" ( " + strings.Join(accumulator.WhereClauses, " AND ") + " ) ")
	}
}

func processWhereOr(tx buxWhereInterface, condition interface{}, engine Engine, varNum *int) {
	or := make([]string, 0)
	orVars := make(map[string]interface{})
	for _, cond := range condition.([]map[string]interface{}) {
		statement := make([]string, 0)
		accumulator := &txAccumulator{
			WhereClauses: make([]string, 0),
			Vars:         make(map[string]interface{}),
		}
		processConditions(accumulator, cond, engine, varNum, nil)
		statement = append(statement, accumulator.WhereClauses...)
		for varName, varValue := range accumulator.Vars {
			orVars[varName] = varValue
		}
		or = append(or, strings.Join(statement[:], " AND "))
	}

	if len(orVars) > 0 {
		tx.Where(" ( ("+strings.Join(or, ") OR (")+") ) ", orVars)
	} else {
		tx.Where(" ( (" + strings.Join(or, ") OR (") + ") ) ")
	}
}

func escapeDBString(s string) string {
	rs := strings.Replace(s, "'", "\\'", -1)
	return strings.Replace(rs, "\"", "\\\"", -1)
}

func whereObject(engine Engine, k string, v interface{}) string {
	queryParts := make([]string, 0)

	// we don't know the type, we handle the rangeValue as a map[string]interface{}
	vJSON, _ := json.Marshal(v)
	var rangeV map[string]interface{}
	// todo: ignoring errors
	_ = json.Unmarshal(vJSON, &rangeV)

	for rangeKey, rangeValue := range rangeV {
		if engine == MySQL || engine == SQLite {
			switch vv := rangeValue.(type) {
			case string:
				rangeValue = "\"" + escapeDBString(rangeValue.(string)) + "\""
				queryParts = append(queryParts, "JSON_EXTRACT("+k+", '$."+rangeKey+"') = "+rangeValue.(string))
			default:
				metadataJSON, _ := json.Marshal(vv)
				var metadata map[string]interface{}
				// todo: ignoring errors
				_ = json.Unmarshal(metadataJSON, &metadata)
				for kk, vvv := range metadata {
					mJSON, _ := json.Marshal(vvv)
					vvv = string(mJSON)
					queryParts = append(queryParts, "JSON_EXTRACT("+k+", '$."+rangeKey+"."+kk+"') = "+vvv.(string))
				}
			}
		} else if engine == PostgreSQL {
			switch vv := rangeValue.(type) {
			case string:
				rangeValue = "\"" + escapeDBString(rangeValue.(string)) + "\""
			default:
				// todo: ignoring errors
				metadataJSON, _ := json.Marshal(vv)
				rangeValue = string(metadataJSON)
			}
			queryParts = append(queryParts, k+"::jsonb @> '{\""+rangeKey+"\":"+rangeValue.(string)+"}'::jsonb")
		} else {
			queryParts = append(queryParts, "JSON_EXTRACT("+k+", '$."+rangeKey+"') = '"+escapeDBString(rangeValue.(string))+"'")
		}
	}

	query := queryParts[0]
	if len(queryParts) > 1 {
		query = "(" + strings.Join(queryParts, " AND ") + ")"
	}

	return query
}

func whereSlice(engine Engine, k string, v interface{}) string {
	if engine == MySQL {
		return "JSON_CONTAINS(" + k + ", CAST('[\"" + v.(string) + "\"]' AS JSON))"
	} else if engine == PostgreSQL {
		return k + "::jsonb @> '[\"" + v.(string) + "\"]'"
	}
	return "EXISTS (SELECT 1 FROM json_each(" + k + ") WHERE value = \"" + v.(string) + "\")"
}
