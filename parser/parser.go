// Package parser provides parsing of KodiScript tokens into an AST.
package parser

import (
	"fmt"
	"strconv"

	"github.com/issadicko/kodi-script-go/ast"
	"github.com/issadicko/kodi-script-go/lexer"
	"github.com/issadicko/kodi-script-go/token"
)

// Precedence levels for operators
const (
	_ int = iota
	LOWEST
	TERNARY     // ?:  (conditional)
	ELVIS       // ?:
	OR          // ||
	AND         // &&
	EQUALS      // == !=
	LESSGREATER // > < >= <=
	SUM         // + -
	PRODUCT     // * /
	PREFIX      // -X or !X
	CALL        // func(x)
	ACCESS      // . ?.
)

var precedences = map[token.Type]int{
	token.QUESTION:    TERNARY,
	token.ELVIS:       ELVIS,
	token.OR:          OR,
	token.AND:         AND,
	token.EQ:          EQUALS,
	token.NOT_EQ:      EQUALS,
	token.LT:          LESSGREATER,
	token.GT:          LESSGREATER,
	token.LT_EQ:       LESSGREATER,
	token.GT_EQ:       LESSGREATER,
	token.PLUS:        SUM,
	token.MINUS:       SUM,
	token.ASTERISK:    PRODUCT,
	token.PERCENT:     PRODUCT,
	token.SLASH:       PRODUCT,
	token.LPAREN:      CALL,
	token.LBRACKET:    ACCESS,
	token.DOT:         ACCESS,
	token.SAFE_ACCESS: ACCESS,
}

