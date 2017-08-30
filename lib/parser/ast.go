package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/mithrandie/csvq/lib/ternary"
)

const TOKEN_UNDEFINED = 0

func IsNull(v Primary) bool {
	_, ok := v.(Null)
	return ok
}

type Statement interface{}

type Expression interface {
	GetBaseExpr() *BaseExpr
	HasParseInfo() bool
	Line() int
	Char() int
	SourceFile() string
}

type QueryExpression interface {
	String() string

	GetBaseExpr() *BaseExpr
	HasParseInfo() bool
	Line() int
	Char() int
	SourceFile() string
}

type BaseExpr struct {
	line       int
	char       int
	sourceFile string
}

func (e *BaseExpr) Line() int {
	return e.line
}

func (e *BaseExpr) Char() int {
	return e.char
}

func (e *BaseExpr) SourceFile() string {
	return e.sourceFile
}

func (e *BaseExpr) HasParseInfo() bool {
	if e == nil {
		return false
	}
	return true
}

func (e *BaseExpr) GetBaseExpr() *BaseExpr {
	return e
}

func NewBaseExpr(token Token) *BaseExpr {
	return &BaseExpr{
		line:       token.Line,
		char:       token.Char,
		sourceFile: token.SourceFile,
	}
}

type Primary interface {
	String() string
	Ternary() ternary.Value
}

type String struct {
	literal string
}

func (s String) String() string {
	return quoteString(s.literal)
}

func NewString(s string) String {
	return String{
		literal: s,
	}
}

func (s String) Value() string {
	return s.literal
}

func (s String) Ternary() ternary.Value {
	if b, err := strconv.ParseBool(s.Value()); err == nil {
		return ternary.ParseBool(b)
	}
	return ternary.UNKNOWN
}

type Integer struct {
	value int64
}

func NewIntegerFromString(s string) Integer {
	i, _ := strconv.ParseInt(s, 10, 64)
	return Integer{
		value: i,
	}
}

func NewInteger(i int64) Integer {
	return Integer{
		value: i,
	}
}

func (i Integer) String() string {
	return Int64ToStr(i.value)
}

func (i Integer) Value() int64 {
	return i.value
}

func (i Integer) Ternary() ternary.Value {
	switch i.Value() {
	case 0:
		return ternary.FALSE
	case 1:
		return ternary.TRUE
	default:
		return ternary.UNKNOWN
	}
}

type Float struct {
	value float64
}

func NewFloatFromString(s string) Float {
	f, _ := strconv.ParseFloat(s, 64)
	return Float{
		value: f,
	}
}

func NewFloat(f float64) Float {
	return Float{
		value: f,
	}
}

func (f Float) String() string {
	return Float64ToStr(f.value)
}

func (f Float) Value() float64 {
	return f.value
}

func (f Float) Ternary() ternary.Value {
	switch f.Value() {
	case 0:
		return ternary.FALSE
	case 1:
		return ternary.TRUE
	default:
		return ternary.UNKNOWN
	}
}

type Boolean struct {
	value bool
}

func NewBoolean(b bool) Boolean {
	return Boolean{
		value: b,
	}
}

func (b Boolean) String() string {
	return strconv.FormatBool(b.value)
}

func (b Boolean) Value() bool {
	return b.value
}

func (b Boolean) Ternary() ternary.Value {
	return ternary.ParseBool(b.Value())
}

type Ternary struct {
	value ternary.Value
}

func NewTernaryFromString(s string) Ternary {
	t, _ := ternary.Parse(s)
	return Ternary{
		value: t,
	}
}

func NewTernary(t ternary.Value) Ternary {
	return Ternary{
		value: t,
	}
}

func (t Ternary) String() string {
	return t.value.String()
}

func (t Ternary) Ternary() ternary.Value {
	return t.value
}

type Datetime struct {
	value time.Time
}

func NewDatetimeFromString(s string) Datetime {
	t, _ := StrToTime(s)
	return Datetime{
		value: t,
	}
}

