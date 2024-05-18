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

// Builder defines methods to construct a CNF formula from a DIMACS file.
type Builder interface {
	// Problem processes the problem line.
	Problem(nVars int, nClauses int)

	// Clause processes the clause from clause line. Implementations of this
	// method should consider tmpClause as a shared buffer and only read from it
	// without retaining it.
	Clause(tmpClause []int)

	// Comment processes a comment line. Lines passed to this function always
	// start with the comment prefix "c". This is useful to process additional
	// information stored in the comments (e.g. problem information, solver
	// configuration, etc.).
	Comment(line string)
}

// Read parses a DIMACS CNF file from the given reader and returns a CNFFormula.
func Read(r io.Reader) (CNFFormula, error) {
	builder := cnfBuilder{}
	if err := ReadBuilder(r, &builder); err != nil {
		return CNFFormula{}, err
	}
	return CNFFormula(builder), nil
}

// cnfBuilder wraps CNFFormula to implement the Builder interface.
type cnfBuilder CNFFormula

func (cnf *cnfBuilder) Problem(v int, c int) {
	cnf.NumVars = v
	cnf.Clauses = make([][]int, 0, c)
}

func (cnf *cnfBuilder) Clause(tmpClause []int) {
	c := make([]int, len(tmpClause))
	copy(c, tmpClause)
	cnf.Clauses = append(cnf.Clauses, c)
}

func (cnf *cnfBuilder) Comment(c string) {} // ignore comments

// ReadBuilder reads a DIMACS CNF file from the given reader and populates
// the given builder. Builder methods are called in the same order as the
// corresponding lines (e.g. comment, problem, clause) in the DIMACS file.
func ReadBuilder(r io.Reader, b Builder) error {
	scanner := bufio.NewScanner(r)
	foundProblemLine := false
	clauseBuf := make([]int, 32)
	nClauses := 0
	parsedClauses := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line[0] {
		case 'c':
			b.Comment(line)
		case 'p':
			if foundProblemLine {
				return fmt.Errorf("duplicate problem line: %q", line)
			}
			parts := strings.Fields(line)
			if len(parts) != 4 || parts[1] != "cnf" {
				return fmt.Errorf("invalid problem line: %q", line)
			}
			nVars, err := strconv.Atoi(parts[2])
			if err != nil {
				return fmt.Errorf("invalid number of variables: %w", err)
			}
			nClauses, err = strconv.Atoi(parts[3]) // set nClauses outside the loop
			if err != nil {
				return fmt.Errorf("invalid number of clauses: %w", err)
			}
			b.Problem(nVars, nClauses)
			foundProblemLine = true
		default:
			if !foundProblemLine {
				return fmt.Errorf("clause found before problem line")
			}
			if parsedClauses >= nClauses {
				return fmt.Errorf("too many clauses: expected %d", nClauses)
			}
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
			b.Clause(clauseBuf)
			parsedClauses++
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if !foundProblemLine {
		return fmt.Errorf("no problem line found")
	}
	if parsedClauses != nClauses {
		return fmt.Errorf("mismatched clause count: expected %d, got %d", nClauses, parsedClauses)
	}

	return nil
}
