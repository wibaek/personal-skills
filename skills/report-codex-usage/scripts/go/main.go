package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type options struct {
	Date         string
	Timezone     string
	SessionsRoot string
	RateCard     string
	SessionIndex string
	GlobalState  string
	ComputerName string
	Format       string
}

func parseOptions(arguments []string) (options, error) {
	skillRoot := defaultSkillRoot()
	home, _ := os.UserHomeDir()
	defaults := options{
		Date:         "yesterday",
		Timezone:     "Asia/Seoul",
		RateCard:     filepath.Join(skillRoot, "references", "rate-card.toml"),
		SessionIndex: filepath.Join(home, ".codex", "session_index.jsonl"),
		GlobalState:  filepath.Join(home, ".codex", ".codex-global-state.json"),
		Format:       "markdown",
	}
	parsed := defaults
	flags := flag.NewFlagSet("report-codex-usage", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&parsed.Date, "date", defaults.Date, "today, yesterday, or YYYY-MM-DD (default: yesterday)")
	flags.StringVar(&parsed.Timezone, "timezone", defaults.Timezone, "IANA timezone used for the calendar-day boundary")
	flags.StringVar(&parsed.SessionsRoot, "sessions-root", "", "override the default active and archived Codex session directories")
	flags.StringVar(&parsed.RateCard, "rate-card", defaults.RateCard, "TOML rate card")
	flags.StringVar(&parsed.SessionIndex, "session-index", defaults.SessionIndex, "Codex session title index")
	flags.StringVar(&parsed.GlobalState, "global-state", defaults.GlobalState, "Codex desktop global state containing UI project assignments")
	flags.StringVar(&parsed.ComputerName, "computer-name", "", "Override the detected computer name")
	flags.StringVar(&parsed.Format, "format", defaults.Format, "markdown or json")
	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if parsed.Format != "markdown" && parsed.Format != "json" {
		return options{}, fmt.Errorf("--format은 markdown 또는 json이어야 함")
	}
	parsed.SessionsRoot = expandUser(parsed.SessionsRoot)
	parsed.RateCard = expandUser(parsed.RateCard)
	parsed.SessionIndex = expandUser(parsed.SessionIndex)
	parsed.GlobalState = expandUser(parsed.GlobalState)
	return parsed, nil
}

func aggregate(
	roots []string,
	targetDate time.Time,
	location *time.Location,
	timezoneName string,
	rateCard string,
	sessionIndex string,
	globalState string,
	computerName string,
) (report, error) {
	rangeStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, location)
	nextDate := targetDate.AddDate(0, 0, 1)
	rangeEnd := time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), 0, 0, 0, 0, location)
	rates, err := loadRates(rateCard)
	if err != nil {
		return report{}, err
	}
	sessionTitles := loadSessionTitles(sessionIndex)
	projectAssignments := loadProjectAssignments(globalState)
	files := collectSessionFiles(roots)
	outcomes := processFiles(files, rangeStart.UTC(), rangeEnd.UTC())

	projects := make(map[string]totals)
	models := make(map[string]totals)
	sessions := make(map[string]map[string]*sessionUsage)
	overall := totals{}
	diagnostics := diagnostics{}
	seen := make(map[eventIdentity]struct{})

	for _, outcome := range outcomes {
		diagnostics.merge(outcome.Diagnostics)
		project := projectAssignments[outcome.ReportSessionID]
		if project == "" {
			project = "미분류"
		}
		for _, event := range outcome.Candidates {
			if _, duplicate := seen[event.Identity]; duplicate {
				diagnostics.DuplicateEvents++
				continue
			}
			seen[event.Identity] = struct{}{}
			eventTotals, err := totalsFromUsage(event.Usage, rateFor(rates, event.Model, targetDate))
			if err != nil {
				diagnostics.InvalidTokenEvents++
				continue
			}

			label := modelLabel(event.Model)
			projectTotal := projects[project]
			projectTotal.merge(eventTotals)
			projects[project] = projectTotal
			modelTotal := models[label]
			modelTotal.merge(eventTotals)
			models[label] = modelTotal

			projectSessions := sessions[project]
			if projectSessions == nil {
				projectSessions = make(map[string]*sessionUsage)
				sessions[project] = projectSessions
			}
			session := projectSessions[outcome.ReportSessionID]
			if session == nil {
				title := sessionTitles[outcome.ReportSessionID]
				if title == "" {
					title = "제목 미확인"
				}
				session = &sessionUsage{
					SessionID: outcome.ReportSessionID,
					Title:     title,
					Models:    make(map[string]struct{}),
				}
				projectSessions[outcome.ReportSessionID] = session
			}
			session.Models[label] = struct{}{}
			session.Totals.merge(eventTotals)
			overall.merge(eventTotals)
			diagnostics.AggregatedEvents++
		}
	}

	return report{
		TargetDate:   targetDate,
		TimezoneName: timezoneName,
		RangeStart:   rangeStart,
		RangeEnd:     rangeEnd,
		GeneratedAt:  time.Now().In(location),
		ComputerName: computerName,
		Projects:     projects,
		Models:       models,
		Sessions:     sessions,
		Total:        overall,
		Diagnostics:  diagnostics,
	}, nil
}

