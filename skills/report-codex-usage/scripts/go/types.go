package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var modelLabelOrder = []string{"sol", "terra", "luna", "5.5", "5.4", "review", "other"}

type rate struct {
	Model                 string
	EffectiveFrom         time.Time
	InputPerMillion       float64
	CachedInputPerMillion float64
	OutputPerMillion      float64
	CacheWritePerMillion  float64
}

type totals struct {
	Total          uint64  `json:"total"`
	CachedInput    uint64  `json:"cached_input"`
	Input          uint64  `json:"input"`
	Output         uint64  `json:"output"`
	CalculatedCost float64 `json:"calculated_cost"`
	Events         uint64  `json:"events"`
}

func totalsFromUsage(raw json.RawMessage, selectedRate *rate) (totals, error) {
	var usage map[string]json.RawMessage
	if err := json.Unmarshal(raw, &usage); err != nil || usage == nil {
		return totals{}, fmt.Errorf("last_token_usage must be an object")
	}

	inputTokens, err := integerToken(usage["input_tokens"], "input_tokens")
	if err != nil {
		return totals{}, err
	}
	cachedInput, err := integerToken(usage["cached_input_tokens"], "cached_input_tokens")
	if err != nil {
		return totals{}, err
	}
	cacheWrite, err := integerToken(usage["cache_write_input_tokens"], "cache_write_input_tokens")
	if err != nil {
		return totals{}, err
	}
	output, err := integerToken(usage["output_tokens"], "output_tokens")
	if err != nil {
		return totals{}, err
	}

	if cachedInput > inputTokens || cacheWrite > inputTokens-cachedInput {
		return totals{}, fmt.Errorf("cached_input_tokens + cache_write_input_tokens exceeds input_tokens")
	}

	result := totals{
		Total:       inputTokens + output,
		CachedInput: cachedInput,
		Input:       inputTokens - cachedInput,
		Output:      output,
		Events:      1,
	}
	if selectedRate != nil {
		regularInput := inputTokens - cachedInput - cacheWrite
		result.CalculatedCost = (float64(regularInput)*selectedRate.InputPerMillion +
			float64(cachedInput)*selectedRate.CachedInputPerMillion +
			float64(cacheWrite)*selectedRate.CacheWritePerMillion +
			float64(output)*selectedRate.OutputPerMillion) / 1_000_000
	}
	return result, nil
}

func integerToken(raw json.RawMessage, key string) (uint64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	value, err := number.Int64()
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return uint64(value), nil
}

func (current *totals) merge(other totals) {
	current.Total += other.Total
	current.CachedInput += other.CachedInput
	current.Input += other.Input
	current.Output += other.Output
	current.CalculatedCost += other.CalculatedCost
	current.Events += other.Events
}

type diagnostics struct {
	FilesScanned            uint64 `json:"files_scanned"`
	FilesWithTargetEvents   uint64 `json:"files_with_target_events"`
	OriginalEvents          uint64 `json:"original_events"`
	DuplicateEvents         uint64 `json:"duplicate_events"`
	ReplayedEvents          uint64 `json:"replayed_events"`
	AggregatedEvents        uint64 `json:"aggregated_events"`
	MalformedLines          uint64 `json:"malformed_lines"`
	UnreadableFiles         uint64 `json:"unreadable_files"`
	TokenEventsWithoutUsage uint64 `json:"token_events_without_usage"`
	InvalidTokenEvents      uint64 `json:"invalid_token_events"`
}

func (current *diagnostics) merge(other diagnostics) {
	current.FilesScanned += other.FilesScanned
	current.FilesWithTargetEvents += other.FilesWithTargetEvents
	current.OriginalEvents += other.OriginalEvents
	current.DuplicateEvents += other.DuplicateEvents
	current.ReplayedEvents += other.ReplayedEvents
	current.AggregatedEvents += other.AggregatedEvents
	current.MalformedLines += other.MalformedLines
	current.UnreadableFiles += other.UnreadableFiles
	current.TokenEventsWithoutUsage += other.TokenEventsWithoutUsage
	current.InvalidTokenEvents += other.InvalidTokenEvents
}

type sessionUsage struct {
	SessionID string
	Title     string
	Models    map[string]struct{}
	Totals    totals
}

type report struct {
	TargetDate   time.Time
	TimezoneName string
	RangeStart   time.Time
	RangeEnd     time.Time
	GeneratedAt  time.Time
	ComputerName string
	Projects     map[string]totals
	Models       map[string]totals
	Sessions     map[string]map[string]*sessionUsage
	Total        totals
	Diagnostics  diagnostics
}

func modelLabel(model string) string {
	normalized := strings.ToLower(model)
	switch {
	case normalized == "gpt-5.6", normalized == "gpt-5.6-sol", strings.Contains(normalized, "5.6-sol"):
		return "sol"
	case strings.Contains(normalized, "terra"):
		return "terra"
	case strings.Contains(normalized, "luna"):
		return "luna"
	case normalized == "gpt-5.5", strings.HasPrefix(normalized, "gpt-5.5-"):
		return "5.5"
	case normalized == "gpt-5.4", strings.HasPrefix(normalized, "gpt-5.4-"):
		return "5.4"
	case normalized == "codex-auto-review", strings.Contains(normalized, "auto-review"):
		return "review"
	default:
		return "other"
	}
}

func displayModels(models map[string]struct{}) string {
	labels := make([]string, 0, len(models))
	for _, label := range modelLabelOrder {
		if _, exists := models[label]; exists {
			labels = append(labels, label)
		}
	}
	return strings.Join(labels, ", ")
}
