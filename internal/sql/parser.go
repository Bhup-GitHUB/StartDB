package sql

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser represents a SQL parser
type Parser struct {
	lexer *Lexer
}

// NewParser creates a new SQL parser
func NewParser(input string) *Parser {
	return &Parser{
		lexer: NewLexer(input),
	}
}

// Parse parses the SQL input and returns an AST
func (p *Parser) Parse() (Statement, error) {
	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	// Check for unexpected tokens
	if p.lexer.Peek().Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token: %s", p.lexer.Peek().Literal)
	}

	return stmt, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	token := p.lexer.Next()
	
	switch strings.ToUpper(token.Literal) {
	case "SELECT":
		return p.parseSelectStatement()
	case "INSERT":
		return p.parseInsertStatement()
	case "UPDATE":
		return p.parseUpdateStatement()
	case "DELETE":
		return p.parseDeleteStatement()
	case "CREATE":
		return p.parseCreateStatement()
	case "DROP":
		return p.parseDropStatement()
	default:
		return nil, fmt.Errorf("unexpected statement: %s", token.Literal)
	}
}

func (p *Parser) parseSelectStatement() (*SelectStatement, error) {
	stmt := &SelectStatement{}

	// Parse fields
	fields, err := p.parseFieldList()
	if err != nil {
		return nil, err
	}
	stmt.Fields = fields

	// Parse FROM clause
	if !p.expectKeyword("FROM") {
		return nil, fmt.Errorf("expected FROM")
	}

	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	// Parse WHERE clause
	if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "WHERE" {
		p.lexer.Next() // consume WHERE
		where, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	// Parse ORDER BY clause
	if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "ORDER" {
		p.lexer.Next() // consume ORDER
		if !p.expectKeyword("BY") {
			return nil, fmt.Errorf("expected BY after ORDER")
		}
		orderBy, err := p.parseFieldList()
		if err != nil {
			return nil, err
		}
		stmt.OrderBy = orderBy
	}

	// Parse LIMIT clause
	if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "LIMIT" {
		p.lexer.Next() // consume LIMIT
		limitToken := p.lexer.Next()
		if limitToken.Type != TokenNumber {
			return nil, fmt.Errorf("expected number after LIMIT")
		}
		limit, err := strconv.Atoi(limitToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value: %s", limitToken.Literal)
		}
		stmt.Limit = limit
	}

	return stmt, nil
}

func (p *Parser) parseInsertStatement() (*InsertStatement, error) {
	stmt := &InsertStatement{}

	// Parse INTO
	if !p.expectKeyword("INTO") {
		return nil, fmt.Errorf("expected INTO")
	}

	// Parse table name
	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	// Parse column list
	if p.lexer.Peek().Type == TokenLeftParen {
		p.lexer.Next() // consume (
		columns, err := p.parseIdentifierList()
		if err != nil {
			return nil, err
		}
		stmt.Columns = columns
		if !p.expectToken(TokenRightParen) {
			return nil, fmt.Errorf("expected )")
		}
	}

	// Parse VALUES
	if !p.expectKeyword("VALUES") {
		return nil, fmt.Errorf("expected VALUES")
	}

	// Parse value lists
	values, err := p.parseValueLists()
	if err != nil {
		return nil, err
	}
	stmt.Values = values

	return stmt, nil
}

func (p *Parser) parseUpdateStatement() (*UpdateStatement, error) {
	stmt := &UpdateStatement{Set: make(map[string]Expression)}

	// Parse table name
	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	// Parse SET clause
	if !p.expectKeyword("SET") {
		return nil, fmt.Errorf("expected SET")
	}

	// Parse SET assignments
	for {
		columnToken := p.lexer.Next()
		if columnToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected column name")
		}
		column := columnToken.Literal

		if !p.expectToken(TokenEquals) {
			return nil, fmt.Errorf("expected =")
		}

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Set[column] = value

		if p.lexer.Peek().Type == TokenComma {
			p.lexer.Next() // consume comma
		} else {
			break
		}
	}

	// Parse WHERE clause
	if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "WHERE" {
		p.lexer.Next() // consume WHERE
		where, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

func (p *Parser) parseDeleteStatement() (*DeleteStatement, error) {
	stmt := &DeleteStatement{}

	// Parse FROM
	if !p.expectKeyword("FROM") {
		return nil, fmt.Errorf("expected FROM")
	}

	// Parse table name
	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	// Parse WHERE clause
	if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "WHERE" {
		p.lexer.Next() // consume WHERE
		where, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

func (p *Parser) parseCreateStatement() (*CreateTableStatement, error) {
	stmt := &CreateTableStatement{}

	// Parse TABLE
	if !p.expectKeyword("TABLE") {
		return nil, fmt.Errorf("expected TABLE")
	}

	// Parse table name
	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	// Parse column definitions
	if !p.expectToken(TokenLeftParen) {
		return nil, fmt.Errorf("expected (")
	}

	columns, err := p.parseColumnDefinitions()
	if err != nil {
		return nil, err
	}
	stmt.Columns = columns

	if !p.expectToken(TokenRightParen) {
		return nil, fmt.Errorf("expected )")
	}

	return stmt, nil
}

func (p *Parser) parseDropStatement() (*DropTableStatement, error) {
	stmt := &DropTableStatement{}

	// Parse TABLE
	if !p.expectKeyword("TABLE") {
		return nil, fmt.Errorf("expected TABLE")
	}

	// Parse table name
	tableToken := p.lexer.Next()
	if tableToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = tableToken.Literal

	return stmt, nil
}

func (p *Parser) parseFieldList() ([]Expression, error) {
	var fields []Expression

	for {
		field, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)

		if p.lexer.Peek().Type == TokenComma {
			p.lexer.Next() // consume comma
		} else {
			break
		}
	}

	return fields, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	return p.parseBinaryExpression(0)
}

func (p *Parser) parseBinaryExpression(precedence int) (Expression, error) {
	left, err := p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}

	for {
		operator := p.lexer.Peek()
		if !isBinaryOperator(operator.Literal) {
			break
		}

		opPrecedence := getOperatorPrecedence(operator.Literal)
		if opPrecedence <= precedence {
			break
		}

		p.lexer.Next() // consume operator
		right, err := p.parseBinaryExpression(opPrecedence)
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseUnaryExpression() (Expression, error) {
	token := p.lexer.Peek()

	switch token.Type {
	case TokenIdentifier:
		p.lexer.Next()
		return &Identifier{Value: token.Literal}, nil
	case TokenString:
		p.lexer.Next()
		return &StringLiteral{Value: token.Literal}, nil
	case TokenNumber:
		p.lexer.Next()
		value, err := strconv.ParseFloat(token.Literal, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", token.Literal)
		}
		return &NumberLiteral{Value: value}, nil
	case TokenKeyword:
		keyword := strings.ToUpper(token.Literal)
		p.lexer.Next()
		switch keyword {
		case "TRUE":
			return &BooleanLiteral{Value: true}, nil
		case "FALSE":
			return &BooleanLiteral{Value: false}, nil
		case "NULL":
			return &NullLiteral{}, nil
		default:
			return nil, fmt.Errorf("unexpected keyword: %s", token.Literal)
		}
	case TokenAsterisk:
		p.lexer.Next()
		return &Identifier{Value: "*"}, nil
	case TokenLeftParen:
		p.lexer.Next() // consume (
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.expectToken(TokenRightParen) {
			return nil, fmt.Errorf("expected )")
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Literal)
	}
}

func (p *Parser) parseIdentifierList() ([]string, error) {
	var identifiers []string

	for {
		token := p.lexer.Next()
		if token.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected identifier")
		}
		identifiers = append(identifiers, token.Literal)

		if p.lexer.Peek().Type == TokenComma {
			p.lexer.Next() // consume comma
		} else {
			break
		}
	}

	return identifiers, nil
}

func (p *Parser) parseValueLists() ([][]Expression, error) {
	var valueLists [][]Expression

	for {
		if !p.expectToken(TokenLeftParen) {
			return nil, fmt.Errorf("expected (")
		}

		var values []Expression
		for {
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			values = append(values, value)

			if p.lexer.Peek().Type == TokenComma {
				p.lexer.Next() // consume comma
			} else {
				break
			}
		}

		valueLists = append(valueLists, values)

		if !p.expectToken(TokenRightParen) {
			return nil, fmt.Errorf("expected )")
		}

		if p.lexer.Peek().Type == TokenComma {
			p.lexer.Next() // consume comma
		} else {
			break
		}
	}

	return valueLists, nil
}

func (p *Parser) parseColumnDefinitions() ([]ColumnDefinition, error) {
	var columns []ColumnDefinition

	for {
		// Parse column name
		nameToken := p.lexer.Next()
		if nameToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected column name")
		}

		// Parse column type
		typeToken := p.lexer.Next()
		if typeToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected column type")
		}

		column := ColumnDefinition{
			Name:     nameToken.Literal,
			Type:     typeToken.Literal,
			Nullable: true,
		}

		// Parse NOT NULL if present
		if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "NOT" {
			p.lexer.Next() // consume NOT
			if !p.expectKeyword("NULL") {
				return nil, fmt.Errorf("expected NULL after NOT")
			}
			column.Nullable = false
		}

		// Parse DEFAULT if present
		if p.lexer.Peek().Type == TokenKeyword && strings.ToUpper(p.lexer.Peek().Literal) == "DEFAULT" {
			p.lexer.Next() // consume DEFAULT
			defaultValue, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			column.Default = defaultValue
		}

		columns = append(columns, column)

		if p.lexer.Peek().Type == TokenComma {
			p.lexer.Next() // consume comma
		} else {
			break
		}
	}

	return columns, nil
}

// Helper methods

func (p *Parser) expectKeyword(keyword string) bool {
	token := p.lexer.Peek()
	if token.Type == TokenKeyword && strings.ToUpper(token.Literal) == strings.ToUpper(keyword) {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) expectToken(tokenType TokenType) bool {
	token := p.lexer.Peek()
	if token.Type == tokenType {
		p.lexer.Next()
		return true
	}
	return false
}

func isBinaryOperator(op string) bool {
	operators := []string{"=", "!=", "<>", "<", ">", "<=", ">=", "AND", "OR", "+", "-", "*", "/"}
	for _, operator := range operators {
		if op == operator {
			return true
		}
	}
	return false
}

func getOperatorPrecedence(op string) int {
	switch op {
	case "OR":
		return 1
	case "AND":
		return 2
	case "=", "!=", "<>", "<", ">", "<=", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/":
		return 5
	default:
		return 0
	}
}