func NewDatetime(t time.Time) Datetime {
	return Datetime{
		value: t,
	}
}

func (dt Datetime) String() string {
	return quoteString(dt.value.Format(time.RFC3339Nano))
}

func (dt Datetime) Value() time.Time {
	return dt.value
}

func (dt Datetime) Ternary() ternary.Value {
	return ternary.UNKNOWN
}

func (dt Datetime) Format(s string) string {
	return dt.value.Format(s)
}

type Null struct{}

func NewNull() Null {
	return Null{}
}

func (n Null) String() string {
	return "NULL"
}

func (n Null) Ternary() ternary.Value {
	return ternary.UNKNOWN
}

type PrimitiveType struct {
	*BaseExpr
	Literal string
	Value   Primary
}

func NewStringValue(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewString(s),
	}
}

func NewIntegerValueFromString(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewIntegerFromString(s),
	}
}

func NewIntegerValue(i int64) PrimitiveType {
	return PrimitiveType{
		Value: NewInteger(i),
	}
}

func NewFloatValueFromString(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewFloatFromString(s),
	}
}

func NewFloatValue(f float64) PrimitiveType {
	return PrimitiveType{
		Value: NewFloat(f),
	}
}

func NewTernaryValueFromString(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewTernaryFromString(s),
	}
}

func NewTernaryValue(t ternary.Value) PrimitiveType {
	return PrimitiveType{
		Value: NewTernary(t),
	}
}

func NewDatetimeValueFromString(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewDatetimeFromString(s),
	}
}

func NewDatetimeValue(t time.Time) PrimitiveType {
	return PrimitiveType{
		Value: NewDatetime(t),
	}
}

func NewNullValueFromString(s string) PrimitiveType {
	return PrimitiveType{
		Literal: s,
		Value:   NewNull(),
	}
}

func NewNullValue() PrimitiveType {
	return PrimitiveType{
		Value: NewNull(),
	}
}

func (e PrimitiveType) String() string {
	if 0 < len(e.Literal) {
		switch e.Value.(type) {
		case String, Datetime:
			return quoteString(e.Literal)
		default:
			return e.Literal
		}
	}
	return e.Value.String()
}

func (e PrimitiveType) IsInteger() bool {
	_, ok := e.Value.(Integer)
	return ok
}

type Identifier struct {
	*BaseExpr
	Literal string
	Quoted  bool
}

func (i Identifier) String() string {
	if i.Quoted {
		return quoteIdentifier(i.Literal)
	}
	return i.Literal
}

type FieldReference struct {
	*BaseExpr
	View   Identifier
	Column Identifier
}

func (e FieldReference) String() string {
	s := e.Column.String()
	if 0 < len(e.View.Literal) {
		s = e.View.String() + "." + s
	}
	return s
}

type ColumnNumber struct {
	*BaseExpr
	View   Identifier
	Number Integer
}

func (e ColumnNumber) String() string {
	return e.View.String() + "." + e.Number.String()
}

type Parentheses struct {
	*BaseExpr
	Expr QueryExpression
}

func (p Parentheses) String() string {
	return putParentheses(p.Expr.String())
}

type RowValue struct {
	*BaseExpr
	Value QueryExpression
}

func (e RowValue) String() string {
	return e.Value.String()
}

type ValueList struct {
	*BaseExpr
	Values []QueryExpression
}

func (e ValueList) String() string {
	return putParentheses(listQueryExpressions(e.Values))
}

type RowValueList struct {
	*BaseExpr
	RowValues []QueryExpression
}

func (e RowValueList) String() string {
	return putParentheses(listQueryExpressions(e.RowValues))
}

type SelectQuery struct {
	*BaseExpr
	WithClause    QueryExpression
	SelectEntity  QueryExpression
	OrderByClause QueryExpression
	LimitClause   QueryExpression
	OffsetClause  QueryExpression
}

