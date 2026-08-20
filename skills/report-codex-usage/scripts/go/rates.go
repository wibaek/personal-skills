package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type rateBuilder struct {
	Rate          rate
	EffectiveFrom string
	Seen          map[string]bool
}

func loadRates(path string) (map[string][]rate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("rate card를 읽을 수 없음: %s: %w", path, err)
	}
	defer file.Close()

	rates := make(map[string][]rate)
	var current *rateBuilder
	flush := func() error {
		if current == nil {
			return nil
		}
		required := []string{
			"model",
			"effective_from",
			"input_per_million",
			"cached_input_per_million",
			"output_per_million",
			"cache_write_per_million",
		}
		for _, key := range required {
			if !current.Seen[key] {
				return fmt.Errorf("rate card 항목 누락: %s", key)
			}
		}
		effectiveFrom, err := time.Parse("2006-01-02", current.EffectiveFrom)
		if err != nil {
			return fmt.Errorf("rate card effective_from 형식이 잘못됨: %w", err)
		}
		current.Rate.EffectiveFrom = effectiveFrom
		rates[current.Rate.Model] = append(rates[current.Rate.Model], current.Rate)
		return nil
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		if line == "[[rate]]" {
			if err := flush(); err != nil {
				return nil, err
			}
			current = &rateBuilder{Seen: make(map[string]bool)}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || current == nil {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if err := assignRateValue(current, key, value); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("rate card를 읽을 수 없음: %w", err)
	}
	if err := flush(); err != nil {
		return nil, err
	}

	for model := range rates {
		sort.Slice(rates[model], func(left, right int) bool {
			return rates[model][left].EffectiveFrom.Before(rates[model][right].EffectiveFrom)
		})
	}
	return rates, nil
}

func assignRateValue(builder *rateBuilder, key, raw string) error {
	builder.Seen[key] = true
	switch key {
	case "model":
		value, err := strconv.Unquote(raw)
		if err != nil {
			return fmt.Errorf("rate card model 형식이 잘못됨: %w", err)
		}
		builder.Rate.Model = value
	case "effective_from":
		value, err := strconv.Unquote(raw)
		if err != nil {
			return fmt.Errorf("rate card effective_from 형식이 잘못됨: %w", err)
		}
		builder.EffectiveFrom = value
	case "input_per_million":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		builder.Rate.InputPerMillion = value
	case "cached_input_per_million":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		builder.Rate.CachedInputPerMillion = value
	case "output_per_million":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		builder.Rate.OutputPerMillion = value
	case "cache_write_per_million":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		builder.Rate.CacheWritePerMillion = value
	default:
		delete(builder.Seen, key)
	}
	return nil
}

func rateFor(rates map[string][]rate, model string, targetDate time.Time) *rate {
	modelRates := rates[model]
	for index := len(modelRates) - 1; index >= 0; index-- {
		if !modelRates[index].EffectiveFrom.After(targetDate) {
			return &modelRates[index]
		}
	}
	return nil
}