func parseTargetDate(value string, location *time.Location) (time.Time, error) {
	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	switch value {
	case "today":
		return today, nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	default:
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, fmt.Errorf("--date는 today, yesterday 또는 YYYY-MM-DD여야 함")
		}
		return parsed, nil
	}
}

func defaultSkillRoot() string {
	if root := os.Getenv("REPORT_CODEX_USAGE_SKILL_ROOT"); root != "" {
		return filepath.Clean(root)
	}
	_, source, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	}
	return "."
}

func expandUser(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	if len(path) > 1 && (path[1] == '/' || path[1] == filepath.Separator) {
		return filepath.Join(home, path[2:])
	}
	return path
}

func detectComputerName() string {
	commands := [][]string{{"scutil", "--get", "ComputerName"}, {"hostname"}}
	for _, command := range commands {
		output, err := exec.Command(command[0], command[1:]...).Output()
		if err == nil {
			if value := string(bytesTrimSpace(output)); value != "" {
				return value
			}
		}
	}
	return "확인 불가"
}

func bytesTrimSpace(value []byte) []byte {
	start := 0
	for start < len(value) && (value[start] == ' ' || value[start] == '\n' || value[start] == '\r' || value[start] == '\t') {
		start++
	}
	end := len(value)
	for end > start && (value[end-1] == ' ' || value[end-1] == '\n' || value[end-1] == '\r' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
}

func run(arguments []string, output io.Writer) (int, error) {
	options, err := parseOptions(arguments)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
		return 2, err
	}
	location, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return 2, fmt.Errorf("알 수 없는 timezone: %s", options.Timezone)
	}
	targetDate, err := parseTargetDate(options.Date, location)
	if err != nil {
		return 2, err
	}
	home, _ := os.UserHomeDir()
	roots := []string{options.SessionsRoot}
	if options.SessionsRoot == "" {
		roots = []string{filepath.Join(home, ".codex", "sessions")}
		archived := filepath.Join(home, ".codex", "archived_sessions")
		if stat, err := os.Stat(archived); err == nil && stat.IsDir() {
			roots = append(roots, archived)
		}
	}
	if stat, err := os.Stat(roots[0]); err != nil || !stat.IsDir() {
		return 2, fmt.Errorf("sessions root가 존재하지 않음: %s", roots[0])
	}
	if stat, err := os.Stat(options.RateCard); err != nil || stat.IsDir() {
		return 2, fmt.Errorf("rate card가 존재하지 않음: %s", options.RateCard)
	}
	computerName := options.ComputerName
	if computerName == "" {
		computerName = detectComputerName()
	}

	report, err := aggregate(
		roots,
		targetDate,
		location,
		options.Timezone,
		options.RateCard,
		options.SessionIndex,
		options.GlobalState,
		computerName,
	)
	if err != nil {
		return 2, err
	}
	failures := make(map[string]uint64)
	if report.Diagnostics.MalformedLines > 0 {
		failures["malformed_lines"] = report.Diagnostics.MalformedLines
	}
	if report.Diagnostics.UnreadableFiles > 0 {
		failures["unreadable_files"] = report.Diagnostics.UnreadableFiles
	}
	if report.Diagnostics.InvalidTokenEvents > 0 {
		failures["invalid_token_events"] = report.Diagnostics.InvalidTokenEvents
	}
	if len(failures) > 0 {
		encoded, _ := json.Marshal(failures)
		return 2, fmt.Errorf("로그 또는 token_count metadata가 불완전함: %s", encoded)
	}
	if err := assertReportIntegrity(report); err != nil {
		return 2, err
	}
	if options.Format == "json" {
		encoded, err := reportToJSON(report)
		if err != nil {
			return 2, err
		}
		fmt.Fprintln(output, encoded)
	} else {
		fmt.Fprintln(output, renderMarkdown(report))
	}
	return 0, nil
}

func main() {
	code, err := run(os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "집계 실패: %v\n", err)
	}
	os.Exit(code)
}