func (e SelectQuery) String() string {
	s := []string{}
	if e.WithClause != nil {
		s = append(s, e.WithClause.String())
	}
	s = append(s, e.SelectEntity.String())
	if e.OrderByClause != nil {
		s = append(s, e.OrderByClause.String())
	}
	if e.LimitClause != nil {
		s = append(s, e.LimitClause.String())
	}
	if e.OffsetClause != nil {
		s = append(s, e.OffsetClause.String())
	}
	return joinWithSpace(s)
}

type SelectSet struct {
	*BaseExpr
	LHS      QueryExpression
	Operator Token
	All      Token
	RHS      QueryExpression
}

func (e SelectSet) String() string {
	s := []string{e.LHS.String(), e.Operator.Literal}
	if !e.All.IsEmpty() {
		s = append(s, e.All.Literal)
	}
	s = append(s, e.RHS.String())
	return joinWithSpace(s)
}

type SelectEntity struct {
	*BaseExpr
	SelectClause  QueryExpression
	FromClause    QueryExpression
	WhereClause   QueryExpression
	GroupByClause QueryExpression
	HavingClause  QueryExpression
}

func (e SelectEntity) String() string {
	s := []string{e.SelectClause.String()}
	if e.FromClause != nil {
		s = append(s, e.FromClause.String())
	}
	if e.WhereClause != nil {
		s = append(s, e.WhereClause.String())
	}
	if e.GroupByClause != nil {
		s = append(s, e.GroupByClause.String())
	}
	if e.HavingClause != nil {
		s = append(s, e.HavingClause.String())
	}
	return joinWithSpace(s)
}

type SelectClause struct {
	*BaseExpr
	Select   string
	Distinct Token
	Fields   []QueryExpression
}

func (sc SelectClause) IsDistinct() bool {
	return !sc.Distinct.IsEmpty()
}

func (sc SelectClause) String() string {
	s := []string{sc.Select}
	if sc.IsDistinct() {
		s = append(s, sc.Distinct.Literal)
	}
	s = append(s, listQueryExpressions(sc.Fields))
	return joinWithSpace(s)
}

type FromClause struct {
	*BaseExpr
	From   string
	Tables []QueryExpression
}

func (f FromClause) String() string {
	s := []string{f.From, listQueryExpressions(f.Tables)}
	return joinWithSpace(s)
}

type WhereClause struct {
	*BaseExpr
	Where  string
	Filter QueryExpression
}

func (w WhereClause) String() string {
	s := []string{w.Where, w.Filter.String()}
	return joinWithSpace(s)
}

type GroupByClause struct {
	*BaseExpr
	GroupBy string
	Items   []QueryExpression
}

func (gb GroupByClause) String() string {
	s := []string{gb.GroupBy, listQueryExpressions(gb.Items)}
	return joinWithSpace(s)
}

type HavingClause struct {
	*BaseExpr
	Having string
	Filter QueryExpression
}

func (h HavingClause) String() string {
	s := []string{h.Having, h.Filter.String()}
	return joinWithSpace(s)
}

type OrderByClause struct {
	*BaseExpr
	OrderBy string
	Items   []QueryExpression
}

func (ob OrderByClause) String() string {
	s := []string{ob.OrderBy, listQueryExpressions(ob.Items)}
	return joinWithSpace(s)
}

type LimitClause struct {
	*BaseExpr
	Limit   string
	Value   QueryExpression
	Percent string
	With    QueryExpression
}

func (e LimitClause) String() string {
	s := []string{e.Limit, e.Value.String()}
	if e.IsPercentage() {
		s = append(s, e.Percent)
	}
	if e.With != nil {
		s = append(s, e.With.String())
	}
	return joinWithSpace(s)
}

func (e LimitClause) IsPercentage() bool {
	return 0 < len(e.Percent)
}

func (e LimitClause) IsWithTies() bool {
	if e.With == nil {
		return false
	}
	return e.With.(LimitWith).Type.Token == TIES
}

