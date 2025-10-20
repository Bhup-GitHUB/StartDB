package sql

import (
	"fmt"
	"time"
)

// Node represents a node in the Abstract Syntax Tree
type Node interface {
	String() string
}

// Statement represents a SQL statement
type Statement interface {
	Node
	statementNode()
}

// Expression represents a SQL expression
type Expression interface {
	Node
	expressionNode()
}

// SelectStatement represents a SELECT statement
type SelectStatement struct {
	Fields    []Expression
	Table     string
	Where     Expression
	OrderBy   []Expression
	Limit     int
	Offset    int
}

func (s *SelectStatement) statementNode() {}
func (s *SelectStatement) String() string {
	return "SELECT statement"
}

// InsertStatement represents an INSERT statement
type InsertStatement struct {
	Table   string
	Columns []string
	Values  [][]Expression
}

func (i *InsertStatement) statementNode() {}
func (i *InsertStatement) String() string {
	return "INSERT statement"
}

// UpdateStatement represents an UPDATE statement
type UpdateStatement struct {
	Table string
	Set   map[string]Expression
	Where Expression
}

func (u *UpdateStatement) statementNode() {}
func (u *UpdateStatement) String() string {
	return "UPDATE statement"
}

// DeleteStatement represents a DELETE statement
type DeleteStatement struct {
	Table string
	Where Expression
}

func (d *DeleteStatement) statementNode() {}
func (d *DeleteStatement) String() string {
	return "DELETE statement"
}

// CreateTableStatement represents a CREATE TABLE statement
type CreateTableStatement struct {
	Table   string
	Columns []ColumnDefinition
}

func (c *CreateTableStatement) statementNode() {}
func (c *CreateTableStatement) String() string {
	return "CREATE TABLE statement"
}

// ColumnDefinition represents a column definition
type ColumnDefinition struct {
	Name     string
	Type     string
	Nullable bool
	Default  Expression
}

// DropTableStatement represents a DROP TABLE statement
type DropTableStatement struct {
	Table string
}

func (d *DropTableStatement) statementNode() {}
func (d *DropTableStatement) String() string {
	return "DROP TABLE statement"
}

// Expression types

// Identifier represents a column or table name
type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string {
	return i.Value
}

// StringLiteral represents a string literal
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) expressionNode() {}
func (s *StringLiteral) String() string {
	return "'" + s.Value + "'"
}

// NumberLiteral represents a numeric literal
type NumberLiteral struct {
	Value float64
}

func (n *NumberLiteral) expressionNode() {}
func (n *NumberLiteral) String() string {
	return fmt.Sprintf("%g", n.Value)
}

// BooleanLiteral represents a boolean literal
type BooleanLiteral struct {
	Value bool
}

func (b *BooleanLiteral) expressionNode() {}
func (b *BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NullLiteral represents a NULL literal
type NullLiteral struct{}

func (n *NullLiteral) expressionNode() {}
func (n *NullLiteral) String() string {
	return "NULL"
}

// BinaryExpression represents a binary operation (e.g., a = b, a > b)
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (b *BinaryExpression) expressionNode() {}
func (b *BinaryExpression) String() string {
	return b.Left.String() + " " + b.Operator + " " + b.Right.String()
}

// FunctionCall represents a function call (e.g., COUNT(*), MAX(column))
type FunctionCall struct {
	Name string
	Args []Expression
}

func (f *FunctionCall) expressionNode() {}
func (f *FunctionCall) String() string {
	return f.Name + "()"
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
	Count   int
	Error   error
}

// TableMetadata represents metadata about a table
type TableMetadata struct {
	Name    string
	Columns []ColumnMetadata
	Created time.Time
}

// ColumnMetadata represents metadata about a column
type ColumnMetadata struct {
	Name     string
	Type     string
	Nullable bool
	Default  interface{}
}
