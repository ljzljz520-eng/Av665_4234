package inspection

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fire-equipment-control/internal/domain"
)

type Grade string

const (
	GradeCurrent Grade = "current"
	GradeDue     Grade = "due"
	GradeLate    Grade = "late"
	GradeInvalid Grade = "invalid"
)

type Assessment struct {
	Code           string
	InspectionDate string
	AsOf           string
	Grade          Grade
	Message        string
	FollowUp       bool
}

type PlanItem struct {
	Code     string
	Building string
	Floor    int
	Owner    string
	Grade    Grade
	Priority int
}

var ErrDateFormat = errors.New("date must use YYYY-MM-DD")

func ParseDate(value string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 3 || len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return 0, 0, 0, ErrDateFormat
	}
	year, yearErr := strconv.Atoi(parts[0])
	month, monthErr := strconv.Atoi(parts[1])
	day, dayErr := strconv.Atoi(parts[2])
	if yearErr != nil || monthErr != nil || dayErr != nil || !ValidMonthDay(month, day) {
		return 0, 0, 0, ErrDateFormat
	}
	return year, month, day, nil
}

func ValidMonthDay(month, day int) bool {
	if month < 1 || month > 12 {
		return false
	}
	return day >= 1 && day <= MonthLength(month)
}

func MonthLength(month int) int {
	switch month {
	case 2:
		return 28
	case 4, 6, 9, 11:
		return 30
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	default:
		return 0
	}
}

func NormalizeDate(value string) string {
	year, month, day, err := ParseDate(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func DateOrdinal(value string) (int, error) {
	year, month, day, err := ParseDate(value)
	if err != nil {
		return 0, err
	}
	ordinal := year*365 + day
	for current := 1; current < month; current++ {
		ordinal += MonthLength(current)
	}
	return ordinal, nil
}

func DaysBetween(start, end string) (int, error) {
	startOrdinal, err := DateOrdinal(start)
	if err != nil {
		return 0, err
	}
	endOrdinal, err := DateOrdinal(end)
	if err != nil {
		return 0, err
	}
	return endOrdinal - startOrdinal, nil
}

func IsDue(inspectionDate, asOf string) (bool, error) {
	days, err := DaysBetween(inspectionDate, asOf)
	if err != nil {
		return false, err
	}
	return days >= 365, nil
}

func Assess(record domain.EquipmentRecord, asOf string) Assessment {
	assessment := Assessment{Code: record.Code, InspectionDate: record.InspectionDate, AsOf: asOf, Grade: GradeInvalid, FollowUp: true}
	if _, _, _, err := ParseDate(record.InspectionDate); err != nil {
		assessment.Message = "检查日期格式无效"
		return assessment
	}
	days, err := DaysBetween(record.InspectionDate, asOf)
	if err != nil {
		assessment.Message = "评估日期格式无效"
		return assessment
	}
	if days < 0 {
		assessment.Grade = GradeCurrent
		assessment.Message = "检查日期在评估日之后"
		assessment.FollowUp = false
		return assessment
	}
	if days >= 730 {
		assessment.Grade = GradeLate
		assessment.Message = "检查已逾期两年"
		return assessment
	}
	if days >= 365 {
		assessment.Grade = GradeDue
		assessment.Message = "需要安排复检"
		return assessment
	}
	assessment.Grade = GradeCurrent
	assessment.Message = "检查仍在有效期"
	assessment.FollowUp = false
	return assessment
}

func BuildPlan(records []domain.EquipmentRecord, asOf string) []PlanItem {
	plan := make([]PlanItem, 0, len(records))
	for _, record := range records {
		assessment := Assess(record, asOf)
		priority := Priority(assessment.Grade)
		if priority > 0 {
			plan = append(plan, PlanItem{Code: record.Code, Building: record.Building, Floor: record.Floor, Owner: record.Owner, Grade: assessment.Grade, Priority: priority})
		}
	}
	SortPlan(plan)
	return plan
}

func Priority(grade Grade) int {
	switch grade {
	case GradeLate:
		return 1
	case GradeDue:
		return 2
	case GradeInvalid:
		return 3
	default:
		return 0
	}
}

func SortPlan(plan []PlanItem) {
	sort.SliceStable(plan, func(i, j int) bool {
		if plan[i].Priority != plan[j].Priority {
			return plan[i].Priority < plan[j].Priority
		}
		if plan[i].Building != plan[j].Building {
			return plan[i].Building < plan[j].Building
		}
		if plan[i].Floor != plan[j].Floor {
			return plan[i].Floor < plan[j].Floor
		}
		return plan[i].Code < plan[j].Code
	})
}

func GroupByBuilding(plan []PlanItem) map[string][]PlanItem {
	grouped := map[string][]PlanItem{}
	for _, item := range plan {
		grouped[item.Building] = append(grouped[item.Building], item)
	}
	return grouped
}

func PlanSummary(plan []PlanItem) string {
	counts := map[Grade]int{}
	for _, item := range plan {
		counts[item.Grade]++
	}
	return fmt.Sprintf("late=%d,due=%d,invalid=%d,total=%d", counts[GradeLate], counts[GradeDue], counts[GradeInvalid], len(plan))
}

func ShouldBlockArchive(record domain.EquipmentRecord, asOf string) bool {
	assessment := Assess(record, asOf)
	return assessment.Grade == GradeInvalid || assessment.Grade == GradeLate
}