type LimitWith struct {
	*BaseExpr
	With string
	Type Token
}

func (e LimitWith) String() string {
	s := []string{e.With, e.Type.Literal}
	return joinWithSpace(s)
}

type OffsetClause struct {
	*BaseExpr
	Offset string
	Value  QueryExpression
}

func (e OffsetClause) String() string {
	s := []string{e.Offset, e.Value.String()}
	return joinWithSpace(s)
}

type WithClause struct {
	*BaseExpr
	With         string
	InlineTables []QueryExpression
}

func (e WithClause) String() string {
	s := []string{e.With, listQueryExpressions(e.InlineTables)}
	return joinWithSpace(s)
}

type InlineTable struct {
	*BaseExpr
	Recursive Token
	Name      Identifier
	Fields    []QueryExpression
	As        string
	Query     SelectQuery
}

func (e InlineTable) String() string {
	s := []string{}
	if !e.Recursive.IsEmpty() {
		s = append(s, e.Recursive.Literal)
	}
	s = append(s, e.Name.String())
	if e.Fields != nil {
		s = append(s, putParentheses(listQueryExpressions(e.Fields)))
	}
	s = append(s, e.As, putParentheses(e.Query.String()))
	return joinWithSpace(s)
}

func (e InlineTable) IsRecursive() bool {
	return !e.Recursive.IsEmpty()
}

type Subquery struct {
	*BaseExpr
	Query SelectQuery
}

func (sq Subquery) String() string {
	return putParentheses(sq.Query.String())
}

type Comparison struct {
	*BaseExpr
	LHS      QueryExpression
	Operator string
	RHS      QueryExpression
}

func (c Comparison) String() string {
	s := []string{c.LHS.String(), c.Operator, c.RHS.String()}
	return joinWithSpace(s)
}

type Is struct {
	*BaseExpr
	Is       string
	LHS      QueryExpression
	RHS      QueryExpression
	Negation Token
}

func (i Is) IsNegated() bool {
	return !i.Negation.IsEmpty()
}

func (i Is) String() string {
	s := []string{i.LHS.String(), i.Is}
	if i.IsNegated() {
		s = append(s, i.Negation.Literal)
	}
	s = append(s, i.RHS.String())
	return joinWithSpace(s)
}

type Between struct {
	*BaseExpr
	Between  string
	And      string
	LHS      QueryExpression
	Low      QueryExpression
	High     QueryExpression
	Negation Token
}

func (b Between) IsNegated() bool {
	return !b.Negation.IsEmpty()
}

func (b Between) String() string {
	s := []string{b.LHS.String()}
	if b.IsNegated() {
		s = append(s, b.Negation.Literal)
	}
	s = append(s, b.Between, b.Low.String(), b.And, b.High.String())
	return joinWithSpace(s)
}

type In struct {
	*BaseExpr
	In       string
	LHS      QueryExpression
	Values   QueryExpression
	Negation Token
}

func (i In) IsNegated() bool {
	return !i.Negation.IsEmpty()
}

func (i In) String() string {
	s := []string{i.LHS.String()}
	if i.IsNegated() {
		s = append(s, i.Negation.Literal)
	}
	s = append(s, i.In, i.Values.String())
	return joinWithSpace(s)
}

type All struct {
	*BaseExpr
	All      string
	LHS      QueryExpression
	Operator string
	Values   QueryExpression
}

func (a All) String() string {
	s := []string{a.LHS.String(), a.Operator, a.All, a.Values.String()}
	return joinWithSpace(s)
}

type Any struct {
	*BaseExpr
	Any      string
	LHS      QueryExpression
	Operator string
	Values   QueryExpression
}

func (a Any) String() string {
	s := []string{a.LHS.String(), a.Operator, a.Any, a.Values.String()}
	return joinWithSpace(s)
}

type Like struct {
	*BaseExpr
	Like     string
	LHS      QueryExpression
	Pattern  QueryExpression
	Negation Token
}

