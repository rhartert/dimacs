// Package dimacs provides utilities to read and parse CNF (Conjunctive Normal
// Form) DIMACS files.
package dimacs

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CNFFormula represents a Conjunctive Normal Form (CNF) formula with a specific
// number of variables and a set of clauses. The clauses are represented in
// accordance with the DIMACS CNF specification. Each variable is denoted by an
// integer from 1 to NumVars (inclusive), where a positive integer i represents
// the positive literal of variable i, and a negative integer -i represents the
// negative literal of variable i.
type CNFFormula struct {
	NumVars int
	Clauses [][]int
}

// ReadCNF parses and returns a DIMACS CNF formula from the given reader.
func ReadCNF(r io.Reader) (CNFFormula, error) {
	builder := cnfBuilder{}
	if err := ReadBuilder(r, &builder); err != nil {
		return CNFFormula{}, err
	}
	if builder.cnf == nil {
		return CNFFormula{}, fmt.Errorf("missing problem line found")
	}
	if got, want := len(builder.cnf.Clauses), cap(builder.cnf.Clauses); got < want {
		return CNFFormula{}, fmt.Errorf("missing clauses: expected %d, got %d", want, got)
	}
	return *builder.cnf, nil
}

type cnfBuilder struct {
	cnf *CNFFormula
}

func (b *cnfBuilder) Problem(p string, v int, c int) error {
	if b.cnf != nil {
		return fmt.Errorf("duplicate problem line")
	}
	if p != "cnf" {
		return fmt.Errorf("expected \"cnf\" problem, got %q", p)
	}
	if v < 0 {
		return fmt.Errorf("number of variables must be non-negative, got: %d", v)
	}
	if c < 0 {
		return fmt.Errorf("number of clauses must be non-negative, got: %d", c)
	}
	b.cnf = &CNFFormula{
		NumVars: v,
		Clauses: make([][]int, 0, c),
	}
	return nil
}

func (b *cnfBuilder) Clause(tmp []int) error {
	if b.cnf == nil {
		return fmt.Errorf("clause found before problem line")
	}
	if s := len(b.cnf.Clauses); s == cap(b.cnf.Clauses) {
		return fmt.Errorf("too many clauses: expected %d", s)
	}
	c := make([]int, len(tmp))
	copy(c, tmp)
	b.cnf.Clauses = append(b.cnf.Clauses, c)
	return nil
}

func (b *cnfBuilder) Comment(c string) error { return nil } // ignore comments

// Builder defines methods to construct a CNF formula from a DIMACS file.
type Builder interface {
	// Problem processes the problem line.
	Problem(problem string, nVars int, nClauses int) error

	// Clause processes the clause from clause line. Implementations of this
	// method should consider tmpClause as a shared buffer and only read from it
	// without retaining it.
	Clause(tmpClause []int) error

	// Comment processes a comment line. Lines passed to this function always
	// start with the comment prefix "c". This is useful to process additional
	// information stored in the comments (e.g. problem information, solver
	// configuration, etc.).
	Comment(line string) error
}

// ReadBuilder reads a DIMACS file from the given reader and populates
// the given builder. Builder methods are called in the same order as the
// corresponding lines (i.e. comment, problem, clause) in the DIMACS file.
func ReadBuilder(r io.Reader, b Builder) error {
	scanner := bufio.NewScanner(r)
	clauseBuf := make([]int, 32)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line[0] {
		case 'c':
			if err := b.Comment(line); err != nil {
				return err
			}
		case 'p':
			parts := strings.Fields(line)
			if len(parts) != 4 {
				return fmt.Errorf("problem line should have 4 parts, got %d: %s", len(parts), line)
			}
			nVars, err := strconv.Atoi(parts[2])
			if err != nil {
				return fmt.Errorf("invalid number of variables: %w", err)
			}
			nClauses, err := strconv.Atoi(parts[3])
			if err != nil {
				return fmt.Errorf("invalid number of clauses: %w", err)
			}
			if err := b.Problem(parts[1], nVars, nClauses); err != nil {
				return err
			}
		default:
			clauseBuf = clauseBuf[:0]
			parts := strings.Fields(line)
			for i, p := range parts {
				l, err := strconv.Atoi(p)
				if err != nil {
					return fmt.Errorf("invalid literal in clause %q: %w", line, err)
				}
				if l == 0 {
					if i != len(parts)-1 {
						return fmt.Errorf("zero found before end of clause line: %q", line)
					}
					break
				}
				clauseBuf = append(clauseBuf, l)
			}
			if err := b.Clause(clauseBuf); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
