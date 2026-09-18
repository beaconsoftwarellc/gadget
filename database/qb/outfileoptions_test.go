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
				Location: "foo.csv",
				Format:   "CSV",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT CSV",
		},
		{
			name: "format with header",
			options: &OutfileOptions{
				Location: "foo.csv",
				Format:   "CSV",
				Header:   true,
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT CSV HEADER",
		},
		{
			name: "format text with header",
			options: &OutfileOptions{
				Location: "foo.csv",
				Format:   "TEXT",
				Header:   true,
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT TEXT HEADER",
		},
		{
			name: "header without format",
			options: &OutfileOptions{
				Location: "foo.csv",
				Header:   true,
			},
			expected: "INTO OUTFILE 'foo.csv'",
		},
		{
			name: "terminated by only",
			options: &OutfileOptions{
				Location:     "foo.csv",
				TerminatedBy: ",",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS TERMINATED BY ','",
		},
		{
			name: "enclosed by only",
			options: &OutfileOptions{
				Location:   "foo.csv",
				EnclosedBy: "\"",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS OPTIONALLY ENCLOSED BY '\"'",
		},
		{
			name: "escaped by only",
			options: &OutfileOptions{
				Location:  "foo.csv",
				EscapedBy: "\\",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS ESCAPED BY '\\'",
		},
		{
			name: "all field options",
			options: &OutfileOptions{
				Location:     "foo.csv",
				TerminatedBy: ",",
				EnclosedBy:   "\"",
				EscapedBy:    "\\",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\'",
		},
		{
			name: "lines starting by only",
			options: &OutfileOptions{
				Location:        "foo.csv",
				LinesStartingBy: ">",
			},
			expected: "INTO OUTFILE 'foo.csv' LINES STARTING BY '>'",
		},
		{
			name: "lines terminated by only",
			options: &OutfileOptions{
				Location:          "foo.csv",
				LinesTerminatedBy: "\n",
			},
			expected: "INTO OUTFILE 'foo.csv' LINES TERMINATED BY '\n'",
		},
		{
			name: "both lines options",
			options: &OutfileOptions{
				Location:          "foo.csv",
				LinesStartingBy:   ">",
				LinesTerminatedBy: "\n",
			},
			expected: "INTO OUTFILE 'foo.csv' LINES STARTING BY '>' TERMINATED BY '\n'",
		},
		{
			name: "format and fields",
			options: &OutfileOptions{
				Location:     "foo.csv",
				Format:       "CSV",
				TerminatedBy: ",",
				EnclosedBy:   "\"",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT CSV FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"'",
		},
		{
			name: "format and lines",
			options: &OutfileOptions{
				Location:          "foo.csv",
				Format:            "TEXT",
				LinesTerminatedBy: "\n",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT TEXT LINES TERMINATED BY '\n'",
		},
		{
			name: "fields and lines",
			options: &OutfileOptions{
				Location:          "foo.csv",
				TerminatedBy:      ",",
				LinesTerminatedBy: "\n",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS TERMINATED BY ',' LINES TERMINATED BY '\n'",
		},
		{
			name: "all options",
			options: &OutfileOptions{
				Location:          "foo.csv",
				Header:            true,
				Format:            "CSV",
				TerminatedBy:      ",",
				EnclosedBy:        "\"",
				EscapedBy:         "\\",
				LinesStartingBy:   ">",
				LinesTerminatedBy: "\\n",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT CSV HEADER FIELDS TERMINATED BY ',' OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\' LINES STARTING BY '>' TERMINATED BY '\\n'",
		},
		{
			name: "special characters in delimiters",
			options: &OutfileOptions{
				Location:          "foo.csv",
				TerminatedBy:      "\t",
				EnclosedBy:        "'",
				EscapedBy:         "\\",
				LinesTerminatedBy: "\r\n",
			},
			expected: "INTO OUTFILE 'foo.csv' FIELDS TERMINATED BY '\t' OPTIONALLY ENCLOSED BY ''' ESCAPED BY '\\' LINES TERMINATED BY '\r\n'",
		},
		{
			name: "format with terminated by and escaped by",
			options: &OutfileOptions{
				Location:     "foo.csv",
				Format:       "CSV",
				TerminatedBy: ",",
				EscapedBy:    "\\",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT CSV FIELDS TERMINATED BY ',' ESCAPED BY '\\'",
		},
		{
			name: "format with enclosed by and escaped by",
			options: &OutfileOptions{
				Location:   "foo.csv",
				Format:     "TEXT",
				Header:     true,
				EnclosedBy: "\"",
				EscapedBy:  "\\",
			},
			expected: "INTO OUTFILE 'foo.csv' FORMAT TEXT HEADER FIELDS OPTIONALLY ENCLOSED BY '\"' ESCAPED BY '\\'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.SQL()
			assert.Equal(t, tt.expected, result)
		})
	}
}