func (l Like) IsNegated() bool {
	return !l.Negation.IsEmpty()
}

func (l Like) String() string {
	s := []string{l.LHS.String()}
	if l.IsNegated() {
		s = append(s, l.Negation.Literal)
	}
	s = append(s, l.Like, l.Pattern.String())
	return joinWithSpace(s)
}

type Exists struct {
	*BaseExpr
	Exists string
	Query  Subquery
}

func (e Exists) String() string {
	s := []string{e.Exists, e.Query.String()}
	return joinWithSpace(s)
}

type Arithmetic struct {
	*BaseExpr
	LHS      QueryExpression
	Operator int
	RHS      QueryExpression
}

func (a Arithmetic) String() string {
	s := []string{a.LHS.String(), string(rune(a.Operator)), a.RHS.String()}
	return joinWithSpace(s)
}

type UnaryArithmetic struct {
	*BaseExpr
	Operand  QueryExpression
	Operator Token
}

func (e UnaryArithmetic) String() string {
	return e.Operator.Literal + e.Operand.String()
}

type Logic struct {
	*BaseExpr
	LHS      QueryExpression
	Operator Token
	RHS      QueryExpression
}

func (l Logic) String() string {
	s := []string{l.LHS.String(), l.Operator.Literal, l.RHS.String()}
	return joinWithSpace(s)
}

type UnaryLogic struct {
	*BaseExpr
	Operand  QueryExpression
	Operator Token
}

func (e UnaryLogic) String() string {
	if e.Operator.Token == NOT {
		s := []string{e.Operator.Literal, e.Operand.String()}
		return joinWithSpace(s)
	}
	return e.Operator.Literal + e.Operand.String()
}

type Concat struct {
	*BaseExpr
	Items []QueryExpression
}

func (c Concat) String() string {
	s := make([]string, len(c.Items))
	for i, v := range c.Items {
		s[i] = v.String()
	}
	return strings.Join(s, " || ")
}

type Function struct {
	*BaseExpr
	Name string
	Args []QueryExpression
}

func (e Function) String() string {
	return e.Name + "(" + listQueryExpressions(e.Args) + ")"
}

type AggregateFunction struct {
	*BaseExpr
	Name     string
	Distinct Token
	Args     []QueryExpression
}

func (e AggregateFunction) String() string {
	s := []string{}
	if !e.Distinct.IsEmpty() {
		s = append(s, e.Distinct.Literal)
	}
	s = append(s, listQueryExpressions(e.Args))

	return e.Name + "(" + joinWithSpace(s) + ")"
}

func (e AggregateFunction) IsDistinct() bool {
	return !e.Distinct.IsEmpty()
}

type Table struct {
	*BaseExpr
	Object QueryExpression
	As     string
	Alias  QueryExpression
}

func (t Table) String() string {
	s := []string{t.Object.String()}
	if 0 < len(t.As) {
		s = append(s, t.As)
	}
	if t.Alias != nil {
		s = append(s, t.Alias.String())
	}
	return joinWithSpace(s)
}

func (t Table) Name() Identifier {
	if t.Alias != nil {
		return t.Alias.(Identifier)
	}

	if file, ok := t.Object.(Identifier); ok {
		return Identifier{
			BaseExpr: file.BaseExpr,
			Literal:  FormatTableName(file.Literal),
		}
	}

	return Identifier{
		BaseExpr: t.Object.GetBaseExpr(),
		Literal:  t.Object.String(),
	}

}

type Join struct {
	*BaseExpr
	Join      string
	Table     QueryExpression
	JoinTable QueryExpression
	Natural   Token
	JoinType  Token
	Direction Token
	Condition QueryExpression
}

