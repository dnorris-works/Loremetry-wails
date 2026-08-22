package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

// LocalConfig holds DB-driven parameters for local metric runners.
type LocalConfig struct {
	FilterWords   []string `json:"filter_words"`
	EchoWindow      int      `json:"echo_window"`
	GlueThreshold   int      `json:"glue_threshold"`
	OpenLoopCues    []string `json:"open_loop_cues"`
	WordsPerPage    int      `json:"words_per_page"`
}

func ParseLocalConfig(raw string) LocalConfig {
	var cfg LocalConfig
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return cfg
	}
	_ = json.Unmarshal([]byte(raw), &cfg)
	return cfg
}

type LocalRunOutcome struct {
	Content  string
	Original string
	Proposed string
	DataJSON string
}

type localRunnerFunc func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome

func localRunnerRegistry() map[string]localRunnerFunc {
	return map[string]localRunnerFunc{
		"print_production": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			wpp := cfg.WordsPerPage
			if wpp <= 0 {
				wpp = 250
			}
			content := runPrintProduction(manuscript, wpp)
			return LocalRunOutcome{
				Content:  content,
				DataJSON: fmt.Sprintf(`{"words":%d}`, len(strings.Fields(manuscript))),
			}
		},
		"line_polish": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{
				Content:  runLinePolish(manuscript, cfg),
				DataJSON: `{"kind":"line_polish"}`,
			}
		},
		"vellum_prep": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{
				Content:  runVellumPrep(blobs),
				DataJSON: `{"kind":"vellum_prep"}`,
			}
		},
		"zeigarnik_analysis": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			original, proposed, dataJSON := runZeigarnik(blobs, cfg)
			return LocalRunOutcome{Original: original, Proposed: proposed, DataJSON: dataJSON}
		},
		"readability_score": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runReadabilityScore(blobs), DataJSON: `{"kind":"readability_score"}`}
		},
		"sentence_length_variation": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runSentenceLengthVariation(blobs), DataJSON: `{"kind":"sentence_length_variation"}`}
		},
		"chapter_balance": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runChapterBalance(blobs), DataJSON: `{"kind":"chapter_balance"}`}
		},
		"dialogue_ratio": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runDialogueRatio(blobs), DataJSON: `{"kind":"dialogue_ratio"}`}
		},
		"dialogue_tag_audit": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runDialogueTagAudit(blobs), DataJSON: `{"kind":"dialogue_tag_audit"}`}
		},
		"passive_voice": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runPassiveVoice(blobs), DataJSON: `{"kind":"passive_voice"}`}
		},
		"sticky_sentences": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			content, dataJSON := runStickySentences(blobs, cfg)
			return LocalRunOutcome{Content: content, DataJSON: dataJSON}
		},
		"repeated_phrases": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runRepeatedPhrases(blobs), DataJSON: `{"kind":"repeated_phrases"}`}
		},
		"paragraph_length": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runParagraphLength(blobs), DataJSON: `{"kind":"paragraph_length"}`}
		},
		"opening_closing": func(blobs []RoleText, manuscript string, cfg LocalConfig) LocalRunOutcome {
			return LocalRunOutcome{Content: runOpeningClosing(blobs), DataJSON: `{"kind":"opening_closing"}`}
		},
	}
}

func RunLocalFromCatalog(detail AnalysisDetail, blobs []RoleText) (LocalRunOutcome, error) {
	runnerKey := strings.TrimSpace(detail.LocalRunner)
	if runnerKey == "" {
		runnerKey = detail.ID
	}
	fn, ok := localRunnerRegistry()[runnerKey]
	if !ok {
		return LocalRunOutcome{}, fmt.Errorf("unknown local runner %q", runnerKey)
	}
	cfg := ParseLocalConfig(detail.LocalConfig)
	manuscript := joinRoleText(blobs, "manuscript")
	return fn(blobs, manuscript, cfg), nil
}
