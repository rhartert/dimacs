package dimacs

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
)

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
			desc:    "invalid variable number",
			reader:  strings.NewReader("p cnf a 3"),
			wantCNF: CNFFormula{},
			wantErr: true,
		},
		{
			desc:    "invalid clause number",
			reader:  strings.NewReader("p cnf 3 a"),
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
			desc: "valid cnf",
			reader: strings.NewReader(`
c valid cnf formula
p cnf 3 4 
1 2 3 0
1 -2 3 0
1 -3 0
-2 -3 0
c
`),
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
			gotCNF, gotErr := Read(tc.reader)

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
