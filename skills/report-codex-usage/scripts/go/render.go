package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type totalsRow struct {
	Name   string
	Totals totals
}

func sortedTotals(values map[string]totals) []totalsRow {
	rows := make([]totalsRow, 0, len(values))
	for name, value := range values {
		rows = append(rows, totalsRow{Name: name, Totals: value})
	}
	sort.Slice(rows, func(left, right int) bool {
		if rows[left].Totals.Total != rows[right].Totals.Total {
			return rows[left].Totals.Total > rows[right].Totals.Total
		}
		return rows[left].Name < rows[right].Name
	})
	return rows
}

func sumTotals(values []totals) totals {
	result := totals{}
	for _, value := range values {
		result.merge(value)
	}
	return result
}

func assertSameTotals(name string, candidate, expected totals) error {
	switch {
	case candidate.Total != expected.Total:
		return fmt.Errorf("%s total token mismatch", name)
	case candidate.CachedInput != expected.CachedInput:
		return fmt.Errorf("%s cached input mismatch", name)
	case candidate.Input != expected.Input:
		return fmt.Errorf("%s input mismatch", name)
	case candidate.Output != expected.Output:
		return fmt.Errorf("%s output mismatch", name)
	case candidate.Events != expected.Events:
		return fmt.Errorf("%s event count mismatch", name)
	case math.Abs(candidate.CalculatedCost-expected.CalculatedCost) > 1e-9:
		return fmt.Errorf("%s calculated cost mismatch", name)
	default:
		return nil
	}
}

func assertReportIntegrity(report report) error {
	projectValues := make([]totals, 0, len(report.Projects))
	for _, value := range report.Projects {
		projectValues = append(projectValues, value)
	}
	modelValues := make([]totals, 0, len(report.Models))
	for _, value := range report.Models {
		modelValues = append(modelValues, value)
	}
	sessionValues := make([]totals, 0)
	for _, sessions := range report.Sessions {
		for _, session := range sessions {
			sessionValues = append(sessionValues, session.Totals)
		}
	}
	if err := assertSameTotals("project", sumTotals(projectValues), report.Total); err != nil {
		return err
	}
	if err := assertSameTotals("model", sumTotals(modelValues), report.Total); err != nil {
		return err
	}
	if err := assertSameTotals("session", sumTotals(sessionValues), report.Total); err != nil {
		return err
	}
	for project, projectTotal := range report.Projects {
		values := make([]totals, 0, len(report.Sessions[project]))
		for _, session := range report.Sessions[project] {
			values = append(values, session.Totals)
		}
		if err := assertSameTotals(project+" session", sumTotals(values), projectTotal); err != nil {
			return err
		}
	}
	return nil
}

func commaInteger(value uint64) string {
	raw := strconv.FormatUint(value, 10)
	var result strings.Builder
	result.Grow(len(raw) + len(raw)/3)
	for index, character := range raw {
		if index > 0 && (len(raw)-index)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(character)
	}
	return result.String()
}

func commaDecimal(value float64) string {
	raw := fmt.Sprintf("%.2f", value)
	parts := strings.SplitN(raw, ".", 2)
	integer, _ := strconv.ParseUint(parts[0], 10, 64)
	return commaInteger(integer) + "." + parts[1]
}

func formatTokens(value uint64) string {
	return commaDecimal(float64(value)/1_000_000) + "M"
}

func percent(value, total uint64) string {
	if total == 0 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", float64(value)/float64(total)*100)
}

func tokenCell(value, total uint64) string {
	return fmt.Sprintf("%s (%s)", formatTokens(value), percent(value, total))
}

func escapeMarkdown(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}

func sortedSessions(sessions map[string]*sessionUsage) []*sessionUsage {
	rows := make([]*sessionUsage, 0, len(sessions))
	for _, session := range sessions {
		rows = append(rows, session)
	}
	sort.Slice(rows, func(left, right int) bool {
		if rows[left].Totals.Total != rows[right].Totals.Total {
			return rows[left].Totals.Total > rows[right].Totals.Total
		}
		if rows[left].Title != rows[right].Title {
			return rows[left].Title < rows[right].Title
		}
		return rows[left].SessionID < rows[right].SessionID
	})
	return rows
}

