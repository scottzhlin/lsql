package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/scottlin/lsql/internal/lexer"
)

// Parser converts a token stream into a SelectStmt AST.
type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*SelectStmt, error) {
	stmt := &SelectStmt{Limit: -1}
	var seenWhere, seenGroup, seenOrder, seenLimit bool

	if err := p.expectType(lexer.TOKEN_SELECT, "SELECT"); err != nil {
		return nil, err
	}
	cols, err := p.parseColumns()
	if err != nil {
		return nil, err
	}
	stmt.Columns = cols

	if err := p.expectType(lexer.TOKEN_FROM, "FROM"); err != nil {
		return nil, err
	}
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	stmt.From = path

	if p.peek().Type == lexer.TOKEN_RECURSIVE {
		p.advance()
		stmt.Recursive = true
	}

	for p.peek().Type != lexer.TOKEN_EOF {
		switch p.peek().Type {
		case lexer.TOKEN_WHERE:
			if seenWhere {
				return nil, fmt.Errorf("parse error: duplicate WHERE clause")
			}
			seenWhere = true
			p.advance()
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			stmt.Where = expr
		case lexer.TOKEN_GROUP:
			if seenGroup {
				return nil, fmt.Errorf("parse error: duplicate GROUP BY clause")
			}
			seenGroup = true
			p.advance()
			if err := p.expectType(lexer.TOKEN_BY, "BY"); err != nil {
				return nil, err
			}
			groups, err := p.parseColList()
			if err != nil {
				return nil, err
			}
			stmt.GroupBy = groups
		case lexer.TOKEN_ORDER:
			if seenOrder {
				return nil, fmt.Errorf("parse error: duplicate ORDER BY clause")
			}
			seenOrder = true
			p.advance()
			if err := p.expectType(lexer.TOKEN_BY, "BY"); err != nil {
				return nil, err
			}
			order, err := p.parseOrderList()
			if err != nil {
				return nil, err
			}
			stmt.OrderBy = order
		case lexer.TOKEN_LIMIT:
			if seenLimit {
				return nil, fmt.Errorf("parse error: duplicate LIMIT clause")
			}
			seenLimit = true
			p.advance()
			tok := p.advance()
			if tok.Type != lexer.TOKEN_NUMBER {
				return nil, fmt.Errorf("parse error: expected number after LIMIT, got %q", tok.Literal)
			}
			n, err := strconv.ParseInt(tok.Literal, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse error: LIMIT value must be an integer, got %q", tok.Literal)
			}
			stmt.Limit = int(n)
		default:
			return nil, fmt.Errorf("parse error: unexpected token %q", p.peek().Literal)
		}
	}

	return stmt, nil
}

func (p *Parser) parseColumns() ([]ColExpr, error) {
	if p.peek().Type == lexer.TOKEN_STAR {
		p.advance()
		return []ColExpr{{Col: "*"}}, nil
	}
	var cols []ColExpr
	for {
		col, err := p.parseColExpr()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
		if p.peek().Type != lexer.TOKEN_COMMA {
			break
		}
		p.advance()
	}
	return cols, nil
}

func (p *Parser) parseColExpr() (ColExpr, error) {
	tok := p.peek()
	switch tok.Type {
	case lexer.TOKEN_COUNT, lexer.TOKEN_SUM, lexer.TOKEN_MIN, lexer.TOKEN_MAX:
		agg := strings.ToUpper(tok.Literal)
		p.advance()
		if err := p.expectType(lexer.TOKEN_LPAREN, "("); err != nil {
			return ColExpr{}, err
		}
		var colName string
		if p.peek().Type == lexer.TOKEN_STAR {
			colName = "*"
			p.advance()
		} else if p.peek().Type == lexer.TOKEN_IDENT {
			colName = p.advance().Literal
		} else {
			return ColExpr{}, fmt.Errorf("parse error: expected column name in aggregate function")
		}
		if err := p.expectType(lexer.TOKEN_RPAREN, ")"); err != nil {
			return ColExpr{}, err
		}
		return ColExpr{Agg: agg, Col: colName}, nil
	case lexer.TOKEN_IDENT:
		return ColExpr{Col: p.advance().Literal}, nil
	}
	return ColExpr{}, fmt.Errorf("parse error: expected column expression, got %q", tok.Literal)
}

func (p *Parser) parsePath() (string, error) {
	tok := p.advance()
	if tok.Type != lexer.TOKEN_IDENT && tok.Type != lexer.TOKEN_STRING {
		return "", fmt.Errorf("parse error: expected path after FROM, got %q", tok.Literal)
	}
	return tok.Literal, nil
}