func (j Join) String() string {
	s := []string{j.Table.String()}
	if !j.Natural.IsEmpty() {
		s = append(s, j.Natural.Literal)
	}
	if !j.Direction.IsEmpty() {
		s = append(s, j.Direction.Literal)
	}
	if !j.JoinType.IsEmpty() {
		s = append(s, j.JoinType.Literal)
	}
	s = append(s, j.Join, j.JoinTable.String())
	if j.Condition != nil {
		s = append(s, j.Condition.String())
	}
	return joinWithSpace(s)
}

type JoinCondition struct {
	*BaseExpr
	Literal string
	On      QueryExpression
	Using   []QueryExpression
}

func (jc JoinCondition) String() string {
	var s []string
	if jc.On != nil {
		s = []string{jc.Literal, jc.On.String()}
	} else {
		s = []string{jc.Literal, putParentheses(listQueryExpressions(jc.Using))}
	}
	return joinWithSpace(s)
}

type Field struct {
	*BaseExpr
	Object QueryExpression
	As     string
	Alias  QueryExpression
}

func (f Field) String() string {
	s := []string{f.Object.String()}
	if 0 < len(f.As) {
		s = append(s, f.As)
	}
	if f.Alias != nil {
		s = append(s, f.Alias.String())
	}
	return joinWithSpace(s)
}

func (f Field) Name() string {
	if f.Alias != nil {
		return f.Alias.(Identifier).Literal
	}
	if t, ok := f.Object.(PrimitiveType); ok {
		return t.Literal
	}
	if fr, ok := f.Object.(FieldReference); ok {
		return fr.Column.Literal
	}
	return f.Object.String()
}

type AllColumns struct {
	*BaseExpr
}

func (ac AllColumns) String() string {
	return "*"
}

type Dual struct {
	*BaseExpr
	Dual string
}

func (d Dual) String() string {
	return d.Dual
}

type Stdin struct {
	*BaseExpr
	Stdin string
}

func (si Stdin) String() string {
	return si.Stdin
}

type OrderItem struct {
	*BaseExpr
	Value     QueryExpression
	Direction Token
	Nulls     string
	Position  Token
}

func (e OrderItem) String() string {
	s := []string{e.Value.String()}
	if !e.Direction.IsEmpty() {
		s = append(s, e.Direction.Literal)
	}
	if 0 < len(e.Nulls) {
		s = append(s, e.Nulls, e.Position.Literal)
	}
	return joinWithSpace(s)
}

type CaseExpr struct {
	*BaseExpr
	Case  string
	End   string
	Value QueryExpression
	When  []QueryExpression
	Else  QueryExpression
}

func (e CaseExpr) String() string {
	s := []string{e.Case}
	if e.Value != nil {
		s = append(s, e.Value.String())
	}
	for _, v := range e.When {
		s = append(s, v.String())
	}
	if e.Else != nil {
		s = append(s, e.Else.String())
	}
	s = append(s, e.End)
	return joinWithSpace(s)
}

type CaseExprWhen struct {
	*BaseExpr
	When      string
	Then      string
	Condition QueryExpression
	Result    QueryExpression
}

func (e CaseExprWhen) String() string {
	s := []string{e.When, e.Condition.String(), e.Then, e.Result.String()}
	return joinWithSpace(s)
}

type CaseExprElse struct {
	*BaseExpr
	Else   string
	Result QueryExpression
}

func (e CaseExprElse) String() string {
	s := []string{e.Else, e.Result.String()}
	return joinWithSpace(s)
}

type ListAgg struct {
	*BaseExpr
	ListAgg     string
	Distinct    Token
	Args        []QueryExpression
	WithinGroup string
	OrderBy     QueryExpression
}

func (e ListAgg) String() string {
	option := []string{}
	if !e.Distinct.IsEmpty() {
		option = append(option, e.Distinct.Literal)
	}
	option = append(option, listQueryExpressions(e.Args))

	s := []string{e.ListAgg + "(" + joinWithSpace(option) + ")"}
	if 0 < len(e.WithinGroup) {
		s = append(s, e.WithinGroup)
		if e.OrderBy != nil {
			s = append(s, "("+e.OrderBy.String()+")")
		} else {
			s = append(s, "()")
		}
	}
	return joinWithSpace(s)
}

