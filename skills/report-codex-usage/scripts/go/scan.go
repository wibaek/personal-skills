package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	recordSessionMeta = iota
	recordModel
	recordTaskStarted
	recordTokenCount
)

type rawRow struct {
	Type      string          `json:"type"`
	Timestamp json.RawMessage `json:"timestamp"`
	Payload   *rawPayload     `json:"payload"`
}

type rawPayload struct {
	ID             json.RawMessage   `json:"id"`
	SessionID      json.RawMessage   `json:"session_id"`
	Model          json.RawMessage   `json:"model"`
	Type           json.RawMessage   `json:"type"`
	ThreadSettings *rawThreadSetting `json:"thread_settings"`
	Info           json.RawMessage   `json:"info"`
}

type rawThreadSetting struct {
	Model json.RawMessage `json:"model"`
}

type tokenInfo struct {
	TotalTokenUsage json.RawMessage `json:"total_token_usage"`
	LastTokenUsage  json.RawMessage `json:"last_token_usage"`
}

type projectedRecord struct {
	Kind         int
	ID           string
	SessionID    string
	Model        string
	Timestamp    time.Time
	HasTimestamp bool
	Info         json.RawMessage
}

type eventIdentity struct {
	RolloutID  string
	Timestamp  string
	TotalUsage string
	LastUsage  string
}

type candidate struct {
	Identity eventIdentity
	Model    string
	Usage    json.RawMessage
}

type fileOutcome struct {
	ReportSessionID string
	Candidates      []candidate
	Diagnostics     diagnostics
}

func processFile(path string, rangeStart, rangeEnd time.Time) fileOutcome {
	result := fileOutcome{
		ReportSessionID: pathStem(path),
		Diagnostics: diagnostics{
			FilesScanned: 1,
		},
	}
	file, err := os.Open(path)
	if err != nil {
		result.Diagnostics.UnreadableFiles++
		return result
	}
	defer file.Close()

	records := make([]projectedRecord, 0, 256)
	fileModels := make(map[string]struct{})
	err = eachJSONLine(file, func(line []byte) {
		var row rawRow
		if err := json.Unmarshal(line, &row); err != nil {
			result.Diagnostics.MalformedLines++
			return
		}
		payload := row.Payload
		if payload == nil {
			payload = &rawPayload{}
		}

		switch row.Type {
		case "session_meta":
			records = append(records, projectedRecord{
				Kind:      recordSessionMeta,
				ID:        decodedString(payload.ID),
				SessionID: decodedString(payload.SessionID),
			})
		case "turn_context":
			model := decodedString(payload.Model)
			if model != "" {
				fileModels[model] = struct{}{}
			}
			records = append(records, projectedRecord{Kind: recordModel, Model: model})
		case "event_msg":
			switch decodedString(payload.Type) {
			case "thread_settings_applied":
				model := ""
				if payload.ThreadSettings != nil {
					model = decodedString(payload.ThreadSettings.Model)
				}
				if model != "" {
					fileModels[model] = struct{}{}
				}
				records = append(records, projectedRecord{Kind: recordModel, Model: model})
			case "task_started":
				records = append(records, projectedRecord{Kind: recordTaskStarted})
			case "token_count":
				timestamp, hasTimestamp := parseTimestamp(row.Timestamp)
				records = append(records, projectedRecord{
					Kind:         recordTokenCount,
					Timestamp:    timestamp,
					HasTimestamp: hasTimestamp,
					Info:         payload.Info,
				})
			}
		}
	})
	if err != nil {
		result.Diagnostics.UnreadableFiles++
	}

	rolloutID := result.ReportSessionID
	reportSessionID := ""
	for _, record := range records {
		if record.Kind != recordSessionMeta {
			continue
		}
		if record.ID != "" {
			rolloutID = record.ID
		}
		reportSessionID = record.SessionID
		break
	}
	if reportSessionID == "" {
		reportSessionID = rolloutID
	}
	result.ReportSessionID = reportSessionID

	lastForeignMeta := -1
	for index, record := range records {
		if record.Kind == recordSessionMeta && record.ID != rolloutID {
			lastForeignMeta = index
		}
	}
	replayCutoff := -1
	if lastForeignMeta >= 0 {
		for index := lastForeignMeta + 1; index < len(records); index++ {
			if records[index].Kind == recordTaskStarted {
				replayCutoff = index
				break
			}
		}
	}
	replayOnlyRollout := lastForeignMeta >= 0 && replayCutoff < 0
	uniqueFileModel := ""
	if len(fileModels) == 1 {
		for model := range fileModels {
			uniqueFileModel = model
		}
	}

	currentModel := ""
	fileHasTargetEvent := false
	for index, record := range records {
		switch record.Kind {
		case recordModel:
			if record.Model != "" {
				currentModel = record.Model
			}
		case recordTokenCount:
			if !record.HasTimestamp || record.Timestamp.Before(rangeStart) || !record.Timestamp.Before(rangeEnd) {
				continue
			}
			result.Diagnostics.OriginalEvents++
			fileHasTargetEvent = true
			if replayOnlyRollout || (replayCutoff >= 0 && index < replayCutoff) {
				result.Diagnostics.ReplayedEvents++
				continue
			}

			var info *tokenInfo
			if len(record.Info) == 0 || json.Unmarshal(record.Info, &info) != nil || info == nil {
				result.Diagnostics.TokenEventsWithoutUsage++
				continue
			}
			if !isJSONObject(info.LastTokenUsage) {
				result.Diagnostics.TokenEventsWithoutUsage++
				continue
			}
			model := currentModel
			if model == "" {
				model = uniqueFileModel
			}
			if model == "" {
				model = "미분류"
			}
			result.Candidates = append(result.Candidates, candidate{
				Identity: eventIdentity{
					RolloutID:  rolloutID,
					Timestamp:  record.Timestamp.UTC().Format(time.RFC3339Nano),
					TotalUsage: canonicalJSON(info.TotalTokenUsage),
					LastUsage:  canonicalJSON(info.LastTokenUsage),
				},
				Model: model,
				Usage: append(json.RawMessage(nil), info.LastTokenUsage...),
			})
		}
	}
	if fileHasTargetEvent {
		result.Diagnostics.FilesWithTargetEvents++
	}
	return result
}

