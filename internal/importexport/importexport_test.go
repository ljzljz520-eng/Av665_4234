package importexport

import (
	"strings"
	"testing"
)

func TestImportParsingAndReport(t *testing.T) {
	input := "code,type,building,floor,inspection_date,owner,status,notes\nV-122,灭火器,A栋,2,2026-08-23,张三,draft,入口\nV-123,消火栓,A栋,1,2026-08-24,李四,draft,\n"
	rows, err := ParseCSV(strings.NewReader(input))
	if err != nil || len(rows) != 2 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	outcome := ValidateRows(rows, map[string]bool{"V-123": true})
	if len(outcome.Accepted) != 1 || len(outcome.Issues) != 1 {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	report, err := ExportReport(outcome)
	if err != nil || !strings.Contains(report, "accepted") || !strings.Contains(report, "rejected") {
		t.Fatalf("unexpected report: %q %v", report, err)
	}
}

func TestImportHeaderAndValidation(t *testing.T) {
	if _, err := ParseCSV(strings.NewReader("bad,header\n")); err == nil {
		t.Fatal("bad header should fail")
	}
	if len(Header()) != 8 {
		t.Fatal("unexpected import header")
	}
}
