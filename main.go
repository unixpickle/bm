package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/unixpickle/essentials"
)

const (
	MaxMatches = 10

	// Escape codes to color code the output of a list.
	ListIDColor    = "\033[32m\033[1m"
	ListResetColor = "\033[0m"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		DieUsage()
	}

	commandName := os.Args[1]
	if (len(commandName) != 1 && len(commandName) != 2) ||
		(len(commandName) == 2 && commandName[1] != 'n') {
		DieUnknownCommand()
	}

	byName := len(commandName) == 2
	cmds := map[byte]func(*DataFile, bool, []string){
		's': CommandSave,
		'u': CommandUpdate,
		'c': CommandSaveAndRun,
		'x': CommandSaveAndRunOverwrite,
		'a': CommandAll,
		'q': CommandQuery,
		'r': CommandQueryAndRun,
		'd': CommandDelete,
	}
	if fn, ok := cmds[commandName[0]]; ok {
		d, err := NewDataFile()
		essentials.Must(err)
		defer d.Close()
		fn(d, byName, os.Args[2:])
	} else {
		DieUnknownCommand()
	}
}

func CommandSave(d *DataFile, byName bool, args []string) {
	commandSave(d, byName, args)
}

func CommandUpdate(d *DataFile, byName bool, args []string) {
	if !byName {
		DieUnknownCommand()
	}
	ok, err := d.CanUseID(args[0])
	essentials.Must(err)
	if !ok {
		essentials.Must(d.Delete(args[0]))
	}
	commandSave(d, byName, args)
}

func commandSave(d *DataFile, byName bool, args []string) *CommandRecord {
	var id string
	if byName {
		id = args[0]
		args = args[1:]
		ok, err := d.CanUseID(id)
		essentials.Must(err)
		if !ok {
			fmt.Fprintln(os.Stderr, "cannot use name:", id)
			os.Exit(1)
		}
	} else {
		var err error
		id, err = d.GenerateUniqueID()
		essentials.Must(err)
	}

	record := &CommandRecord{
		ID:      id,
		Command: strings.Join(args, " "),
		Date:    time.Now(),
	}
	essentials.Must(d.Write(record))

	fmt.Fprintln(os.Stderr, "created record with ID", record.ID)

	return record
}

func CommandSaveAndRun(d *DataFile, byName bool, args []string) {
	record := commandSave(d, byName, args)
	d.Close()
	essentials.Must(Run(record))
}

func CommandSaveAndRunOverwrite(d *DataFile, byName bool, args []string) {
	if !byName {
		DieUnknownCommand()
	}
	ok, err := d.CanUseID(args[0])
	essentials.Must(err)
	if !ok {
		essentials.Must(d.Delete(args[0]))
	}
	record := commandSave(d, byName, args)
	d.Close()
	essentials.Must(Run(record))
}

func CommandAll(d *DataFile, byName bool, args []string) {
	records, err := MatchRecords(d, byName, args)
	essentials.Must(err)

	if len(records) == 0 {
		fmt.Fprintln(os.Stderr, "no records found")
		os.Exit(1)
	} else {
		printRecords(records)
	}
}

func CommandQuery(d *DataFile, byName bool, args []string) {
	records, err := MatchRecords(d, byName, args)
	essentials.Must(err)

	if len(records) == 0 {
		fmt.Fprintln(os.Stderr, "no records match the query")
		os.Exit(1)
	}

	if len(records) > MaxMatches {
		records = records[len(records)-MaxMatches:]
	}
	printRecords(records)
}

func printRecords(records []*CommandRecord) {
	maxIDLen := 0
	for _, r := range records {
		maxIDLen = essentials.MaxInt(maxIDLen, len(r.ID))
	}
	info, err := os.Stdout.Stat()
	essentials.Must(err)
	isTTY := (info.Mode() & os.ModeCharDevice) != 0
	for i, r := range records {
		if isTTY && i == 0 {
			// Space things out to make them easier to read.
			fmt.Println()
		}
		paddedID := r.ID
		for len(paddedID) < maxIDLen {
			paddedID = " " + paddedID
		}
		if isTTY {
			paddedID = ListIDColor + paddedID + ListResetColor
		}
		fmt.Println(" " + paddedID + "  " + r.Command)
		if isTTY {
			// Space things out to make them easier to read.
			fmt.Println()
		}
	}
}

func CommandQueryAndRun(d *DataFile, byName bool, args []string) {
	record := MustMatchRecord(d, byName, args)
	d.Close()
	essentials.Must(Run(record))
}

func CommandDelete(d *DataFile, byName bool, args []string) {
	record := MustMatchRecord(d, byName, args)
	fmt.Fprintln(os.Stderr, "deleting record '"+record.ID+"' with command: "+record.Command)
	essentials.Must(d.Delete(record.ID))
}

func DieUnknownCommand() {
	fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
	fmt.Fprintln(os.Stderr)
	DieUsage()
}

func DieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<sub-command> [args...]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Sub-command usage:")
	fmt.Fprintln(os.Stderr, "    s  [args]           save an un-named command")
	fmt.Fprintln(os.Stderr, "    sn <name> [args]    save a named command")
	fmt.Fprintln(os.Stderr, "    un <name> [args]    like sn, but may replace an existing command")
	fmt.Fprintln(os.Stderr, "    c  [args]           save and run an un-named command")
	fmt.Fprintln(os.Stderr, "    cn <name> [args]    save and run a named command")
	fmt.Fprintln(os.Stderr, "    xn <name> [args]    like cn, but may replace an existing command")
	fmt.Fprintln(os.Stderr, "    q  [query]          search saved commands by content")
	fmt.Fprintln(os.Stderr, "    qn [query]          search saved commands by name")
	fmt.Fprintln(os.Stderr, "    r  [query]          lookup and run a command by content")
	fmt.Fprintln(os.Stderr, "    rn [query]          lookup and run a command by name")
	fmt.Fprintln(os.Stderr, "    d  [query]          delete a command by content")
	fmt.Fprintln(os.Stderr, "    dn [query]          delete a command by name")
	fmt.Fprintln(os.Stderr, "    a  [query]          list all saved commands in order")
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
