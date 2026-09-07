package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestCleanTextNormalizesWhitespace(t *testing.T) {
	got := cleanText("  大模型\t训练\n\n数据   准备  ")
	want := "大模型 训练 数据 准备"
	if got != want {
		t.Fatalf("cleanText() = %q, want %q", got, want)
	}
}

func TestRedactPII(t *testing.T) {
	input := "邮箱 alice@example.com，手机号 13800138000，身份证 11010119900307123X。"
	got := redactPII(input)
	for _, sensitive := range []string{"alice@example.com", "13800138000", "11010119900307123X"} {
		if strings.Contains(got, sensitive) {
			t.Fatalf("redactPII() kept sensitive value %q in %q", sensitive, got)
		}
	}
	for _, marker := range []string{"[EMAIL]", "[PHONE]", "[ID_CARD]"} {
		if !strings.Contains(got, marker) {
			t.Fatalf("redactPII() missing marker %q in %q", marker, got)
		}
	}
}

func TestDeduplicateRemovesNormalizedDuplicates(t *testing.T) {
	docs := []CleanDocument{
		{ID: "a", Text: "Data Pipeline!"},
		{ID: "b", Text: "data pipeline"},
		{ID: "c", Text: "another document"},
	}

	unique, duplicates := deduplicate(docs)
	if duplicates != 1 {
		t.Fatalf("duplicates = %d, want 1", duplicates)
	}
	if len(unique) != 2 {
		t.Fatalf("len(unique) = %d, want 2", len(unique))
	}
	if unique[0].ID != "a" || unique[1].ID != "c" {
		t.Fatalf("unique IDs = %v, want [a c]", []string{unique[0].ID, unique[1].ID})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "Chinese", text: "大模型训练数据", want: "zh"},
		{name: "English", text: "large language model training data", want: "en"},
		{name: "Mixed", text: "训练 data", want: "mixed"},
		{name: "Unknown", text: "12345 !!!", want: "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := detectLanguage(test.text); got != test.want {
				t.Fatalf("detectLanguage(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

func TestQualityScoreRejectsNoisyText(t *testing.T) {
	normal := qualityScore("高质量训练数据需要清晰、完整、可追溯，并记录处理流程。")
	noisy := qualityScore("!!!!!! $$$$ !!!!! !!!!!")
	if normal <= noisy {
		t.Fatalf("normal score %.4f should be higher than noisy score %.4f", normal, noisy)
	}
	if noisy >= 0.45 {
		t.Fatalf("noisy score %.4f should be below filter threshold", noisy)
	}
}

func TestChunkDocumentUsesSlidingWindow(t *testing.T) {
	doc := CleanDocument{ID: "doc", Text: "a b c d e f g"}
	chunks, err := chunkDocument(doc, 4, 1)
	if err != nil {
		t.Fatalf("chunkDocument returned an error: %v", err)
	}

	gotTexts := []string{chunks[0].Text, chunks[1].Text}
	wantTexts := []string{"a b c d", "d e f g"}
	if !reflect.DeepEqual(gotTexts, wantTexts) {
		t.Fatalf("chunk texts = %v, want %v", gotTexts, wantTexts)
	}
	if chunks[0].TokenCount != 4 || chunks[1].TokenCount != 4 {
		t.Fatalf("token counts = [%d %d], want [4 4]", chunks[0].TokenCount, chunks[1].TokenCount)
	}
}

func TestChunkDocumentRejectsInvalidConfig(t *testing.T) {
	doc := CleanDocument{ID: "doc", Text: "a b c"}
	tests := []struct {
		name      string
		maxTokens int
		overlap   int
	}{
		{name: "ZeroMaxTokens", maxTokens: 0, overlap: 0},
		{name: "NegativeOverlap", maxTokens: 4, overlap: -1},
		{name: "OverlapEqualsMax", maxTokens: 4, overlap: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := chunkDocument(doc, test.maxTokens, test.overlap); err == nil {
				t.Fatal("chunkDocument returned nil error")
			}
		})
	}
}

func TestSplitDatasetIsDeterministic(t *testing.T) {
	chunks := []Chunk{
		{ID: "c1"}, {ID: "c2"}, {ID: "c3"}, {ID: "c4"}, {ID: "c5"},
		{ID: "c6"}, {ID: "c7"}, {ID: "c8"}, {ID: "c9"}, {ID: "c10"},
	}

	first := splitDataset(chunks, 0.8, 0.1)
	second := splitDataset(chunks, 0.8, 0.1)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("splitDataset is not deterministic")
	}
	if len(first.Train) != 8 || len(first.Validation) != 1 || len(first.Test) != 1 {
		t.Fatalf("split sizes = %d/%d/%d, want 8/1/1", len(first.Train), len(first.Validation), len(first.Test))
	}
	if !reflect.DeepEqual(chunks, []Chunk{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}, {ID: "c4"}, {ID: "c5"}, {ID: "c6"}, {ID: "c7"}, {ID: "c8"}, {ID: "c9"}, {ID: "c10"}}) {
		t.Fatal("splitDataset mutated input")
	}
}

func TestPipelineBuildsManifest(t *testing.T) {
	result, err := prepareDataset(sampleRawDocuments(), defaultPipelineConfig())
	if err != nil {
		t.Fatalf("prepareDataset returned an error: %v", err)
	}

	manifest := result.Manifest
	if manifest.RawDocuments != len(sampleRawDocuments()) {
		t.Fatalf("RawDocuments = %d, want %d", manifest.RawDocuments, len(sampleRawDocuments()))
	}
	if manifest.CleanedDocuments != len(sampleRawDocuments()) {
		t.Fatalf("CleanedDocuments = %d, want %d", manifest.CleanedDocuments, len(sampleRawDocuments()))
	}
	if manifest.DuplicateDocuments != 1 {
		t.Fatalf("DuplicateDocuments = %d, want 1", manifest.DuplicateDocuments)
	}
	if manifest.FinalDocuments == 0 {
		t.Fatal("FinalDocuments should be positive")
	}
	if manifest.Chunks != len(result.Chunks) {
		t.Fatalf("Chunks = %d, want %d", manifest.Chunks, len(result.Chunks))
	}
	if manifest.TrainChunks+manifest.ValidationChunks+manifest.TestChunks != manifest.Chunks {
		t.Fatalf("split chunk counts do not add up to total chunks")
	}
	if manifest.TotalTokens == 0 {
		t.Fatal("TotalTokens should be positive")
	}
	if len(manifest.TopTokens) == 0 {
		t.Fatal("TopTokens should not be empty")
	}
}
