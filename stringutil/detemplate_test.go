package stringutil

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func compare(expected, actual map[string]string, t *testing.T) bool {
	if len(expected) != len(actual) {
		t.Errorf("Compared maps have differing sizes. %d != %d", len(expected),
			len(actual))
		return false

	}
	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			t.Errorf("Key '%s' not present in actual.", key)
			return false
		}
		if expectedValue != actualValue {
			t.Errorf("Values differ for key '%s'. '%s' != '%s'", key,
				expectedValue, actualValue)
			return false
		}
	}
	return true
}

func runTest(expected map[string]string, template, rendered string, t *testing.T) {
	actual, err := Detemplate(template, rendered)
	if err != nil {
		t.Error(err.Error())
	}
	if !compare(expected, actual, t) {
		t.FailNow()
	}
}

func TestDetemplateSimple(t *testing.T) {
	expected := map[string]string{"foo": "bar"}
	runTest(expected, "template {{foo}} blah", "template bar blah", t)
}

func TestDetemplateEmptyRender(t *testing.T) {
	assert := assert1.New(t)
	actual, err := Detemplate("template {{foo}} blah", "")
	assert.Nil(actual)
	assert.EqualError(err, NewCharacterMismatchError(0, -1).Error())
}

func TestDetemplateEmptyTemplateAndRender(t *testing.T) {
	assert := assert1.New(t)
	actual, err := Detemplate("", "")
	assert.Equal(map[string]string{}, actual)
	assert.NoError(err)
}

func TestFailOnMismatch(t *testing.T) {
	actual, err := Detemplate("template", "rendered")
	if err == nil {
		t.Errorf("Expected to receive an error. Got '%s'", actual)
	} else if err.Error() != "Character in template at index 0"+
		" does not match rendered character at index 0." {
		t.Errorf("Expected error %s", err.Error())
	}
}

func TestFailOnMiddleMismatch(t *testing.T) {
	actual, err := Detemplate("template {{foo}} bar", "template something something")
	if err == nil {
		t.Errorf("Expected to receive an error. Got '%s'", actual)
	} else if err.Error() != "Character in template at index 17"+
		" does not match rendered character at index 10." {
		t.Errorf("Expected error %s", err.Error())
	}
}

func TestDetemplateSimpleSpaces(t *testing.T) {
	expected := map[string]string{"foo": "bar"}
	runTest(expected, "template {{ foo  }} blah", "template bar blah", t)
}

func TestDetemplateSimplePrefix(t *testing.T) {
	expected := map[string]string{"foo": "bar"}
	runTest(expected, "{{ foo  }} blah", "bar blah", t)
}

func TestDetemplateSimplePostfix(t *testing.T) {
	expected := map[string]string{"foo": "bar"}
	runTest(expected, "blah {{ foo  }}", "blah bar", t)
}

func TestDetemplateMulti(t *testing.T) {
	expected := map[string]string{"foo": "asdf", "bar": "qwer", "baz": "zxcv"}
	runTest(expected, "template {{ foo  }} blah {{ bar }}, and {{baz}}",
		"template asdf blah qwer, and zxcv", t)
}

func TestDetemplateConfusing(t *testing.T) {
	expected := map[string]string{"foo": "blah"}
	runTest(expected, "template {{foo}} blah", "template blah blah", t)
}

func TestUrlPath(t *testing.T) {
	expected := map[string]string{"resourceName": "lights", "resourceId": "1234"}
	template := "/api/1/{{resourceName}}/{{resourceId}}\n"
	route := "/api/1/lights/1234\n"
	runTest(expected, template, route, t)
}

func TestIsTerminalEOL(t *testing.T) {
	b := isTerminal([]rune("something"), []rune("something"))
	if !b {
		t.Error("Should have been terminal.")
	}
}

func TestIsTerminalDelim(t *testing.T) {
	b := isTerminal([]rune("something{{foo}}"), []rune("something barf"))
	if !b {
		t.Error("Should have been terminal")
	}
}
