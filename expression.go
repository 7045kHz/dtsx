// expression.go - SSIS Expression Evaluation Engine
//
// This file implements a comprehensive SSIS expression evaluator that supports:
// - Variable references (@[Namespace::Name])
// - Literals (strings, numbers, booleans, dates)
// - Arithmetic operators (+, -, *, /, %)
// - Comparison operators (==, !=, <, >, <=, >=)
// - Logical operators (&&, ||, !)
// - String concatenation
// - Type casting ((DT_type)expr)
// - Built-in functions (SUBSTRING, UPPER, LOWER, etc.)
// - Parentheses for precedence

package dtsx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SSIS built-in functions
var functions = map[string]func([]interface{}) (interface{}, error){
	// String functions
	"UPPER": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("UPPER expects 1 argument")
		}
		if s, ok := args[0].(string); ok {
			return strings.ToUpper(s), nil
		}
		return nil, fmt.Errorf("UPPER expects string")
	},
	"LOWER": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("LOWER expects 1 argument")
		}
		if s, ok := args[0].(string); ok {
			return strings.ToLower(s), nil
		}
		return nil, fmt.Errorf("LOWER expects string")
	},
	"SUBSTRING": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("SUBSTRING expects 3 arguments")
		}
		s, ok1 := args[0].(string)
		start, ok2 := args[1].(float64)
		length, ok3 := args[2].(float64)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("SUBSTRING expects string, number, number")
		}
		runes := []rune(s)
		startIdx := int(start) - 1 // SSIS is 1-based
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx := startIdx + int(length)
		if endIdx > len(runes) {
			endIdx = len(runes)
		}
		if startIdx >= len(runes) {
			return "", nil
		}
		return string(runes[startIdx:endIdx]), nil
	},
	"LEN": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("LEN expects 1 argument")
		}
		if s, ok := args[0].(string); ok {
			return float64(len([]rune(s))), nil
		}
		return nil, fmt.Errorf("LEN expects string")
	},
	"REPLACE": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("REPLACE expects 3 arguments")
		}
		s, ok1 := args[0].(string)
		old, ok2 := args[1].(string)
		new, ok3 := args[2].(string)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("REPLACE expects string, string, string")
		}
		return strings.ReplaceAll(s, old, new), nil
	},
	// Date functions
	"GETDATE": func(args []interface{}) (interface{}, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("GETDATE expects no arguments")
		}
		return time.Now(), nil
	},
	"YEAR": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("YEAR expects 1 argument")
		}
		if t, ok := args[0].(time.Time); ok {
			return float64(t.Year()), nil
		}
		return nil, fmt.Errorf("YEAR expects date")
	},
	"MONTH": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("MONTH expects 1 argument")
		}
		if t, ok := args[0].(time.Time); ok {
			return float64(t.Month()), nil
		}
		return nil, fmt.Errorf("MONTH expects date")
	},
	"DAY": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("DAY expects 1 argument")
		}
		if t, ok := args[0].(time.Time); ok {
			return float64(t.Day()), nil
		}
		return nil, fmt.Errorf("DAY expects date")
	},
	// Math functions
	"ABS": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("ABS expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			if f < 0 {
				return -f, nil
			}
			return f, nil
		}
		return nil, fmt.Errorf("ABS expects number")
	},
	"CEILING": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("CEILING expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return float64(int(f + 0.999999)), nil // Simple ceiling
		}
		return nil, fmt.Errorf("CEILING expects number")
	},
	"FLOOR": func(args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("FLOOR expects 1 argument")
		}
		if f, ok := args[0].(float64); ok {
			return float64(int(f)), nil
		}
		return nil, fmt.Errorf("FLOOR expects number")
	},
	"DATEADD": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("DATEADD requires 3 arguments")
		}
		datePart, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("DATEADD first argument must be string")
		}
		number, ok := args[1].(float64)
		if !ok {
			return nil, fmt.Errorf("DATEADD second argument must be number")
		}
		date, ok := args[2].(time.Time)
		if !ok {
			return nil, fmt.Errorf("DATEADD third argument must be date")
		}
		switch strings.ToUpper(datePart) {
		case "YEAR", "YY", "YYYY":
			return date.AddDate(int(number), 0, 0), nil
		case "QUARTER", "QQ", "Q":
			return date.AddDate(0, int(number)*3, 0), nil
		case "MONTH", "MM", "M":
			return date.AddDate(0, int(number), 0), nil
		case "DAYOFYEAR", "DY", "Y":
			return date.AddDate(0, 0, int(number)), nil
		case "DAY", "DD", "D":
			return date.AddDate(0, 0, int(number)), nil
		case "WEEK", "WK", "WW":
			return date.AddDate(0, 0, int(number)*7), nil
		case "WEEKDAY", "DW", "W":
			return date.AddDate(0, 0, int(number)), nil
		case "HOUR", "HH":
			return date.Add(time.Hour * time.Duration(number)), nil
		case "MINUTE", "MI", "N":
			return date.Add(time.Minute * time.Duration(number)), nil
		case "SECOND", "SS", "S":
			return date.Add(time.Second * time.Duration(number)), nil
		case "MILLISECOND", "MS":
			return date.Add(time.Millisecond * time.Duration(number)), nil
		default:
			return nil, fmt.Errorf("unknown date part: %s", datePart)
		}
	},
	"DATEDIFF": func(args []interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("DATEDIFF requires 3 arguments")
		}
		datePart, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("DATEDIFF first argument must be string")
		}
		startDate, ok := args[1].(time.Time)
		if !ok {
			return nil, fmt.Errorf("DATEDIFF second argument must be date")
		}
		endDate, ok := args[2].(time.Time)
		if !ok {
			return nil, fmt.Errorf("DATEDIFF third argument must be date")
		}
		duration := endDate.Sub(startDate)
		switch strings.ToUpper(datePart) {
		case "YEAR", "YY", "YYYY":
			years := endDate.Year() - startDate.Year()
			if endDate.YearDay() < startDate.YearDay() {
				years--
			}
			return float64(years), nil
		case "QUARTER", "QQ", "Q":
			quarters := (endDate.Year()-startDate.Year())*4 + (int(endDate.Month())-1)/3 - (int(startDate.Month())-1)/3
			return float64(quarters), nil
		case "MONTH", "MM", "M":
			months := (endDate.Year()-startDate.Year())*12 + int(endDate.Month()) - int(startDate.Month())
			return float64(months), nil
		case "DAYOFYEAR", "DY", "Y":
			return float64(endDate.YearDay() - startDate.YearDay()), nil
		case "DAY", "DD", "D":
			return float64(int(duration.Hours() / 24)), nil
		case "WEEK", "WK", "WW":
			return float64(int(duration.Hours() / (24 * 7))), nil
		case "WEEKDAY", "DW", "W":
			return float64(int(duration.Hours() / 24)), nil
		case "HOUR", "HH":
			return float64(int(duration.Hours())), nil
		case "MINUTE", "MI", "N":
			return float64(int(duration.Minutes())), nil
		case "SECOND", "SS", "S":
			return float64(int(duration.Seconds())), nil
		case "MILLISECOND", "MS":
			return float64(int(duration.Milliseconds())), nil
		default:
			return nil, fmt.Errorf("unknown date part: %s", datePart)
		}
	},
}