func markdownProjectSessions(report report) []string {
	lines := []string{
		"| 프로젝트 / 세션 | 모델 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |",
		"|---|---|---:|---:|---:|---:|---:|",
	}
	for _, row := range sortedTotals(report.Projects) {
		sessions := report.Sessions[row.Name]
		item := row.Totals
		lines = append(lines, fmt.Sprintf(
			"| **%s (%d개)** |  | **%s** | **%s** | **%s** | **%s** | **$%.2f** |",
			escapeMarkdown(row.Name),
			len(sessions),
			formatTokens(item.Total),
			tokenCell(item.CachedInput, item.Total),
			tokenCell(item.Input, item.Total),
			tokenCell(item.Output, item.Total),
			item.CalculatedCost,
		))
		for _, session := range sortedSessions(sessions) {
			item := session.Totals
			lines = append(lines, fmt.Sprintf(
				"| └ %s | %s | %s | %s | %s | %s | $%.2f |",
				escapeMarkdown(session.Title),
				displayModels(session.Models),
				formatTokens(item.Total),
				tokenCell(item.CachedInput, item.Total),
				tokenCell(item.Input, item.Total),
				tokenCell(item.Output, item.Total),
				item.CalculatedCost,
			))
		}
	}
	sessionCount := 0
	for _, sessions := range report.Sessions {
		sessionCount += len(sessions)
	}
	lines = append(lines, fmt.Sprintf(
		"| **전체 (%d개 세션)** |  | **%s** | **%s** | **%s** | **%s** | **$%.2f** |",
		sessionCount,
		formatTokens(report.Total.Total),
		tokenCell(report.Total.CachedInput, report.Total.Total),
		tokenCell(report.Total.Input, report.Total.Total),
		tokenCell(report.Total.Output, report.Total.Total),
		report.Total.CalculatedCost,
	))
	return lines
}

func markdownTable(values map[string]totals, total totals) []string {
	lines := []string{
		"| 이름 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |",
		"|---|---:|---:|---:|---:|---:|",
	}
	for _, row := range sortedTotals(values) {
		item := row.Totals
		lines = append(lines, fmt.Sprintf(
			"| %s | %s | %s | %s | %s | $%.2f |",
			row.Name,
			formatTokens(item.Total),
			tokenCell(item.CachedInput, item.Total),
			tokenCell(item.Input, item.Total),
			tokenCell(item.Output, item.Total),
			item.CalculatedCost,
		))
	}
	lines = append(lines, fmt.Sprintf(
		"| 합계 | %s | %s | %s | %s | $%.2f |",
		formatTokens(total.Total),
		tokenCell(total.CachedInput, total.Total),
		tokenCell(total.Input, total.Total),
		tokenCell(total.Output, total.Total),
		total.CalculatedCost,
	))
	return lines
}

func renderMarkdown(report report) string {
	sessionCount := 0
	for _, sessions := range report.Sessions {
		sessionCount += len(sessions)
	}
	diagnostics := report.Diagnostics
	lines := []string{
		fmt.Sprintf("## %s Codex 일일 토큰 보고", report.TargetDate.Format("2006-01-02")),
		"",
		fmt.Sprintf("- 집계 시각: %s %s", report.GeneratedAt.Format("2006-01-02 15:04:05"), report.TimezoneName),
		fmt.Sprintf(
			"- 집계 기간: %s 이상, %s 미만 %s",
			report.RangeStart.Format("2006-01-02 15:04:05"),
			report.RangeEnd.Format("2006-01-02 15:04:05"),
			report.TimezoneName,
		),
		fmt.Sprintf("- 집계 장치: %s", report.ComputerName),
		fmt.Sprintf("- 원본 token_count 이벤트: %s개", commaInteger(diagnostics.OriginalEvents)),
		fmt.Sprintf("- 중복 제거 이벤트: %s개", commaInteger(diagnostics.DuplicateEvents)),
		fmt.Sprintf("- 상속 history 제외 이벤트: %s개", commaInteger(diagnostics.ReplayedEvents)),
		fmt.Sprintf("- 집계 token_count 이벤트: %s개", commaInteger(diagnostics.AggregatedEvents)),
		fmt.Sprintf("- 프로젝트: %s개", commaInteger(uint64(len(report.Projects)))),
		fmt.Sprintf("- 세션: %s개", commaInteger(uint64(sessionCount))),
		fmt.Sprintf("- 모델: %s개", commaInteger(uint64(len(report.Models)))),
		"",
		"### 프로젝트별",
		"",
	}
	lines = append(lines, markdownProjectSessions(report)...)
	lines = append(lines, "", "### 모델별", "")
	lines = append(lines, markdownTable(report.Models, report.Total)...)
	return strings.Join(lines, "\n")
}

