package planetscale

import "github.com/upper/db/v4"

func andExpressions(expr ...*db.RawExpr) []interface{} {
	output := make([]interface{}, len(expr))
	for i, rawExpr := range expr {
		output[i] = interface{}(rawExpr)
	}
	return output
}
