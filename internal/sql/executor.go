package sql

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"startdb/internal/storage"
)

// Executor represents a SQL query executor
type Executor struct {
	storage *storage.Storage
}

// NewExecutor creates a new SQL executor
func NewExecutor(storage *storage.Storage) *Executor {
	return &Executor{
		storage: storage,
	}
}

// Execute executes a SQL statement
func (e *Executor) Execute(stmt Statement) (*QueryResult, error) {
	switch s := stmt.(type) {
	case *SelectStatement:
		return e.executeSelect(s)
	case *InsertStatement:
		return e.executeInsert(s)
	case *UpdateStatement:
		return e.executeUpdate(s)
	case *DeleteStatement:
		return e.executeDelete(s)
	case *CreateTableStatement:
		return e.executeCreateTable(s)
	case *DropTableStatement:
		return e.executeDropTable(s)
	default:
		return nil, fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (e *Executor) executeSelect(stmt *SelectStatement) (*QueryResult, error) {
	// Check if table exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err != nil {
		return nil, fmt.Errorf("table '%s' does not exist", stmt.Table)
	}

	// Get all data for the table
	keys, err := e.storage.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	var rows [][]interface{}
	tablePrefix := stmt.Table + ":"

	// Filter keys for this table
	for _, key := range keys {
		if strings.HasPrefix(key, tablePrefix) {
			value, err := e.storage.Get(key)
			if err != nil {
				continue
			}

			// Parse the stored data
			rowData, err := e.parseRowData(string(value))
			if err != nil {
				continue
			}

			// Apply WHERE clause if present
			if stmt.Where != nil {
				matches, err := e.evaluateWhere(rowData, stmt.Where)
				if err != nil {
					continue
				}
				if !matches {
					continue
				}
			}

			rows = append(rows, rowData)
		}
	}

	// Apply ORDER BY if present
	if len(stmt.OrderBy) > 0 {
		sort.Slice(rows, func(i, j int) bool {
			// Simple ordering by first column for now
			if len(rows[i]) > 0 && len(rows[j]) > 0 {
				return e.compareValues(rows[i][0], rows[j][0]) < 0
			}
			return false
		})
	}

	// Apply LIMIT if present
	if stmt.Limit > 0 && stmt.Limit < len(rows) {
		rows = rows[:stmt.Limit]
	}

	// Determine columns
	columns := []string{"id"}
	if len(rows) > 0 {
		// Get columns from first row
		for i := 1; i < len(rows[0]); i += 2 {
			if i+1 < len(rows[0]) {
				columns = append(columns, rows[0][i].(string))
			}
		}
	}

	return &QueryResult{
		Columns: columns,
		Rows:    rows,
		Count:   len(rows),
	}, nil
}

func (e *Executor) executeInsert(stmt *InsertStatement) (*QueryResult, error) {
	// Check if table exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err != nil {
		return nil, fmt.Errorf("table '%s' does not exist", stmt.Table)
	}

	insertedCount := 0

	for _, valueList := range stmt.Values {
		// Generate a unique ID
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		key := fmt.Sprintf("%s:%s", stmt.Table, id)

		// Build the row data
		var rowData []interface{}
		rowData = append(rowData, id)

		// Get table metadata to determine column names
		tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
		tableMetadata, err := e.storage.Get(tableKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get table metadata: %w", err)
		}
		
		// Parse table metadata to get column names
		tableInfo := string(tableMetadata)
		// Extract column names from metadata
		columnNames := []string{"id", "name", "email"} // Default fallback
		if strings.Contains(tableInfo, "columns:") {
			parts := strings.Split(tableInfo, "columns:")
			if len(parts) > 1 {
				columnNames = strings.Split(parts[1], ",")
			}
		}
		if len(stmt.Columns) > 0 {
			columnNames = stmt.Columns
		}
		
		for i, value := range valueList {
			var columnName string
			if i < len(columnNames) {
				columnName = columnNames[i]
			} else {
				columnName = fmt.Sprintf("column_%d", i+1)
			}
			rowData = append(rowData, columnName, e.evaluateExpression(value))
		}

		// Store the row
		rowStr := e.serializeRowData(rowData)
		err = e.storage.Put(key, []byte(rowStr))
		if err != nil {
			return nil, fmt.Errorf("failed to insert row: %w", err)
		}

		insertedCount++
	}

	return &QueryResult{
		Columns: []string{"affected_rows"},
		Rows:    [][]interface{}{{insertedCount}},
		Count:   1,
	}, nil
}