type namedTotalsJSON struct {
	Name string `json:"name"`
	totals
}

type sessionJSON struct {
	SessionID string   `json:"session_id"`
	Title     string   `json:"title"`
	Models    []string `json:"models"`
	totals
}

type projectSessionsJSON struct {
	Project  string        `json:"project"`
	Total    totals        `json:"total"`
	Sessions []sessionJSON `json:"sessions"`
}

type rangeJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type reportJSON struct {
	Date         string                `json:"date"`
	Timezone     string                `json:"timezone"`
	Range        rangeJSON             `json:"range"`
	GeneratedAt  string                `json:"generated_at"`
	ComputerName string                `json:"computer_name"`
	Diagnostics  diagnostics           `json:"diagnostics"`
	Projects     []namedTotalsJSON     `json:"projects"`
	Models       []namedTotalsJSON     `json:"models"`
	Sessions     []projectSessionsJSON `json:"sessions"`
	Total        totals                `json:"total"`
}

func reportToJSON(report report) (string, error) {
	projects := make([]namedTotalsJSON, 0, len(report.Projects))
	for _, row := range sortedTotals(report.Projects) {
		projects = append(projects, namedTotalsJSON{Name: row.Name, totals: row.Totals})
	}
	models := make([]namedTotalsJSON, 0, len(report.Models))
	for _, row := range sortedTotals(report.Models) {
		models = append(models, namedTotalsJSON{Name: row.Name, totals: row.Totals})
	}
	projectSessions := make([]projectSessionsJSON, 0, len(report.Projects))
	for _, row := range sortedTotals(report.Projects) {
		sessions := make([]sessionJSON, 0, len(report.Sessions[row.Name]))
		for _, session := range sortedSessions(report.Sessions[row.Name]) {
			models := make([]string, 0, len(session.Models))
			for _, label := range modelLabelOrder {
				if _, exists := session.Models[label]; exists {
					models = append(models, label)
				}
			}
			sessions = append(sessions, sessionJSON{
				SessionID: session.SessionID,
				Title:     session.Title,
				Models:    models,
				totals:    session.Totals,
			})
		}
		projectSessions = append(projectSessions, projectSessionsJSON{
			Project:  row.Name,
			Total:    row.Totals,
			Sessions: sessions,
		})
	}

	payload := reportJSON{
		Date:     report.TargetDate.Format("2006-01-02"),
		Timezone: report.TimezoneName,
		Range: rangeJSON{
			Start: report.RangeStart.Format(timeRFC3339()),
			End:   report.RangeEnd.Format(timeRFC3339()),
		},
		GeneratedAt:  report.GeneratedAt.Format(timeRFC3339()),
		ComputerName: report.ComputerName,
		Diagnostics:  report.Diagnostics,
		Projects:     projects,
		Models:       models,
		Sessions:     projectSessions,
		Total:        report.Total,
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 보고서 직렬화 실패: %w", err)
	}
	return string(encoded), nil
}

func timeRFC3339() string {
	return "2006-01-02T15:04:05.999999999Z07:00"
}