func castValue(val interface{}, castType string) (interface{}, error) {
	switch castType {
	case "DT_STR":
		return fmt.Sprintf("%v", val), nil
	case "DT_INT":
		switch v := val.(type) {
		case float64:
			return float64(int(v)), nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return float64(int(f)), nil
			}
		}
		return nil, fmt.Errorf("cannot cast to DT_INT")
	case "DT_DECIMAL", "DT_NUMERIC":
		switch v := val.(type) {
		case float64:
			return v, nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, nil
			}
		}
		return nil, fmt.Errorf("cannot cast to %s", castType)
	case "DT_BOOL":
		switch v := val.(type) {
		case bool:
			return v, nil
		case float64:
			return v != 0, nil
		case string:
			return strings.ToLower(v) == "true" || v == "1", nil
		}
		return nil, fmt.Errorf("cannot cast to DT_BOOL")
	}
	return val, nil // No-op for unknown types
}

// EvaluateExpression evaluates an SSIS expression in the context of a package
func EvaluateExpression(expr string, pkg *Package) (interface{}, error) {
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Get variables
	vars, err := getAllVariables(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to get variables: %v", err)
	}

	// Parse and evaluate the expression
	parsed, err := parseExpression(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %v", err)
	}

	return parsed.Eval(vars)
}

