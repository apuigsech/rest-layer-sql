package sqlStorage

import (
	"fmt"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
)



func buildSelectQuery(tableName string, q *query.Query) (sqlQuery string, sqlParams []interface{}, err error) {
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

	return sqlQuery, sqlParams, nil
}



func buildInsertQuery(tableName string, i *resource.Item) (sqlQuery string, sqlParams []interface{}, err error) {
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

	return sqlQuery, sqlParams, nil
}


func buildUpdateQuery(tableName string, i *resource.Item, o *resource.Item) (sqlQuery string, sqlParams []interface{}, err error) {
	setStr := "etag=?,"
	sqlParams = append(sqlParams, i.ETag)

	for k, v := range i.Payload {
		if k != "id" {
			setStr += k + "=?,"
			sqlParams = append(sqlParams, v)
		}
	}

	setStr = setStr[:len(setStr)-1]

	sqlParams = append(sqlParams, o.ID)
	sqlParams = append(sqlParams, o.ETag)


	sqlQuery = fmt.Sprintf("UPDATE OR ROLLBACK %s SET %s WHERE id=? AND etag=?", tableName, setStr)

	return sqlQuery, sqlParams, nil
}

func buildDeleteQuery(tableName string, i *resource.Item) (sqlQuery string, sqlParams []interface{}, err error) {
	sqlParams = append(sqlParams, i.ID)
	sqlParams = append(sqlParams, i.ETag)

	sqlQuery = fmt.Sprintf("DELETE FROM %s WHERE id = ? AND etag = ?", tableName)

	return sqlQuery, sqlParams, nil
}


func buildClearQuery(tableName string, q *query.Query) (sqlQuery string, sqlParams []interface{}, err error) {
	predicateQuery, predicateParams, err := buildPredicateQuery(q)
	if err != nil {
		return "", []interface{}{}, err
	}
	

	sqlQuery = fmt.Sprintf("DELETE FROM %s", tableName)
	if predicateQuery != "" {
		sqlQuery += fmt.Sprintf(" WHERE %s", predicateQuery)
		sqlParams = append(sqlParams, predicateParams...)
	}

	return sqlQuery, sqlParams, nil
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
	for _, exp := range q {
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