func (p *Parser) parseColList() ([]string, error) {
	var cols []string
	for {
		tok := p.advance()
		if tok.Type != lexer.TOKEN_IDENT {
			return nil, fmt.Errorf("parse error: expected column name, got %q", tok.Literal)
		}
		cols = append(cols, tok.Literal)
		if p.peek().Type != lexer.TOKEN_COMMA {
			break
		}
		p.advance()
	}
	return cols, nil
}

func (p *Parser) parseOrderList() ([]OrderItem, error) {
	var items []OrderItem
	for {
		col, err := p.parseOrderCol()
		if err != nil {
			return nil, err
		}
		item := OrderItem{Col: col}
		if p.peek().Type == lexer.TOKEN_DESC {
			p.advance()
			item.Desc = true
		} else if p.peek().Type == lexer.TOKEN_ASC {
			p.advance()
		}
		items = append(items, item)
		if p.peek().Type != lexer.TOKEN_COMMA {
			break
		}
		p.advance()
	}
	return items, nil
}

// parseOrderCol reads either a plain column name or an aggregate expression for ORDER BY.
func (p *Parser) parseOrderCol() (string, error) {
	switch p.peek().Type {
	case lexer.TOKEN_COUNT, lexer.TOKEN_SUM, lexer.TOKEN_MIN, lexer.TOKEN_MAX:
		col, err := p.parseColExpr()
		if err != nil {
			return "", err
		}
		return col.String(), nil
	case lexer.TOKEN_IDENT:
		return p.advance().Literal, nil
	}
	return "", fmt.Errorf("parse error: expected column name in ORDER BY, got %q", p.peek().Literal)
}

// Expression parsing: OR < AND < NOT < primary (standard SQL precedence)

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_OR {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = LogicalExpr{Left: left, Op: "OR", Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_AND {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = LogicalExpr{Left: left, Op: "AND", Right: right}
	}
	return left, nil
}

func (p *Parser) parseNot() (Expr, error) {
	if p.peek().Type == lexer.TOKEN_NOT {
		p.advance()
		inner, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return NotExpr{Inner: inner}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.peek().Type == lexer.TOKEN_LPAREN {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expectType(lexer.TOKEN_RPAREN, ")"); err != nil {
			return nil, err
		}
		return expr, nil
	}

	colTok := p.advance()
	if colTok.Type != lexer.TOKEN_IDENT {
		return nil, fmt.Errorf("parse error: expected column name, got %q", colTok.Literal)
	}
	col := colTok.Literal

	switch p.peek().Type {
	case lexer.TOKEN_LIKE:
		p.advance()
		val := p.advance()
		if val.Type != lexer.TOKEN_STRING && val.Type != lexer.TOKEN_IDENT {
			return nil, fmt.Errorf("parse error: expected string after LIKE")
		}
		return LikeExpr{Col: col, Pattern: val.Literal}, nil

	case lexer.TOKEN_BETWEEN:
		p.advance()
		low, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if err := p.expectType(lexer.TOKEN_AND, "AND"); err != nil {
			return nil, err
		}
		high, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return BetweenExpr{Col: col, Low: low, High: high}, nil

	default:
		op, err := p.parseOp()
		if err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return BinaryExpr{Col: col, Op: op, Val: val}, nil
	}
}

func (p *Parser) parseOp() (string, error) {
	tok := p.advance()
	switch tok.Type {
	case lexer.TOKEN_EQ:
		return "=", nil
	case lexer.TOKEN_NEQ:
		return "!=", nil
	case lexer.TOKEN_LT:
		return "<", nil
	case lexer.TOKEN_LTE:
		return "<=", nil
	case lexer.TOKEN_GT:
		return ">", nil
	case lexer.TOKEN_GTE:
		return ">=", nil
	}
	return "", fmt.Errorf("parse error: expected comparison operator, got %q", tok.Literal)
}

func (p *Parser) parseValue() (interface{}, error) {
	tok := p.advance()
	switch tok.Type {
	case lexer.TOKEN_STRING:
		if t, err := time.Parse("2006-01-02", tok.Literal); err == nil {
			return t, nil
		}
		if t, err := time.Parse(time.RFC3339, tok.Literal); err == nil {
			return t, nil
		}
		return tok.Literal, nil
	case lexer.TOKEN_NUMBER:
		if n, err := strconv.ParseInt(tok.Literal, 10, 64); err == nil {
			return n, nil
		}
		if f, err := strconv.ParseFloat(tok.Literal, 64); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("parse error: invalid number %q", tok.Literal)
	case lexer.TOKEN_BOOL:
		return strings.ToUpper(tok.Literal) == "TRUE", nil
	case lexer.TOKEN_IDENT:
		return tok.Literal, nil
	}
	return nil, fmt.Errorf("parse error: expected value, got %q", tok.Literal)
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) expectType(tt lexer.TokenType, name string) error {
	tok := p.advance()
	if tok.Type != tt {
		return fmt.Errorf("parse error: expected %s, got %q", name, tok.Literal)
	}
	return nil
}
