package ollama

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

var (
	totalDurationRe      = regexp.MustCompile(`total duration:\s+(.+)`)
	loadDurationRe       = regexp.MustCompile(`load duration:\s+(.+)`)
	promptEvalCountRe    = regexp.MustCompile(`prompt eval count:\s+(\d+)`)
	promptEvalDurationRe = regexp.MustCompile(`prompt eval duration:\s+(.+)`)
	promptEvalRateRe     = regexp.MustCompile(`prompt eval rate:\s+([\d.]+)`)
	// Match "eval ..." at line start only, so "prompt eval ..." (which starts with "prompt") isn't matched
	evalCountRe    = regexp.MustCompile(`(?m)^eval count:\s+(\d+)`)
	evalDurationRe = regexp.MustCompile(`(?m)^eval duration:\s+(.+)`)
	evalRateRe     = regexp.MustCompile(`(?m)^eval rate:\s+([\d.]+)`)
)

func ParseVerboseOutput(output string) (models.OllamaStats, error) {
	var stats models.OllamaStats
	var err error

	stats.TotalDurationS, err = parseDurationField(totalDurationRe, output, "total duration")
	if err != nil {
		return stats, err
	}

	stats.LoadDurationS, err = parseDurationField(loadDurationRe, output, "load duration")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalCount, err = parseIntField(promptEvalCountRe, output, "prompt eval count")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalDurationS, err = parseDurationField(promptEvalDurationRe, output, "prompt eval duration")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalRate, err = parseFloatField(promptEvalRateRe, output, "prompt eval rate")
	if err != nil {
		return stats, err
	}

	stats.EvalCount, err = parseIntField(evalCountRe, output, "eval count")
	if err != nil {
		return stats, err
	}

	stats.EvalDurationS, err = parseDurationField(evalDurationRe, output, "eval duration")
	if err != nil {
		return stats, err
	}

	stats.EvalRate, err = parseFloatField(evalRateRe, output, "eval rate")
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func parseDurationField(re *regexp.Regexp, output, name string) (float64, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return parseDuration(strings.TrimSpace(match[1]))
}

func parseIntField(re *regexp.Regexp, output, name string) (int, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return strconv.Atoi(strings.TrimSpace(match[1]))
}

func parseFloatField(re *regexp.Regexp, output, name string) (float64, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return strconv.ParseFloat(strings.TrimSpace(match[1]), 64)
}

// parseDuration handles Go-style durations: "5.227891234s", "152.345ms", "1m12.423928897s"
func parseDuration(s string) (float64, error) {
	s = strings.TrimSpace(s)

	var totalSeconds float64

	// Handle minutes component: "1m12.423s" -> minutes=1, remainder="12.423s"
	// Skip if "m" is part of "ms" suffix (e.g. "152.345ms")
	if idx := strings.Index(s, "m"); idx > 0 && (idx+1 >= len(s) || s[idx+1] != 's') {
		minutes, err := strconv.ParseFloat(s[:idx], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in duration %q: %w", s, err)
		}
		totalSeconds += minutes * 60
		s = s[idx+1:]
	}

	if strings.HasSuffix(s, "ms") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ms duration %q: %w", s, err)
		}
		totalSeconds += val / 1000
	} else if strings.HasSuffix(s, "s") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds duration %q: %w", s, err)
		}
		totalSeconds += val
	} else if s != "" {
		return 0, fmt.Errorf("unknown duration format: %q", s)
	}

	return totalSeconds, nil
}