// Expr represents an expression AST node
type Expr interface {
	Eval(vars map[string]interface{}) (interface{}, error)
}

// Literal represents a literal value
type Literal struct {
	Value interface{}
}

func (l *Literal) Eval(vars map[string]interface{}) (interface{}, error) {
	return l.Value, nil
}

// Variable represents a variable reference
type Variable struct {
	Name string
}

func (v *Variable) Eval(vars map[string]interface{}) (interface{}, error) {
	if val, ok := vars[v.Name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("variable not found: %s", v.Name)
}

// BinaryOp represents a binary operation
type BinaryOp struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryOp) Eval(vars map[string]interface{}) (interface{}, error) {
	left, err := b.Left.Eval(vars)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(vars)
	if err != nil {
		return nil, err
	}
	switch b.Op {
	case "+":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l + r, nil
			}
		}
		if l, ok := left.(string); ok {
			if r, ok := right.(string); ok {
				return l + r, nil
			}
		}
		return nil, fmt.Errorf("cannot add %T and %T", left, right)
	case "-":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l - r, nil
			}
		}
		return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
	case "*":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l * r, nil
			}
		}
		return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
	case "/":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				if r == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return l / r, nil
			}
		}
		return nil, fmt.Errorf("cannot divide %T and %T", left, right)
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l < r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case ">":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l > r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case "<=":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l <= r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case ">=":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l >= r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case "&&":
		lb := toBool(left)
		rb := toBool(right)
		return lb && rb, nil
	case "||":
		lb := toBool(left)
		rb := toBool(right)
		return lb || rb, nil
	}
	return nil, fmt.Errorf("unknown operator: %s", b.Op)
}

// toBool converts a value to boolean
func toBool(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return false
	}
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name string
	Args []Expr
}

func (f *FunctionCall) Eval(vars map[string]interface{}) (interface{}, error) {
	// Evaluate arguments
	args := make([]interface{}, len(f.Args))
	for i, arg := range f.Args {
		val, err := arg.Eval(vars)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Call the function
	if fn, ok := functions[f.Name]; ok {
		return fn(args)
	}
	return nil, fmt.Errorf("unknown function: %s", f.Name)
}

// Conditional represents a ternary conditional expression
type Conditional struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func (c *Conditional) Eval(vars map[string]interface{}) (interface{}, error) {
	cond, err := c.Condition.Eval(vars)
	if err != nil {
		return nil, err
	}

	// Convert to boolean
	var condition bool
	switch v := cond.(type) {
	case bool:
		condition = v
	case float64:
		condition = v != 0
	case string:
		condition = v != ""
	default:
		return nil, fmt.Errorf("cannot convert %T to boolean", cond)
	}

	if condition {
		return c.TrueExpr.Eval(vars)
	}
	return c.FalseExpr.Eval(vars)
}

// Cast represents a type cast
type Cast struct {
	Type string
	Expr Expr
}

func (c *Cast) Eval(vars map[string]interface{}) (interface{}, error) {
	val, err := c.Expr.Eval(vars)
	if err != nil {
		return nil, err
	}

	return castValue(val, c.Type)
}

// UnaryOp represents a unary operator
type UnaryOp struct {
	Op   string
	Expr Expr
}

func (u *UnaryOp) Eval(vars map[string]interface{}) (interface{}, error) {
	val, err := u.Expr.Eval(vars)
	if err != nil {
		return nil, err
	}

	switch u.Op {
	case "!":
		var b bool
		switch v := val.(type) {
		case bool:
			b = v
		case float64:
			b = v != 0
		case string:
			b = v != ""
		default:
			return nil, fmt.Errorf("cannot convert %T to boolean", val)
		}
		return !b, nil
	case "-":
		if f, ok := val.(float64); ok {
			return -f, nil
		}
		return nil, fmt.Errorf("cannot negate %T", val)
	}
	return nil, fmt.Errorf("unknown unary operator: %s", u.Op)
}

// Token represents a lexical token
type Token struct {
	Type  string
	Value string
}

// parseExpression parses an SSIS expression into an AST
func parseExpression(expr string) (Expr, error) {
	tokens := tokenize(expr)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}
	parsed, _, err := parseExpr(tokens, 0)
	return parsed, err
}