func (e *Executor) executeUpdate(stmt *UpdateStatement) (*QueryResult, error) {
	// Check if table exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err != nil {
		return nil, fmt.Errorf("table '%s' does not exist", stmt.Table)
	}

	keys, err := e.storage.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	updatedCount := 0
	tablePrefix := stmt.Table + ":"

	for _, key := range keys {
		if strings.HasPrefix(key, tablePrefix) {
			value, err := e.storage.Get(key)
			if err != nil {
				continue
			}

			// Parse the stored data
			rowData, err := e.parseRowData(string(value))
			if err != nil {
				continue
			}

			// Apply WHERE clause if present
			if stmt.Where != nil {
				matches, err := e.evaluateWhere(rowData, stmt.Where)
				if err != nil {
					continue
				}
				if !matches {
					continue
				}
			}

			// Update the row
			updatedRowData := e.updateRowData(rowData, stmt.Set)
			updatedRowStr := e.serializeRowData(updatedRowData)
			err = e.storage.Put(key, []byte(updatedRowStr))
			if err != nil {
				return nil, fmt.Errorf("failed to update row: %w", err)
			}

			updatedCount++
		}
	}

	return &QueryResult{
		Columns: []string{"affected_rows"},
		Rows:    [][]interface{}{{updatedCount}},
		Count:   1,
	}, nil
}

func (e *Executor) executeDelete(stmt *DeleteStatement) (*QueryResult, error) {
	// Check if table exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err != nil {
		return nil, fmt.Errorf("table '%s' does not exist", stmt.Table)
	}

	keys, err := e.storage.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	deletedCount := 0
	tablePrefix := stmt.Table + ":"

	for _, key := range keys {
		if strings.HasPrefix(key, tablePrefix) {
			value, err := e.storage.Get(key)
			if err != nil {
				continue
			}

			// Parse the stored data
			rowData, err := e.parseRowData(string(value))
			if err != nil {
				continue
			}

			// Apply WHERE clause if present
			if stmt.Where != nil {
				matches, err := e.evaluateWhere(rowData, stmt.Where)
				if err != nil {
					continue
				}
				if !matches {
					continue
				}
			}

			// Delete the row
			err = e.storage.Delete(key)
			if err != nil {
				return nil, fmt.Errorf("failed to delete row: %w", err)
			}

			deletedCount++
		}
	}

	return &QueryResult{
		Columns: []string{"affected_rows"},
		Rows:    [][]interface{}{{deletedCount}},
		Count:   1,
	}, nil
}

func (e *Executor) executeCreateTable(stmt *CreateTableStatement) (*QueryResult, error) {
	// Check if table already exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err == nil {
		return nil, fmt.Errorf("table '%s' already exists", stmt.Table)
	}

	// Create table metadata
	table := &TableMetadata{
		Name:    stmt.Table,
		Created: time.Now(),
	}

	for _, colDef := range stmt.Columns {
		column := ColumnMetadata{
			Name:     colDef.Name,
			Type:     colDef.Type,
			Nullable: colDef.Nullable,
		}
		if colDef.Default != nil {
			column.Default = e.evaluateExpression(colDef.Default)
		}
		table.Columns = append(table.Columns, column)
	}

	// Store table metadata in storage with column names
	var columnNames []string
	for _, col := range stmt.Columns {
		columnNames = append(columnNames, col.Name)
	}
	tableData := fmt.Sprintf("table:%s:created:%d:columns:%s", stmt.Table, table.Created.Unix(), strings.Join(columnNames, ","))
	err = e.storage.Put(tableKey, []byte(tableData))
	if err != nil {
		return nil, fmt.Errorf("failed to store table metadata: %w", err)
	}

	return &QueryResult{
		Columns: []string{"message"},
		Rows:    [][]interface{}{{"Table created successfully"}},
		Count:   1,
	}, nil
}

