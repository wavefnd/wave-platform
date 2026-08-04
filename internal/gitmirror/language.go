package gitmirror

import (
	"path"
	"sort"
	"strings"
)

var languagesByExtension = map[string]string{
	".rs":   "Rust",
	".wave": "Wave",
	".py":   "Python",
	".sh":   "Shell",
	".bash": "Shell",
	".asm":  "Assembly",
	".s":    "Assembly",
	".S":    "Assembly",
}

func DetectLanguages(entries []TreeEntry) []LanguageStat {
	statistics := make(map[string]*LanguageStat)
	var total int64

	for _, entry := range entries {
		if entry.Size <= 0 || entry.Binary || entry.Generated {
			continue
		}
		language := detectLanguage(entry.Path)
		if language == "" {
			continue
		}

		statistic := statistics[language]
		if statistic == nil {
			statistic = &LanguageStat{Name: language}
			statistics[language] = statistic
		}
		statistic.Bytes += entry.Size
		statistic.Files++
		total += entry.Size
	}

	result := make([]LanguageStat, 0, len(statistics))
	for _, statistic := range statistics {
		if total > 0 {
			statistic.Percentage = float64(statistic.Bytes) / float64(total) * 100
		}
		result = append(result, *statistic)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Bytes == result[right].Bytes {
			return result[left].Name < result[right].Name
		}
		return result[left].Bytes > result[right].Bytes
	})
	return result
}

func detectLanguage(filePath string) string {
	name := path.Base(filePath)
	switch strings.ToLower(name) {
	case "makefile", "gnumakefile":
		return "Makefile"
	case "dockerfile":
		return "Dockerfile"
	}
	return languagesByExtension[path.Ext(name)]
}
