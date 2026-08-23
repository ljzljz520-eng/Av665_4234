package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"fire-equipment-control/internal/domain"
)

type StatusTile struct {
	Status domain.Status
	Label  string
	Count  int
}

type LocationTile struct {
	Building string
	Floor    int
	Count    int
	Codes    []string
}

type AuditLine struct {
	Sequence int
	Code     string
	Actor    string
	Action   string
	Detail   string
}

type View struct {
	Title           string
	Total           int
	StatusTiles     []StatusTile
	LocationTiles   []LocationTile
	AuditLines      []AuditLine
	ActionCounts    map[string]int
	OwnerCounts     map[string]int
	ActiveBuildings []string
}

func Build(records []domain.EquipmentRecord, events []domain.AuditEvent) View {
	view := View{Title: "消防器材管理台账", Total: len(records), ActionCounts: map[string]int{}, OwnerCounts: map[string]int{}}
	view.StatusTiles = buildStatusTiles(records)
	view.LocationTiles = buildLocationTiles(records)
	view.AuditLines = buildAuditLines(events)
	view.ActiveBuildings = activeBuildings(records)
	for _, event := range events {
		view.ActionCounts[event.Action]++
	}
	for _, record := range records {
		view.OwnerCounts[record.Owner]++
	}
	return view
}

func buildStatusTiles(records []domain.EquipmentRecord) []StatusTile {
	counts := map[domain.Status]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	order := []domain.Status{domain.StatusActive, domain.StatusReview, domain.StatusApproved, domain.StatusDraft, domain.StatusDisabled, domain.StatusRetired, domain.StatusArchived}
	tiles := make([]StatusTile, 0, len(order))
	for _, status := range order {
		tiles = append(tiles, StatusTile{Status: status, Label: domain.StatusLabel(status), Count: counts[status]})
	}
	return tiles
}

func buildLocationTiles(records []domain.EquipmentRecord) []LocationTile {
	grouped := map[string]LocationTile{}
	for _, record := range records {
		key := fmt.Sprintf("%s/%d", record.Building, record.Floor)
		tile := grouped[key]
		tile.Building = record.Building
		tile.Floor = record.Floor
		tile.Count++
		tile.Codes = append(tile.Codes, record.Code)
		grouped[key] = tile
	}
	tiles := make([]LocationTile, 0, len(grouped))
	for _, tile := range grouped {
		sort.Strings(tile.Codes)
		tiles = append(tiles, tile)
	}
	sort.Slice(tiles, func(i, j int) bool {
		if tiles[i].Building != tiles[j].Building {
			return tiles[i].Building < tiles[j].Building
		}
		return tiles[i].Floor < tiles[j].Floor
	})
	return tiles
}

func buildAuditLines(events []domain.AuditEvent) []AuditLine {
	ordered := append([]domain.AuditEvent(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Sequence != ordered[j].Sequence {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		return ordered[i].ID < ordered[j].ID
	})
	lines := make([]AuditLine, 0, len(ordered))
	for _, event := range ordered {
		lines = append(lines, AuditLine{Sequence: event.Sequence, Code: event.EquipmentCode, Actor: event.Actor, Action: event.Action, Detail: event.Detail})
	}
	return lines
}

func activeBuildings(records []domain.EquipmentRecord) []string {
	seen := map[string]bool{}
	for _, record := range records {
		if record.Status == domain.StatusActive || record.Status == domain.StatusReview || record.Status == domain.StatusApproved {
			seen[record.Building] = true
		}
	}
	buildings := make([]string, 0, len(seen))
	for building := range seen {
		buildings = append(buildings, building)
	}
	sort.Strings(buildings)
	return buildings
}

func Render(view View) string {
	sections := []string{view.Title, fmt.Sprintf("总数: %d", view.Total), RenderStatuses(view.StatusTiles), RenderLocations(view.LocationTiles), RenderAudits(view.AuditLines)}
	return strings.Join(sections, "\n")
}

func RenderStatuses(tiles []StatusTile) string {
	parts := make([]string, 0, len(tiles))
	for _, tile := range tiles {
		if tile.Count > 0 {
			parts = append(parts, tile.Label+"="+integerText(tile.Count))
		}
	}
	if len(parts) == 0 {
		return "状态: 无记录"
	}
	return "状态: " + strings.Join(parts, " | ")
}

func RenderLocations(tiles []LocationTile) string {
	parts := make([]string, 0, len(tiles))
	for _, tile := range tiles {
		parts = append(parts, fmt.Sprintf("%s F%d[%s]", tile.Building, tile.Floor, strings.Join(tile.Codes, ",")))
	}
	if len(parts) == 0 {
		return "位置: 无记录"
	}
	return "位置: " + strings.Join(parts, " | ")
}

func RenderAudits(lines []AuditLine) string {
	if len(lines) == 0 {
		return "审计: 无记录"
	}
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, fmt.Sprintf("%04d %s %s %s", line.Sequence, line.Code, line.Action, line.Actor))
	}
	return "审计: " + strings.Join(parts, " | ")
}

