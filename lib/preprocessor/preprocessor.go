package preprocessor

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrCircularDependency = errors.New("a circular dependency found")

// The preprocessor is responsible for expanding "include <file>"

// Run the preprocessor against a configuration file.
func Run(f *os.File, stack []*os.File) ([]byte, error) {
	for _, sf := range stack {
		if sf.Name() == f.Name() {
			return []byte{}, ErrCircularDependency
		}
	}

	stack = append(stack, f)

	var b []byte
	dir, err := fileDir(f)
	if err != nil {
		return []byte{}, err
	}

	// read the lines of the file
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		l := scanner.Bytes()
		if strings.HasPrefix(string(l), "include ") {
			r := regexp.MustCompile(`include "(?P<f>[^"]*)"$`)
			incf := r.FindStringSubmatch(string(l))[1]

			// TODO(jh) 20241230: handle absolute paths
			fh, err := os.Open(filepath.Join(dir, incf))
			if err != nil {
				return []byte{}, err
			}

			by, err := Run(fh, stack)
			if err != nil {
				return []byte{}, err
			}

			b = append(b, by...)
			//f = string(l)
		} else {
			b = append(b, l...)
			b = append(b, '\n')
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return []byte{}, err
	}

	// recurse for any include lines

	return b, nil
}

func fileDir(f *os.File) (string, error) {
	absP, err := filepath.Abs(f.Name())
	if err != nil {
		return "", err
	}

	return filepath.Dir(absP), nil
}
