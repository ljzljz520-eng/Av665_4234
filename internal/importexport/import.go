package importexport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"fire-equipment-control/internal/domain"
)

var expectedHeader = []string{"code", "type", "building", "floor", "inspection_date", "owner", "status", "notes"}

func ParseCSV(reader io.Reader) ([]domain.ImportRow, error) {
	parsed := csv.NewReader(reader)
	parsed.TrimLeadingSpace = true
	header, err := parsed.Read()
	if err != nil {
		return nil, fmt.Errorf("read import header: %w", err)
	}
	if !sameHeader(header) {
		return nil, errors.New("import header does not match the required columns")
	}
	rows := []domain.ImportRow{}
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := parsed.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read import row %d: %w", rowNumber, readErr)
		}
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has %d columns, want %d", rowNumber, len(row), len(expectedHeader))
		}
		floor, parseErr := strconv.Atoi(strings.TrimSpace(row[3]))
		if parseErr != nil {
			floor = 0
		}
		rows = append(rows, domain.ImportRow{
			Code: strings.TrimSpace(row[0]), Type: strings.TrimSpace(row[1]), Building: strings.TrimSpace(row[2]), Floor: floor,
			InspectionDate: strings.TrimSpace(row[4]), Owner: strings.TrimSpace(row[5]), Status: domain.Status(strings.TrimSpace(row[6])), Notes: strings.TrimSpace(row[7]),
		})
	}
	return rows, nil
}

func sameHeader(header []string) bool {
	if len(header) != len(expectedHeader) {
		return false
	}
	for index, value := range header {
		if strings.ToLower(strings.TrimSpace(value)) != expectedHeader[index] {
			return false
		}
	}
	return true
}

func ValidateRows(rows []domain.ImportRow, existing map[string]bool) domain.ImportOutcome {
	outcome := domain.ImportOutcome{Accepted: []domain.EquipmentRecord{}, Issues: []domain.ImportIssue{}}
	seen := map[string]bool{}
	for rowIndex, row := range rows {
		record := domain.EquipmentRecord{Code: row.Code, Type: row.Type, Building: row.Building, Floor: row.Floor, InspectionDate: row.InspectionDate, Owner: row.Owner, Status: row.Status, Notes: row.Notes}.Normalized()
		if record.Status == "" {
			record.Status = domain.StatusDraft
		}
		issues := validateRow(record, rowIndex+2, existing, seen)
		if len(issues) > 0 {
			outcome.Issues = append(outcome.Issues, issues...)
			continue
		}
		record.ID = "equipment-" + record.Code
		seen[record.Code] = true
		outcome.Accepted = append(outcome.Accepted, record)
	}
	return outcome
}

func validateRow(record domain.EquipmentRecord, row int, existing, seen map[string]bool) []domain.ImportIssue {
	issues := []domain.ImportIssue{}
	if err := record.Validate(); err != nil {
		issues = append(issues, domain.ImportIssue{Row: row, Field: fieldForError(err), Message: err.Error()})
	}
	if !domain.ValidateDateShape(record.InspectionDate) {
		issues = append(issues, domain.ImportIssue{Row: row, Field: "inspection_date", Message: "inspection date must be YYYY-MM-DD"})
	}
	if existing[record.Code] {
		issues = append(issues, domain.ImportIssue{Row: row, Field: "code", Message: "code already exists"})
	}
	if seen[record.Code] {
		issues = append(issues, domain.ImportIssue{Row: row, Field: "code", Message: "duplicate code in import"})
	}
	return issues
}

func fieldForError(err error) string {
	switch {
	case errors.Is(err, domain.ErrCodeRequired):
		return "code"
	case errors.Is(err, domain.ErrTypeRequired):
		return "type"
	case errors.Is(err, domain.ErrBuildingRequired):
		return "building"
	case errors.Is(err, domain.ErrInvalidFloor):
		return "floor"
	case errors.Is(err, domain.ErrInspectionRequired):
		return "inspection_date"
	case errors.Is(err, domain.ErrOwnerRequired):
		return "owner"
	default:
		return "status"
	}
}

func ExportReport(outcome domain.ImportOutcome) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"row", "code", "result", "field", "message"}); err != nil {
		return "", err
	}
	for index, record := range outcome.Accepted {
		if err := writer.Write([]string{strconv.Itoa(index + 2), record.Code, "accepted", "", ""}); err != nil {
			return "", err
		}
	}
	for _, issue := range outcome.Issues {
		if err := writer.Write([]string{strconv.Itoa(issue.Row), "", "rejected", issue.Field, issue.Message}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func Header() []string {
	return append([]string(nil), expectedHeader...)
}

func AcceptedCodes(outcome domain.ImportOutcome) []string {
	codes := make([]string, 0, len(outcome.Accepted))
	for _, record := range outcome.Accepted {
		codes = append(codes, record.Code)
	}
	return codes
}
