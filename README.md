# DIMACS CNF Parser

A Go library to parse DIMACS CNF files. This library provides easy-to-use 
functions for reading CNF formulas from various input sources. It also allows 
for custom parsing logic for advanced use cases.

## Features

- Parse CNF formulas from DIMACS files.
- Easy support for gzipped and other compressed file formats.
- Customizable parsing logic to directly integrate with SAT solvers.
- Lightweight and efficient.

## Installation

To install the library, use go get:

```sh
go get github.com/yourusername/dimacs
```

## Examples

### Reading a CNF file

For most clients, the parser can be used to extract the CNF formula from a 
DIMACS CNF file as follows:

```go
package main

import (
    "io"
    "log"
    "os"

    "github.com/rhartert/dimacs"
)

func readCNFFile(filePath string) (*dimacs.CNFFormula, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    cnf, err := dimacs.ReadCNF(file)
    if err != nil {
        return nil, err
    }

    return &cnf, nil
}

func main() {
    cnf, err := readDIMACS("path/to/your/file.cnf")
    if err != nil {
        log.Fatalf("Failed to read CNF: %v", err)
    }
    log.Printf("Read CNF: %+v", cnf)
}
```

### Reading a compressed CNF file

You can easily handle compressed files (e.g. gzipped) by providing 
`dimacs.ReadCNF` with the appropriate reader. For example, here's a modified
program to read a gzipped CNF file:

```go
package main

import (
    "compress/gzip"
    "io"
    "log"
    "os"

    "github.com/rhartert/dimacs"
)

func readGzipped(filePath string) (*dimacs.CNFFormula, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader, err := gzip.NewReader(file)
    if err != nil {
        return nil, err
    }
    defer reader.Close()

    cnf, err := dimacs.ReadCNF(reader)
    if err != nil {
        return nil, err
    }

    return &cnf, nil
}

func main() {
    cnf, err := readGzipped("path/to/your/file.cnf.gz")
    if err != nil {
        log.Fatalf("Failed to read gzipped CNF: %v", err)
    }
    log.Printf("Read gzipped CNF: %+v", cnf)
}
```

### Interfacing with a Solver

In some cases, it is preferable to build your own `dimacs.Builder` to 
interpret the DIMACS file. This has two main advantages. First, it allows you 
to use custom validation rules while also exctracting content from the comment 
lines (e.g. solver parameters). The second is that it allows for a more memory
efficient use of the library by avoiding the creation of a `CNFFormula`.

Here's an example of a custom builder to interface a `SATSolver` with the
library.

```go
// SATSolver is an example interface to a SAT solver.
type SATSolver interface {
    AddVariable()
    AddClause(clause []int)
}

// LoadDIMACS reads a DIMACS file from the given reader and directly add it 
// to the solver. 
func LoadDIMACS(reader io.Reader, solver SATSolver) error {
    b := &builder{solver}
    return dimacs.ReadBuilder(reader, b)
}

// builder wraps the solver to implement dimacs.Builder.
type builder struct {
    solver SATSolver
}

func (b *builder) Problem(nVars int, nClauses int) error {
    for i := 0; i < nVars; i++ {
        b.solver.AddVariable()
    }
    return nil
}

func (b *builder) Clause(tmpClause []int) error {
    clause := make([]int, len(tmpClause))
    copy(clause, tmpClause)
    b.solver.AddClause(clause)
    return nil
}

func (b *builder) Comment(_ string) error {
    return nil // ignore comments
}
```

## Contributing

Contributions are welcomed! Feel free to open an issue or submit a pull request.