func FilterStatus(view View, status domain.Status) []StatusTile {
	filtered := []StatusTile{}
	for _, tile := range view.StatusTiles {
		if status == "" || tile.Status == status {
			filtered = append(filtered, tile)
		}
	}
	return filtered
}

func FindLocation(view View, building string, floor int) (LocationTile, bool) {
	for _, tile := range view.LocationTiles {
		if tile.Building == building && tile.Floor == floor {
			return tile, true
		}
	}
	return LocationTile{}, false
}

func OwnerLeaderboard(view View) []string {
	type entry struct {
		owner string
		count int
	}
	entries := make([]entry, 0, len(view.OwnerCounts))
	for owner, count := range view.OwnerCounts {
		entries = append(entries, entry{owner: owner, count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].owner < entries[j].owner
	})
	owners := make([]string, 0, len(entries))
	for _, item := range entries {
		owners = append(owners, fmt.Sprintf("%s=%d", item.owner, item.count))
	}
	return owners
}

func integerText(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 12)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

func EmptyView() View {
	return Build(nil, nil)
}

func HasAuditAction(view View, action string) bool {
	return view.ActionCounts[action] > 0
}

func EquipmentCodes(view View) []string {
	codes := []string{}
	for _, tile := range view.LocationTiles {
		codes = append(codes, tile.Codes...)
	}
	sort.Strings(codes)
	return codes
}

func StatusCount(view View, status domain.Status) int {
	for _, tile := range view.StatusTiles {
		if tile.Status == status {
			return tile.Count
		}
	}
	return 0
}

func AuditCount(view View) int {
	return len(view.AuditLines)
}

func LocationCount(view View) int {
	return len(view.LocationTiles)
}

func IsEmpty(view View) bool {
	return view.Total == 0 && len(view.AuditLines) == 0
}

func StatusLegend() []string {
	return []string{
		domain.StatusLabel(domain.StatusActive),
		domain.StatusLabel(domain.StatusReview),
		domain.StatusLabel(domain.StatusApproved),
		domain.StatusLabel(domain.StatusDraft),
		domain.StatusLabel(domain.StatusDisabled),
		domain.StatusLabel(domain.StatusRetired),
		domain.StatusLabel(domain.StatusArchived),
	}
}

func BuildingNames(view View) []string {
	return append([]string(nil), view.ActiveBuildings...)
}

func ActionCount(view View, action string) int {
	return view.ActionCounts[action]
}

func SummaryRows(view View) [][]string {
	rows := [][]string{{"指标", "数值"}, {"总数", integerText(view.Total)}, {"位置数", integerText(LocationCount(view))}, {"审计事件", integerText(AuditCount(view))}}
	for _, tile := range view.StatusTiles {
		if tile.Count > 0 {
			rows = append(rows, []string{tile.Label, integerText(tile.Count)})
		}
	}
	return rows
}

func ValidateView(view View) error {
	if view.Total < 0 {
		return fmt.Errorf("dashboard total cannot be negative")
	}
	if len(view.StatusTiles) != 7 {
		return fmt.Errorf("dashboard must expose seven status tiles")
	}
	for _, tile := range view.StatusTiles {
		if tile.Count < 0 {
			return fmt.Errorf("status count cannot be negative")
		}
	}
	return nil
}
