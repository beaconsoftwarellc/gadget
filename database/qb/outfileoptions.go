package qb

import (
	"fmt"
	"strings"

	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type OutfileFormat string

const (
	// FormatCSV specifies and outfile of comma-separated values.
	FormatCSV OutfileFormat = "CSV"
	// FormatTEXT specifies an outfile of text.
	FormatTEXT OutfileFormat = "TEXT"
)

// OutfileOptions for the query.
// See: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Integrating.SaveIntoS3.html
type OutfileOptions struct {
	// Header indicates whether the first row of the query results should be treated as a header row.
	Header bool
	// Format is the format of the output file (CSV | TEXT)
	Format OutfileFormat
	// TerminatedBy specifies the character sequence used to separate fields in the output file.
	TerminatedBy string
	// EnclosedBy specifies the character sequence used to enclose fields in the output file.
	EnclosedBy string
	// EscapedBy specifies the character sequence used to escape special characters in the output file.
	EscapedBy string
	// LinesStartingBy specifies the character sequence used to start a new line in the output file.
	LinesStartingBy string
	// LinesTerminatedBy specifies the character sequence used to terminate a line in the output file.
	LinesTerminatedBy string
}

// SQL representation of the passed OutfileOptions.
func (oo *OutfileOptions) SQL() string {
	if oo == nil {
		return ""
	}
	var sql []string
	if !stringutil.IsEmpty(string(oo.Format)) {
		s := fmt.Sprintf("FORMAT %s", oo.Format)
		if oo.Header {
			s = fmt.Sprintf("%s HEADER", s)
		}
		sql = append(sql, s)
	}
	if !stringutil.IsEmpty(oo.TerminatedBy + oo.EnclosedBy + oo.EscapedBy) {
		s := "FIELDS"
		if !stringutil.IsEmpty(oo.TerminatedBy) {
			s = fmt.Sprintf("%s TERMINATED BY '%s'", s, oo.TerminatedBy)
		}
		if !stringutil.IsEmpty(oo.EnclosedBy) {
			s = fmt.Sprintf("%s OPTIONALLY ENCLOSED BY '%s'", s, oo.EnclosedBy)
		}
		if !stringutil.IsEmpty(oo.EscapedBy) {
			s = fmt.Sprintf("%s ESCAPED BY '%s'", s, oo.EscapedBy)
		}
		sql = append(sql, s)
	}
	if !stringutil.IsEmpty(oo.LinesStartingBy + oo.LinesTerminatedBy) {
		s := "LINES"
		if !stringutil.IsEmpty(oo.LinesStartingBy) {
			s = fmt.Sprintf("%s STARTING BY '%s'", s, oo.LinesStartingBy)
		}

		if !stringutil.IsEmpty(oo.LinesTerminatedBy) {
			s = fmt.Sprintf("%s TERMINATED BY '%s'", s, oo.LinesTerminatedBy)
		}
		sql = append(sql, s)
	}
	return strings.Join(sql, " ")
}
