package sql

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v8"
	"github.com/doug-martin/goqu/v8/exp"

	_ "github.com/denisenkom/go-mssqldb"                // Import for MsSQL
	_ "github.com/doug-martin/goqu/v8/dialect/postgres" // Dialect for postfres
	_ "github.com/go-sql-driver/mysql"                  // Import for MySQL
	_ "github.com/lib/pq"                               // Import for postgres

	"github.com/spaceuptech/space-cloud/model"
	"github.com/spaceuptech/space-cloud/utils"
)

// generateReadQuery makes a query for read operation
func (s *SQL) generateReadQuery(ctx context.Context, project, col string, req *model.ReadRequest) (string, []interface{}, error) {
	dialect := goqu.Dialect(s.dbType)
	query := dialect.From(s.getDBName(project, col)).Prepared(true)
	var tarr []string
	if req.Find != nil {
		// Get the where clause from query object
		var err error
		query, tarr, err = s.generateWhereClause(query, req.Find)
		if err != nil {
			return "", nil, err
		}
	}

	selArray := []interface{}{}

	if req.Options != nil {
		// Check if the select clause exists
		if req.Options.Select != nil {
			for key := range req.Options.Select {
				selArray = append(selArray, key)
			}
		}

		if req.Options.Skip != nil {
			query = query.Offset(uint(*req.Options.Skip))
		}

		if req.Options.Limit != nil {
			query = query.Limit(uint(*req.Options.Limit))
		}

		if req.Options.Sort != nil {
			// Format the order array to a suitable type
			orderMap := req.Options.Sort

			orderBys := []exp.OrderedExpression{}

			// Iterate over order array
			for k, value := range orderMap {
				// Add order type based on type attribute of order element
				var exp exp.OrderedExpression
				if value < 0 {
					exp = goqu.I(k).Desc()
				} else {
					exp = goqu.I(k).Asc()
				}

				// Append the order expression to the order expression array
				orderBys = append(orderBys, exp)
			}
			query = query.Order(orderBys...)
		}
	}

	switch req.Operation {
	case utils.Count:
		query = query.Select(goqu.COUNT("*"))
	case utils.Distinct:
		distinct := req.Options.Distinct
		if distinct == nil {
			return "", nil, utils.ErrInvalidParams
		}
		query = query.SelectDistinct(*distinct)
	case utils.One, utils.All:
		query = query.Select(selArray...)
	}

	// Generate the sql string and arguments
	sqlString, args, err := query.ToSQL()
	if err != nil {
		return "", nil, err
	}
	sqlString = strings.Replace(sqlString, "\"", "", -1)

	for _, v := range tarr {
		switch s.dbType {
		case "mysql":
			vReplaced := strings.Replace(v, "=", "REGEXP", -1)
			sqlString = strings.Replace(sqlString, v, vReplaced, -1)
		case "postgres":
			vReplaced := strings.Replace(v, "=", "~", -1)
			sqlString = strings.Replace(sqlString, v, vReplaced, -1)
		}

	}
	return sqlString, args, nil
}

// Read query document(s) from the database
func (s *SQL) Read(ctx context.Context, project, col string, req *model.ReadRequest) (int64, interface{}, error) {
	return s.read(ctx, project, col, req, s.client)
}

func (s *SQL) read(ctx context.Context, project, col string, req *model.ReadRequest, executor executor) (int64, interface{}, error) {
	sqlString, args, err := s.generateReadQuery(ctx, project, col, req)
	if err != nil {
		return 0, nil, err
	}

	stmt, err := executor.PreparexContext(ctx, sqlString)
	if err != nil {
		return 0, nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, args...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var rowTypes []*sql.ColumnType

	switch s.GetDBType() {
	case utils.MySQL, utils.Postgres:
		rowTypes, _ = rows.ColumnTypes()
	}

	switch req.Operation {
	case utils.Count:
		mapping := make(map[string]interface{})
		if !rows.Next() {
			return 0, nil, errors.New("SQL: No response from db")
		}

		err := rows.MapScan(mapping)
		if err != nil {
			return 0, nil, err
		}

		switch s.GetDBType() {
		case utils.MySQL, utils.Postgres:
			mysqlTypeCheck(s.GetDBType(), rowTypes, mapping)
		}

		for _, v := range mapping {
			return v.(int64), v, nil
		}

		return 0, nil, errors.New("unknown error occurred")

	case utils.One:
		mapping := make(map[string]interface{})
		if !rows.Next() {
			return 0, nil, errors.New("SQL: No response from db")
		}

		err := rows.MapScan(mapping)
		if err != nil {
			return 0, nil, err
		}

		switch s.GetDBType() {
		case utils.MySQL, utils.Postgres:
			mysqlTypeCheck(s.GetDBType(), rowTypes, mapping)
		}

		return 1, mapping, nil

	case utils.All, utils.Distinct:
		array := []interface{}{}
		var count int64
		for rows.Next() {

			// Increment the counter
			count++

			mapping := make(map[string]interface{})
			err := rows.MapScan(mapping)
			if err != nil {
				return 0, nil, err
			}

			switch s.GetDBType() {
			case utils.MySQL, utils.Postgres:
				mysqlTypeCheck(s.GetDBType(), rowTypes, mapping)
			}

			array = append(array, mapping)
		}

		return count, array, nil

	default:
		return 0, nil, utils.ErrInvalidParams
	}
}