func (e ListAgg) IsDistinct() bool {
	return !e.Distinct.IsEmpty()
}

type AnalyticFunction struct {
	*BaseExpr
	Name           string
	Distinct       Token
	Args           []QueryExpression
	IgnoreNulls    bool
	IgnoreNullsLit string
	Over           string
	AnalyticClause AnalyticClause
}

func (e AnalyticFunction) String() string {
	option := []string{}
	if !e.Distinct.IsEmpty() {
		option = append(option, e.Distinct.Literal)
	}
	if e.Args != nil {
		option = append(option, listQueryExpressions(e.Args))
	}
	if e.IgnoreNulls {
		option = append(option, e.IgnoreNullsLit)
	}

	s := []string{
		e.Name + "(" + joinWithSpace(option) + ")",
		e.Over,
		"(" + e.AnalyticClause.String() + ")",
	}
	return joinWithSpace(s)
}

func (e AnalyticFunction) IsDistinct() bool {
	return !e.Distinct.IsEmpty()
}

type AnalyticClause struct {
	*BaseExpr
	Partition     QueryExpression
	OrderByClause QueryExpression
}

func (e AnalyticClause) String() string {
	s := []string{}
	if e.Partition != nil {
		s = append(s, e.Partition.String())
	}
	if e.OrderByClause != nil {
		s = append(s, e.OrderByClause.String())
	}
	return joinWithSpace(s)
}

func (e AnalyticClause) PartitionValues() []QueryExpression {
	if e.Partition == nil {
		return nil
	}
	return e.Partition.(Partition).Values
}

type Partition struct {
	*BaseExpr
	PartitionBy string
	Values      []QueryExpression
}

func (e Partition) String() string {
	s := []string{e.PartitionBy, listQueryExpressions(e.Values)}
	return joinWithSpace(s)
}

type Variable struct {
	*BaseExpr
	Name string
}

func (v Variable) String() string {
	return v.Name
}

type VariableSubstitution struct {
	*BaseExpr
	Variable Variable
	Value    QueryExpression
}

func (vs VariableSubstitution) String() string {
	return joinWithSpace([]string{vs.Variable.String(), SUBSTITUTION_OPERATOR, vs.Value.String()})
}

type VariableAssignment struct {
	*BaseExpr
	Variable Variable
	Value    QueryExpression
}

type VariableDeclaration struct {
	*BaseExpr
	Assignments []Expression
}

type DisposeVariable struct {
	*BaseExpr
	Variable Variable
}

type InsertQuery struct {
	*BaseExpr
	WithClause QueryExpression
	Table      QueryExpression
	Fields     []QueryExpression
	ValuesList []QueryExpression
	Query      QueryExpression
}

type UpdateQuery struct {
	*BaseExpr
	WithClause  QueryExpression
	Tables      []QueryExpression
	SetList     []Expression
	FromClause  QueryExpression
	WhereClause QueryExpression
}

type UpdateSet struct {
	*BaseExpr
	Field QueryExpression
	Value QueryExpression
}

type DeleteQuery struct {
	*BaseExpr
	WithClause  QueryExpression
	Tables      []QueryExpression
	FromClause  QueryExpression
	WhereClause QueryExpression
}

type CreateTable struct {
	*BaseExpr
	Table  Identifier
	Fields []QueryExpression
	Query  QueryExpression
}

type AddColumns struct {
	*BaseExpr
	Table    QueryExpression
	Columns  []Expression
	Position Expression
}

type ColumnDefault struct {
	*BaseExpr
	Column Identifier
	Value  QueryExpression
}

type ColumnPosition struct {
	*BaseExpr
	Position Token
	Column   QueryExpression
}

type DropColumns struct {
	*BaseExpr
	Table   QueryExpression
	Columns []QueryExpression
}

