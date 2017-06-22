package query

import (
	"strings"
	"unicode/utf8"

	"github.com/mithrandie/csvq/lib/parser"
	"github.com/mithrandie/csvq/lib/ternary"
)

type ComparisonResult int

const (
	EQUAL ComparisonResult = iota
	NOT_EQUAL
	LESS
	GREATER
	INCOMMENSURABLE
)

var comparisonResultLiterals = map[ComparisonResult]string{
	EQUAL:           "EQUAL",
	NOT_EQUAL:       "NOT_EQUAL",
	LESS:            "LESS",
	GREATER:         "GREATER",
	INCOMMENSURABLE: "INCOMMENSURABLE",
}

func (cr ComparisonResult) String() string {
	return comparisonResultLiterals[cr]
}

func CompareCombinedly(p1 parser.Primary, p2 parser.Primary) ComparisonResult {
	if parser.IsNull(p1) || parser.IsNull(p2) {
		return INCOMMENSURABLE
	}

	if f1 := parser.PrimaryToFloat(p1); !parser.IsNull(f1) {
		if f2 := parser.PrimaryToFloat(p2); !parser.IsNull(f2) {
			v1 := f1.(parser.Float).Value()
			v2 := f2.(parser.Float).Value()
			if v1 == v2 {
				return EQUAL
			} else if v1 < v2 {
				return LESS
			} else {
				return GREATER
			}
		}
	}

	if d1 := parser.PrimaryToDatetime(p1); !parser.IsNull(d1) {
		if d2 := parser.PrimaryToDatetime(p2); !parser.IsNull(d2) {
			v1 := d1.(parser.Datetime).Value()
			v2 := d2.(parser.Datetime).Value()
			if v1.Equal(v2) {
				return EQUAL
			} else if v1.Before(v2) {
				return LESS
			} else {
				return GREATER
			}
		}
	}

	if b1 := parser.PrimaryToBoolean(p1); !parser.IsNull(b1) {
		if b2 := parser.PrimaryToBoolean(p2); !parser.IsNull(b2) {
			v1 := b1.(parser.Boolean).Bool()
			v2 := b2.(parser.Boolean).Bool()
			if v1 == v2 {
				return EQUAL
			} else {
				return NOT_EQUAL
			}
		}
	}

	if s1, ok := p1.(parser.String); ok {
		if s2, ok := p2.(parser.String); ok {
			v1 := strings.ToUpper(s1.Value())
			v2 := strings.ToUpper(s2.Value())

			if v1 == v2 {
				return EQUAL
			} else if v1 < v2 {
				return LESS
			} else {
				return GREATER
			}
		}
	}

	return INCOMMENSURABLE
}

func EqualTo(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE {
		return ternary.ParseBool(r == EQUAL)
	}
	return ternary.UNKNOWN
}

func NotEqualTo(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE {
		return ternary.ParseBool(r != EQUAL)
	}
	return ternary.UNKNOWN
}

func LessThan(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE && r != NOT_EQUAL {
		return ternary.ParseBool(r == LESS)
	}
	return ternary.UNKNOWN
}

func GreaterThan(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE && r != NOT_EQUAL {
		return ternary.ParseBool(r == GREATER)
	}
	return ternary.UNKNOWN
}

func LessThanOrEqualTo(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE && r != NOT_EQUAL {
		return ternary.ParseBool(r != GREATER)
	}
	return ternary.UNKNOWN
}

func GreaterThanOrEqualTo(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if r := CompareCombinedly(p1, p2); r != INCOMMENSURABLE && r != NOT_EQUAL {
		return ternary.ParseBool(r != LESS)
	}
	return ternary.UNKNOWN
}

func Compare(p1 parser.Primary, p2 parser.Primary, operator string) ternary.Value {
	switch operator {
	case "=":
		return EqualTo(p1, p2)
	case ">":
		return GreaterThan(p1, p2)
	case "<":
		return LessThan(p1, p2)
	case ">=":
		return GreaterThanOrEqualTo(p1, p2)
	case "<=":
		return LessThanOrEqualTo(p1, p2)
	default: //case "<>", "!=":
		return NotEqualTo(p1, p2)
	}
}

