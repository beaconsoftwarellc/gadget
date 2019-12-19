package stringutil

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// Detemplate takes a pattern with template style markup and a rendered
// template and attempts to construct the context that would be fed to generate
// the rendering.
// NOTE: This function does not support sequential template variables.
// e.g. 'test {{id1}}{{id2}}' as there is no way of knowing where one would end
//      and the next would begin.
func Detemplate(template string, rendered string) (map[string]string, error) {
	// traverse template looking for first non-matching character
	// ensure the character is a '{'
	// if not abort
	// otherwise extract the name of the variable by looking for the
	// first non-escaped '}'
	context := make(map[string]string)
	if len(template) == 0 && len(rendered) == 0 {
		return context, nil
	}
	if len(rendered) == 0 {
		return nil, NewCharacterMismatchError(0, -1)
	}
	var err error
	runeTemplate := []rune(template)
	runeRendered := []rune(rendered)
	var variableName, variableValue string
	// i + 1 since we check for runeTemplate[i+1] in the if clause
	for i, j := 0, 0; i+DOpenLen < len(runeTemplate); i, j = i+1, j+1 {
		if string(runeTemplate[i:i+DOpenLen]) == DOpen {
			variableName, i = extractTemplateName(runeTemplate, i)
			variableValue, j = extractTemplateValue(runeTemplate, runeRendered,
				i, j)
			context[variableName] = variableValue
		} else if runeTemplate[i] != runeRendered[j] {
			err = NewCharacterMismatchError(i, j)
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return context, nil
}

const (
	// DOpen is the character sequence used to determine when a
	// block of templated variable starts.
	DOpen = "{{"
	// DClose is the character sequence used to determine when a
	// block of templated variable closes.
	DClose = "}}"
	// DOpenLen is the length of the DOpen characters
	DOpenLen = len(DOpen)
	// DCloseLen is the length of the DClose characters
	DCloseLen = len(DClose)
)

// CharacterMismatchError returned when the characters in a rendered template
// do not match those in the template outside of a phrase.
type CharacterMismatchError struct {
	TemplateIndex int
	RenderedIndex int
	trace         []string
}

// NewCharacterMismatchError occurs during demplating and indicates where in the template the issue happened
func NewCharacterMismatchError(templateIndex int, renderedIndex int) errors.TracerError {
	return &CharacterMismatchError{
		TemplateIndex: templateIndex,
		RenderedIndex: renderedIndex,
		trace:         errors.GetStackTrace(),
	}
}

func (err *CharacterMismatchError) Error() string {
	return fmt.Sprintf("Character in template at index %d does not match "+
		"rendered character at index %d.", err.TemplateIndex, err.RenderedIndex)
}

// Trace returns the stack trace for the error
func (err *CharacterMismatchError) Trace() []string {
	return err.trace
}

func isTerminal(template, rendered []rune) bool {
	terminal := true
	for i, j := 0, 0; i+DOpenLen < len(template) && j < len(rendered); i, j = i+1, j+1 {
		// if we found another template open we are done.
		if string(template[i:i+DOpenLen]) == DOpen {
			break
		}
		if template[i] != rendered[j] {
			terminal = false
			break
		}
	}
	return terminal
}

// extractTemplateName takes a string containing a phrase, the index
// specifying the start of the phrase and returns the text in the phrase as
// well as the ending index of the phrase or -1 if no ending was found.
func extractTemplateName(template []rune, startIndex int) (string, int) {
	var name string
	endIndex := startIndex + DCloseLen
	for ; endIndex < len(template); endIndex++ {
		if string(template[endIndex:endIndex+DCloseLen]) == DClose {
			endIndex += DCloseLen
			break
		}
	}
	name = strings.TrimSpace(string(template[startIndex+DOpenLen : endIndex-DCloseLen]))
	return strings.TrimSpace(name), endIndex
}

// extractTemplateValue takes a template and a rendering of that template and
// extracts a value from the the rendered starting at the renderedIndex.
// templateIndex should be the index at which the template phrase ends.
// Returns the new redered index and the value.
func extractTemplateValue(template []rune, rendered []rune, templateIndex int,
	renderedIndex int) (string, int) {
	endValueIndex := renderedIndex
	var terminal bool
	if templateIndex >= len(template) {
		endValueIndex = len(rendered)
	} else {
		for i := renderedIndex; i < len(rendered); i++ {
			if rendered[i] == template[templateIndex] {
				terminal = isTerminal(template[templateIndex:],
					rendered[i:])
				if terminal {
					endValueIndex = i
					break
				}
			}
		}
	}
	return string(rendered[renderedIndex:endValueIndex]), endValueIndex
}