func (e *Executor) executeDropTable(stmt *DropTableStatement) (*QueryResult, error) {
	// Check if table exists
	tableKey := fmt.Sprintf("_table_metadata:%s", stmt.Table)
	_, err := e.storage.Get(tableKey)
	if err != nil {
		return nil, fmt.Errorf("table '%s' does not exist", stmt.Table)
	}

	// Delete all rows for this table
	keys, err := e.storage.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	tablePrefix := stmt.Table + ":"
	for _, key := range keys {
		if strings.HasPrefix(key, tablePrefix) {
			e.storage.Delete(key)
		}
	}

	// Remove table metadata
	e.storage.Delete(tableKey)

	return &QueryResult{
		Columns: []string{"message"},
		Rows:    [][]interface{}{{"Table dropped successfully"}},
		Count:   1,
	}, nil
}

// Helper methods

func (e *Executor) parseRowData(data string) ([]interface{}, error) {
	// Simple CSV-like parsing for now
	parts := strings.Split(data, "|")
	var rowData []interface{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		rowData = append(rowData, part)
	}
	return rowData, nil
}

func (e *Executor) serializeRowData(rowData []interface{}) string {
	var parts []string
	for _, value := range rowData {
		parts = append(parts, fmt.Sprintf("%v", value))
	}
	return strings.Join(parts, "|")
}

func (e *Executor) evaluateExpression(expr Expression) interface{} {
	switch e := expr.(type) {
	case *StringLiteral:
		return e.Value
	case *NumberLiteral:
		return e.Value
	case *BooleanLiteral:
		return e.Value
	case *NullLiteral:
		return nil
	case *Identifier:
		return e.Value
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func (e *Executor) evaluateExpressionWithRowData(rowData []interface{}, expr Expression) interface{} {
	switch e := expr.(type) {
	case *StringLiteral:
		return e.Value
	case *NumberLiteral:
		return e.Value
	case *BooleanLiteral:
		return e.Value
	case *NullLiteral:
		return nil
	case *Identifier:
		// Look up the column value in the row data
		columnName := e.Value
		for i := 1; i < len(rowData); i += 2 {
			if i+1 < len(rowData) {
				if rowData[i] == columnName {
					return rowData[i+1]
				}
			}
		}
		return nil
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func (e *Executor) evaluateWhere(rowData []interface{}, where Expression) (bool, error) {
	switch w := where.(type) {
	case *BinaryExpression:
		left := e.evaluateExpressionWithRowData(rowData, w.Left)
		right := e.evaluateExpressionWithRowData(rowData, w.Right)

		switch w.Operator {
		case "=":
			return e.compareValues(left, right) == 0, nil
		case "!=", "<>":
			return e.compareValues(left, right) != 0, nil
		case "<":
			return e.compareValues(left, right) < 0, nil
		case ">":
			return e.compareValues(left, right) > 0, nil
		case "<=":
			return e.compareValues(left, right) <= 0, nil
		case ">=":
			return e.compareValues(left, right) >= 0, nil
		case "AND":
			leftResult, err := e.evaluateWhere(rowData, w.Left)
			if err != nil {
				return false, err
			}
			rightResult, err := e.evaluateWhere(rowData, w.Right)
			if err != nil {
				return false, err
			}
			return leftResult && rightResult, nil
		case "OR":
			leftResult, err := e.evaluateWhere(rowData, w.Left)
			if err != nil {
				return false, err
			}
			rightResult, err := e.evaluateWhere(rowData, w.Right)
			if err != nil {
				return false, err
			}
			return leftResult || rightResult, nil
		default:
			return false, fmt.Errorf("unsupported operator: %s", w.Operator)
		}
	default:
		return false, fmt.Errorf("unsupported where expression: %T", where)
	}
}

func (e *Executor) compareValues(a, b interface{}) int {
	switch aVal := a.(type) {
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			if !aVal && bVal {
				return -1
			} else if aVal && !bVal {
				return 1
			}
			return 0
		}
	}
	return 0
}

func (e *Executor) updateRowData(rowData []interface{}, setMap map[string]Expression) []interface{} {
	// Create a map for easier column access
	columnMap := make(map[string]interface{})
	for i := 1; i < len(rowData); i += 2 {
		if i+1 < len(rowData) {
			columnMap[rowData[i].(string)] = rowData[i+1]
		}
	}

	// Update columns
	for column, expr := range setMap {
		columnMap[column] = e.evaluateExpression(expr)
	}

	// Rebuild row data
	var newRowData []interface{}
	newRowData = append(newRowData, rowData[0]) // Keep ID
	for column, value := range columnMap {
		newRowData = append(newRowData, column, value)
	}

	return newRowData
}