func EquivalentTo(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if parser.IsNull(p1) && parser.IsNull(p2) {
		return ternary.TRUE
	}
	return EqualTo(p1, p2)
}

func Is(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	return p1.Ternary().EqualTo(p2.Ternary())
}

func Between(p parser.Primary, low parser.Primary, high parser.Primary) ternary.Value {
	return ternary.And(GreaterThanOrEqualTo(p, low), LessThanOrEqualTo(p, high))

}

func Like(p1 parser.Primary, p2 parser.Primary) ternary.Value {
	if parser.IsNull(p1) || parser.IsNull(p2) {
		return ternary.UNKNOWN
	}

	s1 := parser.PrimaryToString(p1)
	if parser.IsNull(s1) {
		return ternary.UNKNOWN
	}
	s2 := parser.PrimaryToString(p2)
	if parser.IsNull(s2) {
		return ternary.UNKNOWN
	}

	s := strings.ToUpper(p1.(parser.String).Value())
	pattern := strings.ToUpper(p2.(parser.String).Value())

	if s == pattern {
		return ternary.TRUE
	}
	if len(pattern) < 1 {
		return ternary.FALSE
	}

	patternRunes := []rune(pattern)
	patternPos := 0

	for {
		anyRunesMinLen, anyRunexMaxLen, search, pos := stringPattern(patternRunes, patternPos)
		patternPos = pos

		anyString := s
		if 0 < len(search) {
			idx := strings.Index(s, search)
			if idx < 0 {
				return ternary.FALSE
			}
			anyString = s[:idx]
		}

		if utf8.RuneCountInString(anyString) < anyRunesMinLen {
			return ternary.FALSE
		}
		if -1 < anyRunexMaxLen && anyRunexMaxLen < utf8.RuneCountInString(anyString) {
			return ternary.FALSE
		}

		if len(patternRunes) <= patternPos {
			break
		}

		s = s[len(anyString+search):]
	}

	return ternary.TRUE
}

func stringPattern(pattern []rune, position int) (int, int, string, int) {
	anyRunesMinLen := 0
	anyRunesMaxLen := 0
	search := []rune{}
	returnPostion := position

	escaped := false
	for i := position; i < len(pattern); i++ {
		r := pattern[i]

		if escaped {
			switch r {
			case '%', '_':
				search = append(search, r)
			default:
				search = append(search, '\\', r)
			}
			returnPostion++
			escaped = false
			continue
		}

		if (r == '%' || r == '_') && 0 < len(search) {
			break
		}
		returnPostion++

		switch r {
		case '%':
			anyRunesMaxLen = -1
		case '_':
			anyRunesMinLen++
			if -1 < anyRunesMaxLen {
				anyRunesMaxLen++
			}
		case '\\':
			escaped = true
		default:
			search = append(search, r)
		}
	}
	if escaped {
		search = append(search, '\\')
	}

	return anyRunesMinLen, anyRunesMaxLen, string(search), returnPostion
}

func Any(p parser.Primary, list []parser.Primary, operator string) ternary.Value {
	result := ternary.FALSE

	for _, v := range list {
		r := Compare(p, v, operator)
		if r == ternary.TRUE {
			result = ternary.TRUE
			break
		}
		if result == ternary.FALSE && r == ternary.UNKNOWN {
			result = ternary.UNKNOWN
		}
	}
	return result
}

func All(p parser.Primary, list []parser.Primary, operator string) ternary.Value {
	result := ternary.TRUE

	for _, v := range list {
		r := Compare(p, v, operator)
		if r == ternary.FALSE {
			result = ternary.FALSE
			break
		}
		if result == ternary.TRUE && r == ternary.UNKNOWN {
			result = ternary.UNKNOWN
		}
	}
	return result
}