// tokenize breaks the expression into tokens
func tokenize(expr string) []Token {
	var tokens []Token
	i := 0
	for i < len(expr) {
		switch {
		case expr[i] == '@':
			if i+1 < len(expr) && expr[i+1] == '[' {
				// Variable reference
				start := i
				i += 2
				for i < len(expr) && expr[i] != ']' {
					i++
				}
				if i < len(expr) {
					i++
				}
				tokens = append(tokens, Token{Type: "variable", Value: expr[start:i]})
			} else {
				tokens = append(tokens, Token{Type: "unknown", Value: string(expr[i])})
				i++
			}
		case expr[i] == '"' || expr[i] == '\'':
			// String literal
			quote := expr[i]
			start := i
			i++
			for i < len(expr) && expr[i] != quote {
				if expr[i] == '\\' {
					i++
				}
				i++
			}
			if i < len(expr) {
				i++
			}
			tokens = append(tokens, Token{Type: "string", Value: expr[start:i]})
		case expr[i] >= '0' && expr[i] <= '9' || expr[i] == '.':
			// Number
			start := i
			for i < len(expr) && (expr[i] >= '0' && expr[i] <= '9' || expr[i] == '.') {
				i++
			}
			tokens = append(tokens, Token{Type: "number", Value: expr[start:i]})
		case expr[i] == '+' || expr[i] == '-' || expr[i] == '*' || expr[i] == '/':
			tokens = append(tokens, Token{Type: "operator", Value: string(expr[i])})
			i++
		case expr[i] == '=' && i+1 < len(expr) && expr[i+1] == '=':
			tokens = append(tokens, Token{Type: "operator", Value: "=="})
			i += 2
		case expr[i] == '!' && i+1 < len(expr) && expr[i+1] == '=':
			tokens = append(tokens, Token{Type: "operator", Value: "!="})
			i += 2
		case expr[i] == '<' && i+1 < len(expr) && expr[i+1] == '=':
			tokens = append(tokens, Token{Type: "operator", Value: "<="})
			i += 2
		case expr[i] == '>' && i+1 < len(expr) && expr[i+1] == '=':
			tokens = append(tokens, Token{Type: "operator", Value: ">="})
			i += 2
		case expr[i] == '&' && i+1 < len(expr) && expr[i+1] == '&':
			tokens = append(tokens, Token{Type: "operator", Value: "&&"})
			i += 2
		case expr[i] == '|' && i+1 < len(expr) && expr[i+1] == '|':
			tokens = append(tokens, Token{Type: "operator", Value: "||"})
			i += 2
		case expr[i] == '(':
			// Check for cast: (DT_type)
			if i+3 < len(expr) && expr[i+1] == 'D' && expr[i+2] == 'T' && expr[i+3] == '_' {
				start := i
				i += 4
				for i < len(expr) && expr[i] != ')' {
					i++
				}
				if i < len(expr) {
					i++
				}
				tokens = append(tokens, Token{Type: "cast", Value: expr[start:i]})
			} else {
				tokens = append(tokens, Token{Type: "lparen", Value: "("})
				i++
			}
		case expr[i] == ')':
			tokens = append(tokens, Token{Type: "rparen", Value: ")"})
			i++
		case expr[i] == ',':
			tokens = append(tokens, Token{Type: "comma", Value: ","})
			i++
		case expr[i] == '?':
			tokens = append(tokens, Token{Type: "question", Value: "?"})
			i++
		case expr[i] == ':':
			tokens = append(tokens, Token{Type: "colon", Value: ":"})
			i++
		case expr[i] == '!':
			tokens = append(tokens, Token{Type: "operator", Value: "!"})
			i++
		case expr[i] == '<':
			tokens = append(tokens, Token{Type: "operator", Value: "<"})
			i++
		case expr[i] == '>':
			tokens = append(tokens, Token{Type: "operator", Value: ">"})
			i++
		case (expr[i] >= 'a' && expr[i] <= 'z') || (expr[i] >= 'A' && expr[i] <= 'Z') || expr[i] == '_':
			// Identifier (function name)
			start := i
			for i < len(expr) && ((expr[i] >= 'a' && expr[i] <= 'z') || (expr[i] >= 'A' && expr[i] <= 'Z') || (expr[i] >= '0' && expr[i] <= '9') || expr[i] == '_') {
				i++
			}
			tokens = append(tokens, Token{Type: "identifier", Value: expr[start:i]})
		case expr[i] == ' ' || expr[i] == '\t' || expr[i] == '\n':
			i++
		default:
			tokens = append(tokens, Token{Type: "unknown", Value: string(expr[i])})
			i++
		}
	}
	return tokens
}