func eachJSONLine(reader io.Reader, visit func([]byte)) error {
	buffered := bufio.NewReaderSize(reader, 1024*1024)
	var accumulated []byte
	for {
		fragment, err := buffered.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			accumulated = append(accumulated, fragment...)
			continue
		}

		line := fragment
		if len(accumulated) > 0 {
			accumulated = append(accumulated, fragment...)
			line = accumulated
		}
		if len(line) > 0 {
			visit(line)
		}
		accumulated = accumulated[:0]

		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func decodedString(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return value
}

func parseTimestamp(raw json.RawMessage) (time.Time, bool) {
	value := decodedString(raw)
	if value == "" {
		return time.Time{}, false
	}
	if timestamp, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return timestamp.UTC(), true
	}
	for _, layout := range []string{"2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05"} {
		if timestamp, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return timestamp, true
		}
	}
	return time.Time{}, false
}

func isJSONObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}'
}

func canonicalJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "null"
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "null"
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(encoded)
}

func pathStem(path string) string {
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func collectSessionFiles(roots []string) []string {
	files := make([]string, 0, 4096)
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			if filepath.Ext(path) == ".jsonl" {
				files = append(files, path)
			}
			return nil
		})
	}
	sort.Strings(files)
	return files
}

func processFiles(files []string, rangeStart, rangeEnd time.Time) []fileOutcome {
	outcomes := make([]fileOutcome, len(files))
	if len(files) == 0 {
		return outcomes
	}
	workerCount := runtime.GOMAXPROCS(0)
	if workerCount > len(files) {
		workerCount = len(files)
	}
	jobs := make(chan int)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				outcomes[index] = processFile(files[index], rangeStart, rangeEnd)
			}
		}()
	}
	for index := range files {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	return outcomes
}

type sessionIndexRow struct {
	ID         string `json:"id"`
	ThreadName string `json:"thread_name"`
}

func loadSessionTitles(path string) map[string]string {
	titles := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return titles
	}
	defer file.Close()
	_ = eachJSONLine(file, func(line []byte) {
		var row sessionIndexRow
		if json.Unmarshal(line, &row) == nil && row.ID != "" && row.ThreadName != "" {
			titles[row.ID] = row.ThreadName
		}
	})
	return titles
}

type globalState struct {
	LocalProjects            map[string]localProject      `json:"local-projects"`
	ThreadProjectAssignments map[string]projectAssignment `json:"thread-project-assignments"`
}

type localProject struct {
	Name string `json:"name"`
}

type projectAssignment struct {
	ProjectKind string `json:"projectKind"`
	ProjectID   string `json:"projectId"`
}

func loadProjectAssignments(path string) map[string]string {
	assignments := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return assignments
	}
	defer file.Close()
	var state globalState
	if json.NewDecoder(file).Decode(&state) != nil {
		return assignments
	}
	projectNames := make(map[string]string)
	for projectID, project := range state.LocalProjects {
		if project.Name != "" {
			projectNames[projectID] = project.Name
		}
	}
	for threadID, assignment := range state.ThreadProjectAssignments {
		if assignment.ProjectKind != "local" {
			continue
		}
		if name := projectNames[assignment.ProjectID]; name != "" {
			assignments[threadID] = name
		}
	}
	return assignments
}