type RenameColumn struct {
	*BaseExpr
	Table QueryExpression
	Old   QueryExpression
	New   Identifier
}

type FunctionDeclaration struct {
	*BaseExpr
	Name       Identifier
	Parameters []Expression
	Statements []Statement
}

type AggregateDeclaration struct {
	*BaseExpr
	Name       Identifier
	Cursor     Identifier
	Parameters []Expression
	Statements []Statement
}

type Return struct {
	*BaseExpr
	Value QueryExpression
}

type Print struct {
	*BaseExpr
	Value QueryExpression
}

type Printf struct {
	*BaseExpr
	Format string
	Values []QueryExpression
}

type Source struct {
	*BaseExpr
	FilePath QueryExpression
}

type SetFlag struct {
	*BaseExpr
	Name  string
	Value Primary
}

type If struct {
	*BaseExpr
	Condition  QueryExpression
	Statements []Statement
	ElseIf     []Expression
	Else       Expression
}

type ElseIf struct {
	*BaseExpr
	Condition  QueryExpression
	Statements []Statement
}

type Else struct {
	*BaseExpr
	Statements []Statement
}

type Case struct {
	*BaseExpr
	Value QueryExpression
	When  []Expression
	Else  Expression
}

type CaseWhen struct {
	*BaseExpr
	Condition  QueryExpression
	Statements []Statement
}

type CaseElse struct {
	*BaseExpr
	Statements []Statement
}

type While struct {
	*BaseExpr
	Condition  QueryExpression
	Statements []Statement
}

type WhileInCursor struct {
	*BaseExpr
	Variables  []Variable
	Cursor     Identifier
	Statements []Statement
}

type CursorDeclaration struct {
	*BaseExpr
	Cursor Identifier
	Query  SelectQuery
}

type OpenCursor struct {
	*BaseExpr
	Cursor Identifier
}

type CloseCursor struct {
	*BaseExpr
	Cursor Identifier
}

type DisposeCursor struct {
	*BaseExpr
	Cursor Identifier
}

type FetchCursor struct {
	*BaseExpr
	Position  Expression
	Cursor    Identifier
	Variables []Variable
}

type FetchPosition struct {
	*BaseExpr
	Position Token
	Number   QueryExpression
}

type CursorStatus struct {
	*BaseExpr
	CursorLit string
	Cursor    Identifier
	Is        string
	Negation  Token
	Type      int
	TypeLit   string
}

func (e CursorStatus) String() string {
	s := []string{e.CursorLit, e.Cursor.String(), e.Is}
	if !e.Negation.IsEmpty() {
		s = append(s, e.Negation.Literal)
	}
	s = append(s, e.TypeLit)
	return joinWithSpace(s)
}

type CursorAttrebute struct {
	*BaseExpr
	CursorLit string
	Cursor    Identifier
	Attrebute Token
}

func (e CursorAttrebute) String() string {
	s := []string{e.CursorLit, e.Cursor.String(), e.Attrebute.Literal}
	return joinWithSpace(s)
}

type TableDeclaration struct {
	*BaseExpr
	Table  Identifier
	Fields []QueryExpression
	Query  QueryExpression
}

type DisposeTable struct {
	*BaseExpr
	Table Identifier
}

type TransactionControl struct {
	*BaseExpr
	Token int
}

type FlowControl struct {
	*BaseExpr
	Token int
}

type Trigger struct {
	*BaseExpr
	Token   int
	Message QueryExpression
	Code    Primary
}

func putParentheses(s string) string {
	return "(" + s + ")"
}

func joinWithSpace(s []string) string {
	return strings.Join(s, " ")
}

func listQueryExpressions(exprs []QueryExpression) string {
	s := make([]string, len(exprs))
	for i, v := range exprs {
		s[i] = v.String()
	}
	return strings.Join(s, ", ")
}

func quoteString(s string) string {
	return "'" + s + "'"
}

func quoteIdentifier(s string) string {
	return "`" + s + "`"
}
