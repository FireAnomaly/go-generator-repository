package main

import (
	"os"
	"testing"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func TestOutputCLITables(t *testing.T) {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	if err := table.Start(); err != nil {
		t.Fatalf("Failed to start table: %v", err)
	}

	defer table.Close()

	table.Header("Name", "Value")
	_ = table.Append([]string{"a", "b"})
	time.Sleep(3 * time.Second)

	_ = table.Append([]string{"aaaaa", "bbbbb"})

}
