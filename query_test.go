package sqlStorage

import (
	"testing"
	"reflect"

	"github.com/rs/rest-layer/schema/query"
)

func TestTranslatePredicate(t *testing.T) {
	testCases := []struct{
		predicate string
		sqlQuery string
		sqlParams []interface{}
	}{
		// Regression Test: Predicate items must be joined with AND
		{
			"{a: 1, b: 1}",
			"a IS ? AND b IS ?",
			[]interface{}{1.0, 1.0},
		},
		{
			"{int: 1}",
			"int IS ?",
			[]interface{}{1.0},
		},
		{
			"{str: \"foo\"}",
			"str LIKE ?",
			[]interface{}{"foo"},
		},
		{
			"{$or: [{int: 1}, {str: \"foo\"}]}",
			"(int IS ? OR str LIKE ?)",
			[]interface{}{1.0, "foo"},
		},
		{
			"{$and: [{int: 1}, {str: \"foo\"}]}",
			"(int IS ? AND str LIKE ?)",
			[]interface{}{1.0, "foo"},
		},
		{
			"{int: {$in: [1, 2]}}",
			"int IN (?)",
			[]interface{}{[]interface{}{1.0, 2.0}},
		},
		{
			"{str: {$in: [\"foo\", \"bar\"]}}",
			"str IN (?)",
			[]interface{}{[]interface{}{"foo", "bar"}},
		},
		{
			"{int: {$nin: [1, 2]}}",
			"int NOT IN (?)",
			[]interface{}{[]interface{}{1.0, 2.0}},
		},
		{
			"{str: {$nin: [\"foo\", \"bar\"]}}",
			"str NOT IN (?)",
			[]interface{}{[]interface{}{"foo", "bar"}},
		},
		{
			"{int: {$lt: 1}}",
			"int < ?",
			[]interface{}{1.0},
		},
		{
			"{int: {$lte: 1}}",
			"int <= ?",
			[]interface{}{1.0},
		},
		{
			"{int: {$gt: 1}}",
			"int > ?",
			[]interface{}{1.0},
		},
		{
			"{int: {$gte: 1}}",
			"int >= ?",
			[]interface{}{1.0},
		},
		{
			"{$or: [{$and: [{int: 1}, {str: \"foo\"}]}, {$and: [{int: 2}, {str: \"bar\"}]}]}",
			"((int IS ? AND str LIKE ?) OR (int IS ? AND str LIKE ?))",
			[]interface{}{1.0, "foo", 2.0, "bar"},
		},
	}

	for _, testCase := range testCases {
		p, err := query.ParsePredicate(testCase.predicate)
		if err != nil {
			t.Fail()
		}

		sqlQuery, sqlParams, err := translatePredicate(p)
		if err != nil || sqlQuery != testCase.sqlQuery || !reflect.DeepEqual(sqlParams, testCase.sqlParams){
			t.Log(testCase.predicate)
			t.Fail()
		}
	}
}