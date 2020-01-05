package sqlStorage

import (
	"fmt"
	"reflect"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"

	"github.com/lib/pq"
)


func buildCreateQuery(tableName string, s *schema.Schema, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	schemaQuery, schemaParams, err := buildSchemaQuery(s)
	if err != nil {
		return "", []interface{}{}, err
	}

	sqlQuery = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, schemaQuery)
	sqlParams = append(sqlParams, schemaParams...)

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}


func buildSelectQuery(tableName string, q *query.Query, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	predicateQuery, predicateParams, err := buildPredicateQuery(q)
	if err != nil {
		return "", []interface{}{}, err
	}
	sortQuery, sortParams, err := buildSortQuery(q)
	if err != nil {
		return "", []interface{}{}, err
	}	

	sqlQuery = fmt.Sprintf("SELECT * FROM %s", tableName)
	if predicateQuery != "" {
		sqlQuery += fmt.Sprintf(" WHERE %s", predicateQuery)
		sqlParams = append(sqlParams, predicateParams...)
	}

	if sortQuery != "" {
		sqlQuery += fmt.Sprintf(" ORDER BY %s", sortQuery)
		sqlParams = append(sqlParams, sortParams...)
	}

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}



func buildInsertQuery(tableName string, i *resource.Item, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	columnsStr := "etag,"
	valuesStr := "?,"
	sqlParams = append(sqlParams, i.ETag)

	for k, v := range i.Payload {
		columnsStr += k + ","
		valuesStr += "?," 
		sqlParams = append(sqlParams, v)
	}

	columnsStr = columnsStr[:len(columnsStr)-1]
	valuesStr = valuesStr[:len(valuesStr)-1]

	sqlQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, valuesStr)

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}


func buildUpdateQuery(tableName string, i *resource.Item, o *resource.Item, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	setStr := "etag=?,"
	sqlParams = append(sqlParams, i.ETag)

	for k, v := range i.Payload {
		if k != "id" {
			setStr += fmt.Sprintf("%s=?,", k)
			sqlParams = append(sqlParams, v)
		}
	}

	for k, _ := range o.Payload {
		if _, ok := i.Payload[k]; ok == false {
			setStr += fmt.Sprintf("%s=NULL,", k)
		}
	}

	setStr = setStr[:len(setStr)-1]

	sqlParams = append(sqlParams, o.ID)
	sqlParams = append(sqlParams, o.ETag)


	sqlQuery = fmt.Sprintf("UPDATE %s SET %s WHERE id=? AND etag=?", tableName, setStr)

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}

func buildDeleteQuery(tableName string, i *resource.Item, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	sqlParams = append(sqlParams, i.ID)
	sqlQuery = fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	if i.ETag != "" {
		sqlParams = append(sqlParams, i.ETag)
		sqlQuery += " AND etag = ?"
	}

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}


func buildClearQuery(tableName string, q *query.Query, sqlBackend string) (sqlQuery string, sqlParams []interface{}, err error) {
	predicateQuery, predicateParams, err := buildPredicateQuery(q)
	if err != nil {
		return "", []interface{}{}, err
	}
	

	sqlQuery = fmt.Sprintf("DELETE FROM %s", tableName)
	if predicateQuery != "" {
		sqlQuery += fmt.Sprintf(" WHERE %s", predicateQuery)
		sqlParams = append(sqlParams, predicateParams...)
	}

	return transformQuery(sqlQuery, sqlBackend), transformParams(sqlParams, sqlBackend), nil
}


func buildPredicateQuery(q *query.Query) (sqlQuery string, sqlParams []interface{}, err error) {
	return translatePredicate(q.Predicate)
}

func buildSortQuery(q *query.Query) (sqlQuery string, sqlParams []interface{}, err error) {
	if len(q.Sort) == 0 {
		return "", []interface{}{}, nil
	}
	for _, s := range q.Sort {
		sqlQuery += s.Name
		if s.Reversed {
			sqlQuery += " DESC"
		}
		sqlQuery += ","
	}
	return sqlQuery[:len(sqlQuery)-1], []interface{}{}, nil
}

