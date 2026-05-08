package qb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutfileOptions_SQL(t *testing.T) {
	tests := []struct {
		name     string
		options  *OutfileOptions
		expected string
	}{
		{
			name:     "nil options",
			options:  nil,
			expected: "",
		},
		{
			name:     "empty options",
			options:  &OutfileOptions{},
			expected: "",
		},
		{
			name: "format only without header",
			options: &OutfileOptions{
				Format: "CSV",
			},
			expected: "FORMAT CSV",
		},
		{
			name: "format with header",
			options: &OutfileOptions{
				Format: "CSV",
				Header: true,
			},
			expected: "FORMAT CSV HEADER",
		},
		{
			name: "format text with header",
			options: &OutfileOptions{
				Format: "TEXT",
				Header: true,
			},
			expected: "FORMAT TEXT HEADER",
		},
		{
			name: "header without format",
			options: &OutfileOptions{
				Header: true,
			},
			expected: "",
		},
		{
			name: "terminated by only",
			options: &OutfileOptions{
				TerminatedBy: ",",
			},
			expected: "FIELDS TERMINATED BY ','",
		},
		{
			name: "enclosed by only",
			options: &OutfileOptions{
				EnclosedBy: "\"",
			},
			expected: "FIELDS OPTIONALLY ENCLOSED BY '\"'",
		},
		{
			name: "escaped by only",
			options: &OutfileOptions{
				EscapedBy: "\\",
			},
			expected: "FIELDS ESCAPED BY '\\'",
		},
		{
			name: "all field options",
			options: &OutfileOptions{
				TerminatedBy: ",",
				EnclosedBy:   "\"",
				EscapedBy:    "\\",
			},
			expected: "FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\'",
		},
		{
			name: "lines starting by only",
			options: &OutfileOptions{
				LinesStartingBy: ">",
			},
			expected: "LINES STARTING BY '>'",
		},
		{
			name: "lines terminated by only",
			options: &OutfileOptions{
				LinesTerminatedBy: "\n",
			},
			expected: "LINES TERMINATED BY '\n'",
		},
		{
			name: "both lines options",
			options: &OutfileOptions{
				LinesStartingBy:   ">",
				LinesTerminatedBy: "\n",
			},
			expected: "LINES STARTING BY '>' TERMINATED BY '\n'",
		},
		{
			name: "format and fields",
			options: &OutfileOptions{
				Format:       "CSV",
				TerminatedBy: ",",
				EnclosedBy:   "\"",
			},
			expected: "FORMAT CSV FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"'",
		},
		{
			name: "format and lines",
			options: &OutfileOptions{
				Format:            "TEXT",
				LinesTerminatedBy: "\n",
			},
			expected: "FORMAT TEXT LINES TERMINATED BY '\n'",
		},
		{
			name: "fields and lines",
			options: &OutfileOptions{
				TerminatedBy:      ",",
				LinesTerminatedBy: "\n",
			},
			expected: "FIELDS TERMINATED BY ',' LINES TERMINATED BY '\n'",
		},
		{
			name: "all options",
			options: &OutfileOptions{
				Header:            true,
				Format:            "CSV",
				TerminatedBy:      ",",
				EnclosedBy:        "\"",
				EscapedBy:         "\\",
				LinesStartingBy:   ">",
				LinesTerminatedBy: "\\n",
			},
			expected: "FORMAT CSV HEADER FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\' LINES STARTING BY '>' TERMINATED BY '\\n'",
		},
		{
			name: "special characters in delimiters",
			options: &OutfileOptions{
				TerminatedBy:      "\t",
				EnclosedBy:        "'",
				EscapedBy:         "\\",
				LinesTerminatedBy: "\r\n",
			},
			expected: "FIELDS TERMINATED BY '\t' OPTIONALLY ENCLOSED BY ''' ESCAPED BY '\\' LINES TERMINATED BY '\r\n'",
		},
		{
			name: "format with terminated by and escaped by",
			options: &OutfileOptions{
				Format:       "CSV",
				TerminatedBy: ",",
				EscapedBy:    "\\",
			},
			expected: "FORMAT CSV FIELDS TERMINATED BY ',' ESCAPED BY '\\'",
		},
		{
			name: "format with enclosed by and escaped by",
			options: &OutfileOptions{
				Format:     "TEXT",
				Header:     true,
				EnclosedBy: "\"",
				EscapedBy:  "\\",
			},
			expected: "FORMAT TEXT HEADER FIELDS OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.SQL()
			assert.Equal(t, tt.expected, result)
		})
	}
}
