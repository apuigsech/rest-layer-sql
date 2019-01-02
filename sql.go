package sqlStorage

import (
	"fmt"
	"context"
	"database/sql"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

type SQLHandler struct {
	driverName	string
	session		*sql.DB
	tableName 	string
}

func NewHandler(driverName string, dataSourceName string, tableName string) (h *SQLHandler, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	h = &SQLHandler{
		driverName: driverName,
		session:	db,
		tableName:	tableName,
	}

	return h, nil
}

func (h *SQLHandler) Create(ctx context.Context, s *schema.Schema) (err error) {
	txPtr, err := h.session.Begin()

	sqlQuery, sqlParams, err := buildCreateQuery(h.tableName, s, h.driverName)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	_, err = h.session.ExecContext(ctx, sqlQuery, sqlParams...)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	txPtr.Commit()

	return nil
}

func (h *SQLHandler) Find(ctx context.Context, q *query.Query, ) (list *resource.ItemList, err error) {
	list = &resource.ItemList{
		Total: 0,
		Limit: 10,
		Items: []*resource.Item{},
	}

	sqlQuery, sqlParams, err := buildSelectQuery(h.tableName, q, h.driverName)
	if err != nil {
		return nil, err
	}

	rows, err := h.session.QueryContext(ctx, sqlQuery, sqlParams...)
	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		rowMap := make(map[string]interface{})
		rowVals := make([]interface{}, len(cols))
		rowValPtrs := make([]interface{}, len(cols))
		var etag string

		for i, _ := range cols {
			rowValPtrs[i] = &rowVals[i]
		}

		err := rows.Scan(rowValPtrs...)
		if err != nil {
			return nil, err
		}

		for i, v := range rowVals {
			b, ok := v.([]byte)
			if ok {
				v = string(b)
			}

			if (cols[i] == "etag") {
				etag = v.(string)
			} else {
				rowMap[cols[i]] = v
			}
		}

		item := &resource.Item{
			ID:      rowMap["id"],
			ETag:    etag,
			//Updated: rowMap["updated"],
			Payload: rowMap,
		}

		list.Items = append(list.Items, item)
	}

	return list, nil
}

func (h *SQLHandler)Insert(ctx context.Context, items []*resource.Item) (err error) {
	txPtr, err := h.session.Begin()

	for _, i := range items {
		sqlQuery, sqlParams, err := buildInsertQuery(h.tableName, i, h.driverName)
		if err != nil {
			txPtr.Rollback()
			return err
		}

		fmt.Println(sqlQuery)

		_, err = h.session.ExecContext(ctx, sqlQuery, sqlParams...)
		if err != nil {
			txPtr.Rollback()
			return err
		}
	}

	txPtr.Commit()
	
	return nil
}

func (h *SQLHandler) Update(ctx context.Context, item *resource.Item, original *resource.Item) (err error) {
	txPtr, err := h.session.Begin()

	sqlQuery, sqlParams, err := buildUpdateQuery(h.tableName, item, original, h.driverName)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	_, err = h.session.ExecContext(ctx, sqlQuery, sqlParams...)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	txPtr.Commit()

	return nil
}

func (h *SQLHandler) Delete(ctx context.Context, item *resource.Item) (err error) {
	txPtr, err := h.session.Begin()

	sqlQuery, sqlParams, err := buildDeleteQuery(h.tableName, item, h.driverName)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	_, err = h.session.ExecContext(ctx, sqlQuery, sqlParams...)
	if err != nil {
		txPtr.Rollback()
		return err
	}

	txPtr.Commit()

	return nil
}

func (h *SQLHandler) Clear(ctx context.Context, q *query.Query) (total int, err error) {
	txPtr, err := h.session.Begin()

	sqlQuery, sqlParams, err := buildClearQuery(h.tableName, q, h.driverName)
	if err != nil {
		txPtr.Rollback()
		return 0,err
	}

	res, err := h.session.ExecContext(ctx, sqlQuery, sqlParams...)
	if err != nil {
		txPtr.Rollback()
		return 0,err
	}
	count, err := res.RowsAffected()
	if err != nil {
		txPtr.Rollback()
		return 0,err		
	}
	txPtr.Commit()

	return int(count), nil
}

