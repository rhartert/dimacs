package dimacs

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
)

const validCNF_noComments = `
p cnf 3 4 
1 2 3 0
1 -2 3 0
1 -3 0
-2 -3 0
`

const validCNF_manyComments = `
c comment 1
c comment 2
p cnf 3 4 
c comment 3
1 2 3 0
1 -2 3 0
1 -3 0
c comment 4
-2 -3 0
c comment 5
`

const validCNF_endOfFile = `
p cnf 3 4 
1 2 3 0
1 -2 3 0
1 -3 0
-2 -3 0
%
0
c comment 
`

func TestRead(t *testing.T) {
	testCases := []struct {
		desc    string
		reader  io.Reader
		wantCNF CNFFormula
		wantErr bool
	}{
		{
			desc:    "error reader",
			reader:  iotest.ErrReader(errors.New("test error")),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "empty file",
			reader:  strings.NewReader(""),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "comments only",
			reader:  strings.NewReader("c no problem or clause"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "not a CNF",
			reader:  strings.NewReader("p foo 3 4"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "missing clause number",
			reader:  strings.NewReader("p cnf 3"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid variable number (not a number)",
			reader:  strings.NewReader("p cnf a 3"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid clause number (not a number)",
			reader:  strings.NewReader("p cnf 3 a"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid variable number (negative)",
			reader:  strings.NewReader("p cnf -1 3"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid clause number (negative)",
			reader:  strings.NewReader("p cnf 3 -1"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "duplicate problem lines",
			reader:  strings.NewReader("p cnf 3 4\np cnf 3 4"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "clause before problem line",
			reader:  strings.NewReader("1 2 3 0\np cnf 3 4"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "too many clauses",
			reader:  strings.NewReader("p cnf 3 1\n1 2 3 0\n2 3 0"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "missing clauses",
			reader:  strings.NewReader("p cnf 3 2\n1 2 3 0"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid literal",
			reader:  strings.NewReader("p cnf 3 1\n1 a 3 0"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "literal zero",
			reader:  strings.NewReader("p cnf 3 1\n1 0 3 0"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:   "valid cnf (no comments)",
			reader: strings.NewReader(validCNF_noComments),
			wantCNF: CNFFormula{
				NumVars: 3,
				Clauses: [][]int{
					{1, 2, 3},
					{1, -2, 3},
					{1, -3},
					{-2, -3},
				},
			},
			wantErr: false,
		},
		{
			desc:   "valid cnf (many comments)",
			reader: strings.NewReader(validCNF_manyComments),
			wantCNF: CNFFormula{
				NumVars: 3,
				Clauses: [][]int{
					{1, 2, 3},
					{1, -2, 3},
					{1, -3},
					{-2, -3},
				},
			},
			wantErr: false,
		},
		{
			desc:   "valid cnf (early end of file)",
			reader: strings.NewReader(validCNF_endOfFile),
			wantCNF: CNFFormula{
				NumVars: 3,
				Clauses: [][]int{
					{1, 2, 3},
					{1, -2, 3},
					{1, -3},
					{-2, -3},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotCNF, gotErr := ReadCNF(tc.reader)

			if tc.wantErr && gotErr == nil {
				t.Errorf("Read(): want error, got nil")
			}
			if !tc.wantErr && gotErr != nil {
				t.Errorf("Read(): want no error, got %s", gotErr)
			}
			if diff := cmp.Diff(tc.wantCNF, gotCNF); diff != "" {
				t.Errorf("Read(): CNF mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type testBuilder struct {
	ProblemErr, ClauseErr, CommentErr error
}

func (tb *testBuilder) Problem(_ string, _ int, _ int) error { return tb.ProblemErr }
func (tb *testBuilder) Clause(_ []int) error                 { return tb.ClauseErr }
func (tb *testBuilder) Comment(_ string) error               { return tb.CommentErr }

func errorEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if (a == nil) != (b == nil) {
		return false
	}
	return a.Error() == b.Error()
}

func TestReadBuilder(t *testing.T) {
	testCases := []struct {
		desc    string
		builder Builder
		wantErr error
	}{
		{
			desc:    "problem error",
			builder: &testBuilder{ProblemErr: errors.New("problem error")},
			wantErr: errors.New("problem error"),
		},
		{
			desc:    "clause error",
			builder: &testBuilder{ClauseErr: errors.New("clause error")},
			wantErr: errors.New("clause error"),
		},
		{
			desc:    "comment error",
			builder: &testBuilder{CommentErr: errors.New("comment error")},
			wantErr: errors.New("comment error"),
		},
		{
			desc:    "no error",
			builder: &testBuilder{},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			reader := strings.NewReader(validCNF_manyComments)

			gotErr := ReadBuilder(reader, tc.builder)

			if !errorEqual(gotErr, tc.wantErr) {
				t.Errorf("ReadBuilder(): want error %s, got error %s", tc.wantErr, gotErr)
			}
		})
	}
}