func translatePredicate(q query.Predicate) (sqlQuery string, sqlParams []interface{}, err error) {
	for i, exp := range q {
		if i > 0 {
			sqlQuery += " AND "
		}
		switch t := exp.(type) {
		case *query.And:
			var s string
			for _, subExp := range *t {
				sb, extraSqlParams, err := translatePredicate(query.Predicate{subExp})
				if err != nil {
					return "", []interface{}{}, err
				}
				sqlParams = append(sqlParams, extraSqlParams...)
				s += sb + " AND "
			}
			sqlQuery += "(" + s[:len(s)-5] + ")"
		case *query.Or:
			var s string
			for _, subExp := range *t {
				sb, extraSqlParams, err := translatePredicate(query.Predicate{subExp})
				if err != nil {
					return "", []interface{}{}, err
				}
				sqlParams = append(sqlParams, extraSqlParams...)
				s += sb + " OR "
			}
			sqlQuery += "(" + s[:len(s)-4] + ")"
		case *query.In:
			sqlParams = append(sqlParams, t.Values)
			sqlQuery += t.Field + " IN (?)"
		case *query.NotIn:
			sqlParams = append(sqlParams, t.Values)
			sqlQuery += t.Field + " NOT IN (?)"
		case *query.Equal:
			sqlParams = append(sqlParams, t.Value)
			switch t.Value.(type) {
			case string:
				sqlQuery += t.Field + " LIKE ?"
			default:
				sqlQuery += t.Field + " IS ?"
			}
		case *query.NotEqual:
			sqlParams = append(sqlParams, t.Value)
			switch t.Value.(type) {
			case string:
				sqlQuery += t.Field + " NOT LIKE ?"
			default:
				sqlQuery += t.Field + " IS NOT ?"
			}
		case *query.GreaterThan:
			sqlParams = append(sqlParams, t.Value)
			sqlQuery += t.Field + " > ?"
		case *query.GreaterOrEqual:
			sqlParams = append(sqlParams, t.Value)
			sqlQuery += t.Field + " >= ?"
		case *query.LowerThan:
			sqlParams = append(sqlParams, t.Value)
			sqlQuery += t.Field + " < ?"
		case *query.LowerOrEqual:
			sqlParams = append(sqlParams, t.Value)
			sqlQuery += t.Field + " <= ?"
		default:
			return "", []interface{}{}, resource.ErrNotImplemented
		}
	}
	return sqlQuery, sqlParams, nil
}

func buildSchemaQuery(s *schema.Schema) (sqlQuery string, sqlParams []interface{}, err error) {
	sqlQuery = "etag VARCHAR(512),"
	for fieldName, field := range s.Fields {
		switch field.Validator.(type) {
		case *schema.String:
			f := field.Validator.(*schema.String)
			sqlQuery += fmt.Sprintf("%s VARCHAR", fieldName)
			if f.MaxLen > 0 {
				sqlQuery += fmt.Sprintf("(%d)", f.MaxLen)
			}
		case *schema.Integer:
			sqlQuery += fmt.Sprintf("%s INTEGER", fieldName)
		case *schema.Float:
			sqlQuery += fmt.Sprintf("%s FLOAT", fieldName)
		case *schema.Bool:
			sqlQuery += fmt.Sprintf("%s BIT(1)", fieldName)
		case *schema.Time:
			sqlQuery += fmt.Sprintf("%s TIMESTAMP", fieldName)
		case *schema.URL:
			sqlQuery += fmt.Sprintf("%s VARCHAR", fieldName)
		default:
			return "", []interface{}{}, resource.ErrNotImplemented
		}
		sqlQuery += ","
	}

	return sqlQuery[:len(sqlQuery)-1], []interface{}{}, nil
}

func transformQuery(sqlQuery string, sqlBackend string) (string) {
	switch sqlBackend {
		case "sqlite3", "mysql":
			return sqlQuery
		case "postgres":
			return transformQuery_postgres(sqlQuery)
	}
	return sqlQuery
}

func transformQuery_postgres(sqlQuery string) (newSqlQuery string) {
	idx := 1
	for _, ch := range sqlQuery {
		if ch == '?' {
			newSqlQuery += fmt.Sprintf("$%d", idx)
			idx++
		} else {
			newSqlQuery += string(ch)
		}
		
	}
	return
}

func transformParams(sqlParams []interface{}, sqlBackend string) ([]interface{}) {
	switch sqlBackend {
		case "sqlite3", "mysql":
			return sqlParams
		case "postgres":
			return transformParams_postgres(sqlParams)
	}
	return sqlParams
}


func transformParams_postgres(sqlParams []interface{}) ([]interface{}) {
	var newSqlParams []interface{}

	for _,p := range sqlParams {
		t := reflect.TypeOf(p)
		switch t.Kind() {
			case reflect.Slice, reflect.Array:
				newSqlParams = append(newSqlParams, pq.Array(p))
			default:
				newSqlParams = append(newSqlParams, p)
		} 
	}

	return newSqlParams
}