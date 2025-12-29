// Package ast defines the Abstract Syntax Tree nodes for KodiScript.
package ast

import (
	"github.com/issadicko/kodi-script-go/token"
)

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// VarDecl represents a variable declaration: let x = expr
type VarDecl struct {
	Token token.Token // the LET token
	Name  *Identifier
	Value Expression
}

func (v *VarDecl) statementNode()       {}
func (v *VarDecl) TokenLiteral() string { return v.Token.Literal }

// Assignment represents an assignment: x = expr
type Assignment struct {
	Token token.Token // the IDENT token
	Name  *Identifier
	Value Expression
}

func (a *Assignment) statementNode()       {}
func (a *Assignment) TokenLiteral() string { return a.Token.Literal }

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// IfStatement represents: if (condition) { consequence } else { alternative }
type IfStatement struct {
	Token       token.Token // the IF token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

// BlockStatement represents a block of statements: { ... }
type BlockStatement struct {
	Token      token.Token // the LBRACE token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// ReturnStatement represents: return [expr]
type ReturnStatement struct {
	Token token.Token // the RETURN token
	Value Expression  // optional, can be nil
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// ForStatement represents: for (variable in iterable) { body }
type ForStatement struct {
	Token    token.Token     // the FOR token
	Variable *Identifier     // loop variable
	Iterable Expression      // expression that produces an array
	Body     *BlockStatement // loop body
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }

// Identifier represents a variable name.
type Identifier struct {
	Token token.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// NumberLiteral represents a numeric value.
type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// BooleanLiteral represents true or false.
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

// NullLiteral represents null.
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }

// BinaryExpr represents a binary operation: left op right
type BinaryExpr struct {
	Token    token.Token // the operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpr) expressionNode()      {}
func (be *BinaryExpr) TokenLiteral() string { return be.Token.Literal }

// UnaryExpr represents a unary operation: op expr
type UnaryExpr struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (ue *UnaryExpr) expressionNode()      {}
func (ue *UnaryExpr) TokenLiteral() string { return ue.Token.Literal }

// SafeAccessExpr represents optional chaining: obj?.property
type SafeAccessExpr struct {
	Token    token.Token // the ?. token
	Object   Expression
	Property *Identifier
}

func (sa *SafeAccessExpr) expressionNode()      {}
func (sa *SafeAccessExpr) TokenLiteral() string { return sa.Token.Literal }

// ElvisExpr represents null coalescing: expr ?: default
type ElvisExpr struct {
	Token   token.Token // the ?: token
	Left    Expression
	Default Expression
}

func (ee *ElvisExpr) expressionNode()      {}
func (ee *ElvisExpr) TokenLiteral() string { return ee.Token.Literal }

// PropertyAccessExpr represents property access: obj.property
type PropertyAccessExpr struct {
	Token    token.Token // the . token
	Object   Expression
	Property *Identifier
}

func (pa *PropertyAccessExpr) expressionNode()      {}
func (pa *PropertyAccessExpr) TokenLiteral() string { return pa.Token.Literal }

// CallExpr represents a function call: func(args...)
type CallExpr struct {
	Token     token.Token // the LPAREN token
	Function  Expression  // Identifier or expression
	Arguments []Expression
}

func (ce *CallExpr) expressionNode()      {}
func (ce *CallExpr) TokenLiteral() string { return ce.Token.Literal }
