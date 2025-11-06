package sql

import (
	"fmt"
	"strings"

	"startdb/internal/storage"
)

type PlanType string

const (
	PlanTypeIndexScan PlanType = "index_scan"
	PlanTypeTableScan PlanType = "table_scan"
	PlanTypeIndexRange PlanType = "index_range"
)

type ExecutionPlan struct {
	Type        PlanType
	Table       string
	IndexName   string
	IndexColumn string
	IndexValue  interface{}
	Where       Expression
	OrderBy     []Expression
	Limit       int
	Offset      int
	EstimatedCost int
}

type Planner struct {
	storage *storage.Storage
}

func NewPlanner(storage *storage.Storage) *Planner {
	return &Planner{
		storage: storage,
	}
}

func (p *Planner) PlanSelect(stmt *SelectStatement) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Table:   stmt.Table,
		Where:   stmt.Where,
		OrderBy: stmt.OrderBy,
		Limit:   stmt.Limit,
		Offset:  stmt.Offset,
	}

	if stmt.Where == nil {
		plan.Type = PlanTypeTableScan
		plan.EstimatedCost = 1000
		return plan, nil
	}

	columnName, columnValue, canUseIndex := p.extractIndexableColumn(stmt.Where)
	if !canUseIndex || columnName == "" || columnValue == nil {
		plan.Type = PlanTypeTableScan
		plan.EstimatedCost = 1000
		return plan, nil
	}

	indexManager := p.storage.GetIndexManager()
	allIndexes := indexManager.ListIndexes()
	
	var bestIndex string
	var foundIndex string
	
	for _, idx := range allIndexes {
		if idx == fmt.Sprintf("%s_%s_idx", stmt.Table, columnName) {
			foundIndex = idx
			break
		}
		
		indexMetadataKey := fmt.Sprintf("_index_metadata:%s", idx)
		indexMetadata, err := p.storage.Get(indexMetadataKey)
		if err == nil {
			metadata := string(indexMetadata)
			if strings.Contains(metadata, fmt.Sprintf("table:%s", stmt.Table)) && 
			   strings.Contains(metadata, fmt.Sprintf("column:%s", columnName)) {
				foundIndex = idx
				break
			}
		}
	}

	if foundIndex != "" {
		plan.Type = PlanTypeIndexScan
		plan.IndexName = foundIndex
		plan.IndexColumn = columnName
		plan.IndexValue = columnValue
		plan.EstimatedCost = 10
		bestIndex = foundIndex
	} else {
		plan.Type = PlanTypeTableScan
		plan.EstimatedCost = 1000
		return plan, nil
	}

	if p.hasOrderBy(stmt.OrderBy, columnName) {
		plan.EstimatedCost = 5
	}

	if stmt.Limit > 0 && stmt.Limit < 100 {
		plan.EstimatedCost = max(1, plan.EstimatedCost-2)
	}

	return plan, nil
}

func (p *Planner) PlanInsert(stmt *InsertStatement) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Type:        PlanTypeTableScan,
		Table:       stmt.Table,
		EstimatedCost: 50,
	}
	return plan, nil
}

func (p *Planner) PlanUpdate(stmt *UpdateStatement) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Type:        PlanTypeTableScan,
		Table:       stmt.Table,
		Where:       stmt.Where,
		EstimatedCost: 500,
	}

	if stmt.Where != nil {
		columnName, columnValue, canUseIndex := p.extractIndexableColumn(stmt.Where)
		if canUseIndex && columnName != "" && columnValue != nil {
			indexManager := p.storage.GetIndexManager()
			allIndexes := indexManager.ListIndexes()
			
			for _, idx := range allIndexes {
				indexMetadataKey := fmt.Sprintf("_index_metadata:%s", idx)
				indexMetadata, err := p.storage.Get(indexMetadataKey)
				if err == nil {
					metadata := string(indexMetadata)
					if strings.Contains(metadata, fmt.Sprintf("table:%s", stmt.Table)) && 
					   strings.Contains(metadata, fmt.Sprintf("column:%s", columnName)) {
						plan.Type = PlanTypeIndexScan
						plan.IndexName = idx
						plan.IndexColumn = columnName
						plan.IndexValue = columnValue
						plan.EstimatedCost = 100
						break
					}
				}
			}
		}
	}

	return plan, nil
}

func (p *Planner) PlanDelete(stmt *DeleteStatement) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Type:        PlanTypeTableScan,
		Table:       stmt.Table,
		Where:       stmt.Where,
		EstimatedCost: 500,
	}

	if stmt.Where != nil {
		columnName, columnValue, canUseIndex := p.extractIndexableColumn(stmt.Where)
		if canUseIndex && columnName != "" && columnValue != nil {
			indexManager := p.storage.GetIndexManager()
			allIndexes := indexManager.ListIndexes()
			
			for _, idx := range allIndexes {
				indexMetadataKey := fmt.Sprintf("_index_metadata:%s", idx)
				indexMetadata, err := p.storage.Get(indexMetadataKey)
				if err == nil {
					metadata := string(indexMetadata)
					if strings.Contains(metadata, fmt.Sprintf("table:%s", stmt.Table)) && 
					   strings.Contains(metadata, fmt.Sprintf("column:%s", columnName)) {
						plan.Type = PlanTypeIndexScan
						plan.IndexName = idx
						plan.IndexColumn = columnName
						plan.IndexValue = columnValue
						plan.EstimatedCost = 100
						break
					}
				}
			}
		}
	}

	return plan, nil
}

func (p *Planner) extractIndexableColumn(where Expression) (string, interface{}, bool) {
	switch w := where.(type) {
	case *BinaryExpression:
		if w.Operator == "=" {
			leftIdent, okLeft := w.Left.(*Identifier)
			if okLeft {
				rightVal := p.evaluateExpression(w.Right)
				if rightVal != nil {
					return leftIdent.Value, rightVal, true
				}
			}
			rightIdent, okRight := w.Right.(*Identifier)
			if okRight {
				leftVal := p.evaluateExpression(w.Left)
				if leftVal != nil {
					return rightIdent.Value, leftVal, true
				}
			}
		}
	}
	return "", nil, false
}

func (p *Planner) evaluateExpression(expr Expression) interface{} {
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

func (p *Planner) hasOrderBy(orderBy []Expression, columnName string) bool {
	for _, expr := range orderBy {
		if ident, ok := expr.(*Identifier); ok {
			if ident.Value == columnName {
				return true
			}
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