// parseExpr parses an expression with precedence
func parseExpr(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseLogicalOr(tokens, pos)
	if err != nil {
		return nil, pos, err
	}

	// Check for conditional
	if pos < len(tokens) && tokens[pos].Type == "question" {
		pos++ // consume ?
		trueExpr, pos, err := parseExpr(tokens, pos)
		if err != nil {
			return nil, pos, err
		}
		if pos >= len(tokens) || tokens[pos].Type != "colon" {
			return nil, pos, fmt.Errorf("expected : in conditional")
		}
		pos++ // consume :
		falseExpr, pos, err := parseExpr(tokens, pos)
		if err != nil {
			return nil, pos, err
		}
		left = &Conditional{Condition: left, TrueExpr: trueExpr, FalseExpr: falseExpr}
	}

	return left, pos, nil
}

// parseLogicalOr parses logical OR
func parseLogicalOr(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseLogicalAnd(tokens, pos)
	if err != nil {
		return nil, pos, err
	}
	for pos < len(tokens) && tokens[pos].Type == "operator" && tokens[pos].Value == "||" {
		op := tokens[pos].Value
		pos++
		right, newPos, err := parseLogicalAnd(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
		pos = newPos
	}
	return left, pos, nil
}

// parseLogicalAnd parses logical AND
func parseLogicalAnd(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseComparison(tokens, pos)
	if err != nil {
		return nil, pos, err
	}
	for pos < len(tokens) && tokens[pos].Type == "operator" && tokens[pos].Value == "&&" {
		op := tokens[pos].Value
		pos++
		right, newPos, err := parseComparison(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
		pos = newPos
	}
	return left, pos, nil
}

// parseComparison parses comparison operators
func parseComparison(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseAddSub(tokens, pos)
	if err != nil {
		return nil, pos, err
	}
	for pos < len(tokens) && tokens[pos].Type == "operator" && (tokens[pos].Value == "==" || tokens[pos].Value == "!=" || tokens[pos].Value == "<" || tokens[pos].Value == ">" || tokens[pos].Value == "<=" || tokens[pos].Value == ">=") {
		op := tokens[pos].Value
		pos++
		right, newPos, err := parseAddSub(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
		pos = newPos
	}
	return left, pos, nil
}

// parseAddSub parses addition and subtraction
func parseAddSub(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseTerm(tokens, pos)
	if err != nil {
		return nil, pos, err
	}
	for pos < len(tokens) && tokens[pos].Type == "operator" && (tokens[pos].Value == "+" || tokens[pos].Value == "-") {
		op := tokens[pos].Value
		pos++
		right, newPos, err := parseTerm(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
		pos = newPos
	}
	return left, pos, nil
}

// parseTerm parses multiplication and division
func parseTerm(tokens []Token, pos int) (Expr, int, error) {
	left, pos, err := parseFactor(tokens, pos)
	if err != nil {
		return nil, pos, err
	}
	for pos < len(tokens) && tokens[pos].Type == "operator" && (tokens[pos].Value == "*" || tokens[pos].Value == "/") {
		op := tokens[pos].Value
		pos++
		right, newPos, err := parseFactor(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		left = &BinaryOp{Left: left, Op: op, Right: right}
		pos = newPos
	}
	return left, pos, nil
}

// parseFactor parses literals, variables, functions, casts, and unary operators
func parseFactor(tokens []Token, pos int) (Expr, int, error) {
	if pos >= len(tokens) {
		return nil, pos, fmt.Errorf("unexpected end of expression")
	}

	token := tokens[pos]

	// Handle unary operators
	if token.Type == "operator" && (token.Value == "!" || token.Value == "-") {
		pos++
		expr, newPos, err := parseFactor(tokens, pos)
		if err != nil {
			return nil, newPos, err
		}
		return &UnaryOp{Op: token.Value, Expr: expr}, newPos, nil
	}

	// Handle casts
	if token.Type == "cast" {
		// Extract type from (DT_type)
		castStr := token.Value
		if len(castStr) > 4 && castStr[0] == '(' && castStr[len(castStr)-1] == ')' {
			castType := castStr[1 : len(castStr)-1]
			pos++
			expr, newPos, err := parseFactor(tokens, pos)
			if err != nil {
				return nil, newPos, err
			}
			return &Cast{Type: castType, Expr: expr}, newPos, nil
		}
	}

	pos++
	switch token.Type {
	case "number":
		if val, err := strconv.ParseFloat(token.Value, 64); err == nil {
			return &Literal{Value: val}, pos, nil
		}
		return nil, pos, fmt.Errorf("invalid number: %s", token.Value)
	case "string":
		// Remove quotes
		val := token.Value[1 : len(token.Value)-1]
		// Handle escapes if needed
		return &Literal{Value: val}, pos, nil
	case "variable":
		// Remove @[ and ]
		name := token.Value[2 : len(token.Value)-1]
		return &Variable{Name: name}, pos, nil
	case "identifier":
		// Function call
		if pos < len(tokens) && tokens[pos].Type == "lparen" {
			pos++ // consume (
			var args []Expr
			for pos < len(tokens) && tokens[pos].Type != "rparen" {
				arg, newPos, err := parseExpr(tokens, pos)
				if err != nil {
					return nil, newPos, err
				}
				args = append(args, arg)
				pos = newPos
				if pos < len(tokens) && tokens[pos].Type == "comma" {
					pos++ // consume ,
				} else if pos < len(tokens) && tokens[pos].Type != "rparen" {
					return nil, pos, fmt.Errorf("expected , or ) in function call")
				}
			}
			if pos >= len(tokens) || tokens[pos].Type != "rparen" {
				return nil, pos, fmt.Errorf("expected ) in function call")
			}
			pos++ // consume )
			return &FunctionCall{Name: token.Value, Args: args}, pos, nil
		}
		return nil, pos, fmt.Errorf("unexpected identifier: %s", token.Value)
	case "lparen":
		// Parenthesized expression
		expr, newPos, err := parseExpr(tokens, pos-1) // -1 because we already consumed (
		if err != nil {
			return nil, newPos, err
		}
		if newPos >= len(tokens) || tokens[newPos].Type != "rparen" {
			return nil, newPos, fmt.Errorf("expected )")
		}
		return expr, newPos + 1, nil
	default:
		return nil, pos, fmt.Errorf("unexpected token: %v", token)
	}
}

// getAllVariables extracts all variables from the package as a map
func getAllVariables(pkg *Package) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	if pkg == nil || pkg.Variables == nil || pkg.Variables.Variable == nil {
		return vars, nil
	}
	for _, v := range pkg.Variables.Variable {
		if v.NamespaceAttr == nil || v.ObjectNameAttr == nil {
			continue
		}
		fullName := *v.NamespaceAttr + "::" + *v.ObjectNameAttr
		var value interface{}
		if v.VariableValue != nil {
			// Try to parse as number
			if num, err := strconv.ParseFloat(v.VariableValue.Value, 64); err == nil {
				value = num
			} else {
				value = v.VariableValue.Value
			}
		} else {
			// From properties
			for _, prop := range v.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "Value" {
					if num, err := strconv.ParseFloat(prop.Value, 64); err == nil {
						value = num
					} else {
						value = prop.Value
					}
					break
				}
			}
		}
		if value != nil {
			vars[fullName] = value
		}
	}
	return vars, nil
}

// evaluateSimpleExpression provides basic variable substitution (deprecated, use EvaluateExpression)
func evaluateSimpleExpression(expr string, pkg *Package) (interface{}, error) {
	// Fallback to old method
	vars, err := getAllVariables(pkg)
	if err != nil {
		return nil, err
	}
	result := expr
	for name, val := range vars {
		placeholder := fmt.Sprintf("@[%s]", name)
		if str, ok := val.(string); ok {
			result = strings.ReplaceAll(result, placeholder, str)
		} else if num, ok := val.(float64); ok {
			result = strings.ReplaceAll(result, placeholder, strconv.FormatFloat(num, 'f', -1, 64))
		}
	}
	// Try to parse as number
	if num, err := strconv.ParseFloat(strings.TrimSpace(result), 64); err == nil {
		return num, nil
	}
	return result, nil
}
