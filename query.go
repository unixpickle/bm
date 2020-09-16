package main

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/unixpickle/essentials"
)

// MustMatchRecord is like MatchRecords but returns the
// first record and automatically fails with an error if
// no records are found.
func MustMatchRecord(d *DataFile, byName bool, args []string) *CommandRecord {
	records, err := MatchRecords(d, byName, args)
	essentials.Must(err)
	if len(records) == 0 {
		fmt.Fprintln(os.Stderr, "no records match the query")
		os.Exit(1)
	}
	return records[len(records)-1]
}

// MatchRecords searches the records in a DataFile and
// returns the results in order of least to most relevant.
func MatchRecords(d *DataFile, byName bool, args []string) ([]*CommandRecord, error) {
	queryRegexpStr := ""
	for i, arg := range args {
		if i > 0 {
			queryRegexpStr += "[ \t\n]+?"
		}
		queryRegexpStr += regexp.QuoteMeta(arg)
	}
	fuzzyMatch := regexp.MustCompilePOSIX(queryRegexpStr)
	prefixMatch := regexp.MustCompilePOSIX("^" + queryRegexpStr)
	exactMatch := regexp.MustCompilePOSIX("^" + queryRegexpStr + "$")

	d.Reset()
	var fuzzyMatches []*CommandRecord
	var prefixMatches []*CommandRecord
	var exactMatches []*CommandRecord
	for {
		record, err := d.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		var matchStr string
		if byName {
			matchStr = record.ID
		} else {
			matchStr = record.Command
		}
		if exactMatch.MatchString(matchStr) {
			exactMatches = append(exactMatches, record)
		} else if prefixMatch.MatchString(matchStr) {
			prefixMatches = append(prefixMatches, record)
		} else if fuzzyMatch.MatchString(matchStr) {
			fuzzyMatches = append(fuzzyMatches, record)
		}
	}

	matches := fuzzyMatches
	matches = append(matches, prefixMatches...)
	matches = append(matches, exactMatches...)
	return matches, nil
}
