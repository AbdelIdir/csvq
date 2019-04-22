package query

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/mithrandie/csvq/lib/parser"
	"github.com/mithrandie/csvq/lib/value"

	"github.com/mithrandie/ternary"
)

type CursorMap struct {
	*SyncMap
}

func NewCursorMap() CursorMap {
	return CursorMap{
		NewSyncMap(),
	}
}

func (m CursorMap) IsEmpty() bool {
	return m.SyncMap == nil
}

func (m CursorMap) Store(name string, val *Cursor) {
	m.store(strings.ToUpper(name), val)
}

func (m CursorMap) Load(name string) (*Cursor, bool) {
	if v, ok := m.load(strings.ToUpper(name)); ok {
		return v.(*Cursor), true
	}
	return nil, false
}

func (m CursorMap) Delete(name string) {
	m.delete(strings.ToUpper(name))
}

func (m CursorMap) Exists(name string) bool {
	return m.exists(strings.ToUpper(name))
}

func (m CursorMap) Declare(expr parser.CursorDeclaration) error {
	if m.Exists(expr.Cursor.Literal) {
		return NewCursorRedeclaredError(expr.Cursor)
	}
	m.Store(expr.Cursor.Literal, NewCursor(expr))
	return nil
}

func (m CursorMap) AddPseudoCursor(name parser.Identifier, values []value.Primary) error {
	if m.Exists(name.Literal) {
		return NewCursorRedeclaredError(name)
	}
	m.Store(name.Literal, NewPseudoCursor(name.Literal, values))
	return nil
}

func (m CursorMap) Dispose(name parser.Identifier) error {
	if cur, ok := m.Load(name.Literal); ok {
		if cur.isPseudo {
			return NewPseudoCursorError(name)
		}
		m.Delete(name.Literal)
		return nil
	}
	return NewUndeclaredCursorError(name)
}

func (m CursorMap) Open(ctx context.Context, scope *ReferenceScope, name parser.Identifier, values []parser.ReplaceValue) error {
	if cur, ok := m.Load(name.Literal); ok {
		return cur.Open(ctx, scope, name, values)
	}
	return NewUndeclaredCursorError(name)
}

func (m CursorMap) Close(name parser.Identifier) error {
	if cur, ok := m.Load(name.Literal); ok {
		return cur.Close(name)
	}
	return NewUndeclaredCursorError(name)
}

func (m CursorMap) Fetch(name parser.Identifier, position int, number int) ([]value.Primary, error) {
	if cur, ok := m.Load(name.Literal); ok {
		return cur.Fetch(name, position, number)
	}
	return nil, NewUndeclaredCursorError(name)
}

func (m CursorMap) IsOpen(name parser.Identifier) (ternary.Value, error) {
	if cur, ok := m.Load(name.Literal); ok {
		return cur.IsOpen(), nil
	}
	return ternary.FALSE, NewUndeclaredCursorError(name)
}

func (m CursorMap) IsInRange(name parser.Identifier) (ternary.Value, error) {
	if cur, ok := m.Load(name.Literal); ok {
		t, err := cur.IsInRange()
		if err != nil {
			return ternary.FALSE, NewCursorClosedError(name)
		}
		return t, nil
	}
	return ternary.FALSE, NewUndeclaredCursorError(name)
}

func (m CursorMap) Count(name parser.Identifier) (int, error) {
	if cur, ok := m.Load(name.Literal); ok {
		i, err := cur.Count()
		if err != nil {
			return 0, NewCursorClosedError(name)
		}
		return i, nil
	}
	return 0, NewUndeclaredCursorError(name)
}

type Cursor struct {
	name      string
	query     parser.SelectQuery
	statement parser.Identifier
	view      *View
	index     int
	fetched   bool

	isPseudo bool

	mtx *sync.Mutex
}

func NewCursor(e parser.CursorDeclaration) *Cursor {
	return &Cursor{
		name:      e.Cursor.Literal,
		query:     e.Query,
		statement: e.Statement,
		mtx:       &sync.Mutex{},
	}
}

func NewPseudoCursor(name string, values []value.Primary) *Cursor {
	header := NewHeader("", []string{"c1"})

	records := make(RecordSet, len(values))
	for i, v := range values {
		records[i] = NewRecord([]value.Primary{v})
	}
	view := NewView()
	view.Header = header
	view.RecordSet = records

	return &Cursor{
		name:     name,
		view:     view,
		index:    -1,
		fetched:  false,
		isPseudo: true,
		mtx:      &sync.Mutex{},
	}
}

func (c *Cursor) Open(ctx context.Context, scope *ReferenceScope, name parser.Identifier, values []parser.ReplaceValue) error {
	if c.isPseudo {
		return NewPseudoCursorError(name)
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.view != nil {
		return NewCursorOpenError(name)
	}

	var view *View
	var err error
	if c.query.SelectEntity != nil {
		view, err = Select(ctx, scope, c.query)
	} else {
		prepared, e := scope.Tx.PreparedStatements.Get(c.statement)
		if e != nil {
			return e
		}
		if len(prepared.Statements) != 1 {
			return NewInvalidCursorStatementError(c.statement)
		}
		stmt, ok := prepared.Statements[0].(parser.SelectQuery)
		if !ok {
			return NewInvalidCursorStatementError(c.statement)
		}
		view, err = Select(ContextForPreparedStatement(ctx, NewReplaceValues(values)), scope, stmt)
	}
	if err != nil {
		return err
	}

	c.view = view
	c.index = -1
	c.fetched = false
	return nil
}

func (c *Cursor) Close(name parser.Identifier) error {
	if c.isPseudo {
		return NewPseudoCursorError(name)
	}

	c.mtx.Lock()

	c.view = nil
	c.index = 0
	c.fetched = false

	c.mtx.Unlock()
	return nil
}

func (c *Cursor) Fetch(name parser.Identifier, position int, number int) ([]value.Primary, error) {
	if c.view == nil {
		return nil, NewCursorClosedError(name)
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if !c.fetched {
		c.fetched = true
	}

	switch position {
	case parser.ABSOLUTE:
		c.index = number
	case parser.RELATIVE:
		c.index = c.index + number
	case parser.FIRST:
		c.index = 0
	case parser.LAST:
		c.index = c.view.RecordLen() - 1
	case parser.PRIOR:
		c.index = c.index - 1
	default: // NEXT
		c.index = c.index + 1
	}

	if c.index < 0 {
		c.index = -1
		return nil, nil
	}

	if c.view.RecordLen() <= c.index {
		c.index = c.view.RecordLen()
		return nil, nil
	}

	list := make([]value.Primary, len(c.view.RecordSet[c.index]))
	for i, cell := range c.view.RecordSet[c.index] {
		list[i] = cell.Value()
	}

	return list, nil
}

func (c *Cursor) IsOpen() ternary.Value {
	return ternary.ConvertFromBool(c.view != nil)
}

func (c *Cursor) IsInRange() (ternary.Value, error) {
	if c.view == nil {
		return ternary.FALSE, errors.New("cursor is closed")
	}
	if !c.fetched {
		return ternary.UNKNOWN, nil
	}
	return ternary.ConvertFromBool(-1 < c.index && c.index < c.view.RecordLen()), nil
}

func (c *Cursor) Count() (int, error) {
	if c.view == nil {
		return 0, errors.New("cursor is closed")
	}
	return c.view.RecordLen(), nil
}

func (c *Cursor) Pointer() (int, error) {
	return c.index, nil
}
