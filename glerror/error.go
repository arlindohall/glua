package glerror

import "strings"

type GluaErrorChain struct {
	errors []GluaError
}

func (chain GluaErrorChain) Error() string {
	var errors []string

	for _, err := range chain.errors {
		errors = append(errors, err.Error())
	}

	return strings.Join(errors, "\n")
}

func (chain *GluaErrorChain) IsEmpty() bool {
	return len(chain.errors) == 0
}

func (chain *GluaErrorChain) Append(err error) {
	chain.errors = append(chain.errors, GluaError(err))
}

func (chain *GluaErrorChain) AppendAll(source *GluaErrorChain) {
	for _, err := range source.errors {
		chain.Append(err)
	}
}

type GluaError error