// Parser parses tokens from a lexer into an AST.
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.STRING_TEMPLATE, p.parseStringTemplate)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseObjectLiteral)
	p.registerPrefix(token.FN, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.ELVIS, p.parseElvisExpression)
	p.registerInfix(token.QUESTION, p.parseTernaryExpression)
	p.registerInfix(token.DOT, p.parsePropertyAccess)
	p.registerInfix(token.SAFE_ACCESS, p.parseSafeAccess)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Errors returns the parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, fmt.Sprintf("line %d, col %d: %s", p.curToken.Line, p.curToken.Column, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.addError("expected %s, got %s", t, p.peekToken.Type)
	return false
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// consumeEndOfStatement consumes optional statement terminators (;, NEWLINE).
func (p *Parser) consumeEndOfStatement() {
	for p.curTokenIs(token.SEMICOLON) || p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
}

// ParseProgram parses the entire program.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		p.consumeEndOfStatement()
		if p.curTokenIs(token.EOF) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// Move to the next token after parsing a statement
		// This ensures we don't get stuck on the last token of the expression
		if !p.curTokenIs(token.EOF) && !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.NEWLINE) {
			p.nextToken()
		}
		p.consumeEndOfStatement()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseVarDecl()
	case token.IF:
		return p.parseIfStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.TRY:
		return p.parseTryStatement()
	case token.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case token.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	case token.FN:
		// Named function declaration: fn name(...) { ... }
		if p.peekTokenIs(token.IDENT) {
			return p.parseFunctionDeclaration()
		}
		return p.parseExpressionStatement()
	case token.IDENT:
		switch p.peekToken.Type {
		case token.ASSIGN:
			return p.parseAssignment()
		case token.PLUS_EQ, token.MINUS_EQ, token.ASTERISK_EQ, token.SLASH_EQ:
			return p.parseCompoundAssignment()
		case token.PLUS_PLUS, token.MINUS_MINUS:
			return p.parseIncDec()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseCompoundAssignment desugars `x += e` into `x = x + e` (and -=, *=, /=).
func (p *Parser) parseCompoundAssignment() ast.Statement {
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	startToken := p.curToken

	p.nextToken() // move onto the compound-assign operator
	opTok := p.curToken
	var op string
	switch opTok.Type {
	case token.PLUS_EQ:
		op = "+"
	case token.MINUS_EQ:
		op = "-"
	case token.ASTERISK_EQ:
		op = "*"
	case token.SLASH_EQ:
		op = "/"
	}

	p.nextToken() // move onto the right-hand expression
	right := p.parseExpression(LOWEST)

	return &ast.Assignment{
		Token: startToken,
		Name:  name,
		Value: &ast.BinaryExpr{Token: opTok, Left: name, Operator: op, Right: right},
	}
}

// parseIncDec desugars `x++` into `x = x + 1` (and `x--` into `x = x - 1`).
func (p *Parser) parseIncDec() ast.Statement {
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	startToken := p.curToken

	p.nextToken() // move onto ++ / -- ; caller advances past it
	opTok := p.curToken
	op := "+"
	if opTok.Type == token.MINUS_MINUS {
		op = "-"
	}

	one := &ast.NumberLiteral{Token: token.Token{Type: token.NUMBER, Literal: "1"}, Value: 1}
	return &ast.Assignment{
		Token: startToken,
		Name:  name,
		Value: &ast.BinaryExpr{Token: opTok, Left: name, Operator: op, Right: one},
	}
}

func (p *Parser) parseVarDecl() ast.Statement {
	letTok := p.curToken

	// Destructuring: let [a, b] = expr  /  let {a, b} = expr
	if p.peekTokenIs(token.LBRACKET) {
		return p.parseDestructure(letTok, token.LBRACKET, token.RBRACKET, true)
	}
	if p.peekTokenIs(token.LBRACE) {
		return p.parseDestructure(letTok, token.LBRACE, token.RBRACE, false)
	}

	stmt := &ast.VarDecl{Token: letTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseDestructure parses `let [a, b] = expr` (isArray) or `let {a, b} = expr`.
func (p *Parser) parseDestructure(letTok token.Token, open, close token.Type, isArray bool) ast.Statement {
	p.nextToken() // cur = open bracket/brace

	var names []*ast.Identifier
	if !p.peekTokenIs(close) {
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		names = append(names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			names = append(names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		}
	}

	if !p.expectPeek(close) {
		return nil
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	value := p.parseExpression(LOWEST)

	if isArray {
		return &ast.ArrayDestructure{Token: letTok, Names: names, Value: value}
	}
	return &ast.ObjectDestructure{Token: letTok, Names: names, Value: value}
}

// parseReturnStatement parses: return [expr]
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// Check if there's an expression after return
	// If next token is a statement terminator or EOF, return without value
	if p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.NEWLINE) || p.peekTokenIs(token.EOF) || p.peekTokenIs(token.RBRACE) {
		return stmt
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseForStatement parses: for (variable in iterable) { body }
func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Expect (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Expect identifier (loop variable)
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect 'in'
	if !p.expectPeek(token.IN) {
		return nil
	}

	// Parse iterable expression
	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	// Expect )
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Expect {
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse body
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseWhileStatement parses: while (condition) { body }
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	// Expect (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Parse condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	// Expect )
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Expect {
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse body
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseTryStatement parses: try { body } catch [ (e) ] { handler }
func (p *Parser) parseTryStatement() ast.Statement {
	stmt := &ast.TryStatement{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	// Allow a newline between the try block's `}` and `catch`.
	for p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}
	if !p.expectPeek(token.CATCH) {
		return nil
	}

	// Optional error variable: catch (e) { ... } or catch { ... }
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // cur = (
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.CatchVar = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Catch = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseAssignment() *ast.Assignment {
	stmt := &ast.Assignment{Token: p.curToken}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // consume ASSIGN
	p.nextToken() // move to expression

	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // cur = else

		if p.peekTokenIs(token.IF) {
			// else if: parse the nested if and wrap it as the alternative block
			p.nextToken() // cur = if
			nested := p.parseIfStatement()
			stmt.Alternative = &ast.BlockStatement{
				Token:      p.curToken,
				Statements: []ast.Statement{nested},
			}
		} else {
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // consume the opening brace

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		p.consumeEndOfStatement()
		if p.curTokenIs(token.RBRACE) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken() // move past the statement
		p.consumeEndOfStatement()
	}

	// Note: we leave curToken on RBRACE, the caller should advance if needed

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addError("no prefix parse function for %s", p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) &&
		!p.peekTokenIs(token.NEWLINE) &&
		!p.peekTokenIs(token.EOF) &&
		precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addError("could not parse %q as number", p.curToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseStringTemplate parses a string template like "Hello ${name}!"
// The literal contains the raw template with ${...} markers.
// We parse it into alternating string parts and expressions.
func (p *Parser) parseStringTemplate() ast.Expression {
	template := &ast.StringTemplate{Token: p.curToken}
	template.Parts = []ast.Expression{}

	literal := p.curToken.Literal
	i := 0

	for i < len(literal) {
		// Find next ${
		start := i
		for i < len(literal) && !(i+1 < len(literal) && literal[i] == '$' && literal[i+1] == '{') {
			i++
		}

		// Add string part if non-empty
		if i > start {
			part := &ast.StringLiteral{
				Token: token.Token{Type: token.STRING, Literal: literal[start:i]},
				Value: literal[start:i],
			}
			template.Parts = append(template.Parts, part)
		}

		// If we found ${, parse the expression
		if i+1 < len(literal) && literal[i] == '$' && literal[i+1] == '{' {
			i += 2 // skip ${

			// Find matching }
			braceCount := 1
			exprStart := i
			for i < len(literal) && braceCount > 0 {
				if literal[i] == '{' {
					braceCount++
				} else if literal[i] == '}' {
					braceCount--
				}
				if braceCount > 0 {
					i++
				}
			}

			// Extract and parse the expression
			exprStr := literal[exprStart:i]
			if i < len(literal) {
				i++ // skip closing }
			}

			// Create a new lexer and parser for the expression
			exprLexer := lexer.New(exprStr)
			exprParser := New(exprLexer)
			expr := exprParser.parseExpression(LOWEST)

			if len(exprParser.errors) > 0 {
				p.errors = append(p.errors, exprParser.errors...)
			}

			if expr != nil {
				template.Parts = append(template.Parts, expr)
			}
		}
	}

	return template
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpr{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseListElement())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseListElement())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseListElement parses one element of an array literal or argument list,
// allowing a spread element (...expr).
func (p *Parser) parseListElement() ast.Expression {
	if p.curTokenIs(token.ELLIPSIS) {
		tok := p.curToken
		p.nextToken()
		return &ast.SpreadExpr{Token: tok, Value: p.parseExpression(LOWEST)}
	}
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseObjectLiteral() ast.Expression {
	object := &ast.ObjectLiteral{Token: p.curToken}
	object.Pairs = make(map[string]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return object
	}

	for !p.peekTokenIs(token.RBRACE) {
		// Skip newlines before a key (supports multi-line object literals)
		for p.peekTokenIs(token.NEWLINE) {
			p.nextToken()
		}
		if p.peekTokenIs(token.RBRACE) {
			break
		}

		p.nextToken()

		// Support both string "key" and identifier key
		var key string
		if p.curTokenIs(token.STRING) || p.curTokenIs(token.IDENT) {
			key = p.curToken.Literal
		} else {
			p.errors = append(p.errors, "expected string or identifier as object key")
			return nil
		}

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		object.Pairs[key] = value

		// Separator can be COMMA or NEWLINE (or the closing RBRACE)
		if !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.NEWLINE) {
			p.addError("expected comma, newline or }")
			return nil
		}

		// Consume a comma separator if present; newlines are skipped at loop top
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return object
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpr{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpr{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	expr := &ast.TernaryExpr{Token: p.curToken, Condition: condition}

	p.nextToken() // move onto the consequent
	expr.Consequent = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken() // move onto the alternative
	expr.Alternative = p.parseExpression(LOWEST)

	return expr
}

func (p *Parser) parseElvisExpression(left ast.Expression) ast.Expression {
	expression := &ast.ElvisExpr{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken()
	expression.Default = p.parseExpression(ELVIS)

	return expression
}

func (p *Parser) parsePropertyAccess(left ast.Expression) ast.Expression {
	expression := &ast.PropertyAccessExpr{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Property = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return expression
}

func (p *Parser) parseSafeAccess(left ast.Expression) ast.Expression {
	expression := &ast.SafeAccessExpr{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Property = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return expression
}

// parseFunctionDeclaration desugars `fn name(a, b) { ... }` into
// `let name = fn(a, b) { ... }` (a VarDecl), which also makes recursion work
// because the name is bound in the closure's environment.
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	fnToken := p.curToken // the 'fn' token

	p.nextToken() // move onto the function name
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	lit := &ast.FunctionLiteral{Token: fnToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()

	return &ast.VarDecl{Token: fnToken, Name: name, Value: lit}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpr{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseListElement())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseListElement())
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}
