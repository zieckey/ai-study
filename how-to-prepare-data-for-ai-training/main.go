package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type RawDocument struct {
	ID     string
	Source string
	Text   string
}

type CleanDocument struct {
	ID           string
	Source       string
	Text         string
	Language     string
	QualityScore float64
}

type Chunk struct {
	ID         string
	DocumentID string
	Text       string
	TokenCount int
}

type DatasetSplit struct {
	Train      []Chunk
	Validation []Chunk
	Test       []Chunk
}

type TopToken struct {
	Token string `json:"token"`
	Count int    `json:"count"`
}

type Manifest struct {
	RawDocuments          int        `json:"raw_documents"`
	CleanedDocuments      int        `json:"cleaned_documents"`
	DuplicateDocuments    int        `json:"duplicate_documents"`
	FilteredDocuments     int        `json:"filtered_documents"`
	FinalDocuments        int        `json:"final_documents"`
	Chunks                int        `json:"chunks"`
	TrainChunks           int        `json:"train_chunks"`
	ValidationChunks      int        `json:"validation_chunks"`
	TestChunks            int        `json:"test_chunks"`
	TotalTokens           int        `json:"total_tokens"`
	AverageTokensPerChunk float64    `json:"average_tokens_per_chunk"`
	TopTokens             []TopToken `json:"top_tokens"`
}

type PipelineConfig struct {
	MinQuality      float64
	MinTokens       int
	MaxChunkTokens  int
	ChunkOverlap    int
	TrainRatio      float64
	ValidationRatio float64
	TopTokenLimit   int
}

type PipelineResult struct {
	CleanedDocuments   []CleanDocument
	FinalDocuments     []CleanDocument
	DuplicateDocuments int
	FilteredDocuments  int
	Chunks             []Chunk
	Split              DatasetSplit
	Manifest           Manifest
}

var (
	emailPattern  = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	phonePattern  = regexp.MustCompile(`\b1[3-9]\d{9}\b`)
	idCardPattern = regexp.MustCompile(`\b\d{17}[0-9Xx]\b`)
)

func main() {
	config := defaultPipelineConfig()
	rawDocuments := sampleRawDocuments()
	result, err := prepareDataset(rawDocuments, config)
	if err != nil {
		panic(err)
	}

	fmt.Println("LLM 训练数据准备示例")
	fmt.Println()
	fmt.Println("1. 收集原始数据")
	fmt.Printf("Raw documents: %d\n", len(rawDocuments))
	fmt.Println()

	fmt.Println("2. 文本清洗与隐私脱敏")
	fmt.Printf("Cleaned documents: %d\n", len(result.CleanedDocuments))
	for _, doc := range result.CleanedDocuments {
		if strings.Contains(doc.Text, "[EMAIL]") || strings.Contains(doc.Text, "[PHONE]") || strings.Contains(doc.Text, "[ID_CARD]") {
			fmt.Printf("PII redacted sample: %s\n", doc.Text)
			break
		}
	}
	fmt.Println()

	fmt.Println("3. 去重")
	fmt.Printf("Removed duplicates: %d\n", result.DuplicateDocuments)
	fmt.Println()

	fmt.Println("4. 语言识别与质量过滤")
	fmt.Printf("Kept documents: %d\n", len(result.FinalDocuments))
	fmt.Printf("Filtered documents: %d\n", result.FilteredDocuments)
	for _, doc := range result.FinalDocuments {
		fmt.Printf("- %s language=%s quality=%.2f\n", doc.ID, doc.Language, doc.QualityScore)
	}
	fmt.Println()

	fmt.Println("5. Chunk 切分")
	fmt.Printf("Chunks: %d\n", len(result.Chunks))
	if len(result.Chunks) > 0 {
		first := result.Chunks[0]
		fmt.Printf("First chunk: %s tokens=%d text=%q\n", first.ID, first.TokenCount, first.Text)
	}
	fmt.Println()

	fmt.Println("6. 数据集划分")
	fmt.Printf("Train: %d, Validation: %d, Test: %d\n", len(result.Split.Train), len(result.Split.Validation), len(result.Split.Test))
	fmt.Println()

	fmt.Println("7. Token 统计")
	for _, token := range result.Manifest.TopTokens {
		fmt.Printf("- %s: %d\n", token.Token, token.Count)
	}
	fmt.Println()

	fmt.Println("8. Manifest")
	manifestJSON, err := json.MarshalIndent(result.Manifest, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(manifestJSON))
}

func defaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		MinQuality:      0.45,
		MinTokens:       8,
		MaxChunkTokens:  24,
		ChunkOverlap:    6,
		TrainRatio:      0.8,
		ValidationRatio: 0.1,
		TopTokenLimit:   8,
	}
}

func sampleRawDocuments() []RawDocument {
	return []RawDocument{
		{
			ID:     "doc-001",
			Source: "tutorial",
			Text:   "大模型训练数据通常来自网页、书籍、代码、问答和人工标注数据。原始文本需要先清洗、去重、过滤，再转换成模型可以学习的 token 序列。",
		},
		{
			ID:     "doc-002",
			Source: "tutorial-copy",
			Text:   " 大模型训练数据通常来自网页、书籍、代码、问答和人工标注数据。\n原始文本需要先清洗、去重、过滤，再转换成模型可以学习的 token 序列。 ",
		},
		{
			ID:     "doc-003",
			Source: "blog",
			Text:   "Before model training, a data pipeline cleans documents, removes duplicates, filters low quality text, splits data into chunks, and records a reproducible manifest.",
		},
		{
			ID:     "doc-004",
			Source: "support-ticket",
			Text:   "用户反馈：我的邮箱是 alice@example.com，手机号是 13800138000，身份证号是 11010119900307123X。请在训练前删除或替换这些个人信息。",
		},
		{
			ID:     "doc-005",
			Source: "mixed-note",
			Text:   "Tokenization 把文本转换成 token。中文、English words、数字 2026 和标点会被切成更小的训练单元。",
		},
		{
			ID:     "doc-006",
			Source: "short-message",
			Text:   "好",
		},
		{
			ID:     "doc-007",
			Source: "spam",
			Text:   "!!!!!! $$$$ !!!!! 点击领取大奖!!!!!! 加微信!!!!!!",
		},
		{
			ID:     "doc-008",
			Source: "manual",
			Text:   "高质量训练数据需要记录来源、处理版本、过滤规则和统计指标。这样当模型效果异常时，团队可以回溯是哪一批数据发生了变化。",
		},
	}
}

func prepareDataset(rawDocuments []RawDocument, config PipelineConfig) (PipelineResult, error) {
	cleanedDocuments := make([]CleanDocument, 0, len(rawDocuments))
	for _, raw := range rawDocuments {
		text := redactPII(cleanText(raw.Text))
		cleanedDocuments = append(cleanedDocuments, CleanDocument{
			ID:           raw.ID,
			Source:       raw.Source,
			Text:         text,
			Language:     detectLanguage(text),
			QualityScore: qualityScore(text),
		})
	}

	uniqueDocuments, duplicateCount := deduplicate(cleanedDocuments)
	finalDocuments, filteredCount := filterDocuments(uniqueDocuments, config.MinQuality, config.MinTokens)

	chunks := make([]Chunk, 0, len(finalDocuments))
	for _, doc := range finalDocuments {
		docChunks, err := chunkDocument(doc, config.MaxChunkTokens, config.ChunkOverlap)
		if err != nil {
			return PipelineResult{}, err
		}
		chunks = append(chunks, docChunks...)
	}

	split := splitDataset(chunks, config.TrainRatio, config.ValidationRatio)
	manifest := buildManifest(len(rawDocuments), len(cleanedDocuments), duplicateCount, filteredCount, len(finalDocuments), chunks, split, config.TopTokenLimit)

	return PipelineResult{
		CleanedDocuments:   cleanedDocuments,
		FinalDocuments:     finalDocuments,
		DuplicateDocuments: duplicateCount,
		FilteredDocuments:  filteredCount,
		Chunks:             chunks,
		Split:              split,
		Manifest:           manifest,
	}, nil
}

func cleanText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			builder.WriteRune(' ')
		case unicode.IsControl(r):
			continue
		default:
			builder.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func redactPII(text string) string {
	text = emailPattern.ReplaceAllString(text, "[EMAIL]")
	text = phonePattern.ReplaceAllString(text, "[PHONE]")
	text = idCardPattern.ReplaceAllString(text, "[ID_CARD]")
	return text
}

func detectLanguage(text string) string {
	chineseCount := 0
	latinCount := 0
	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			chineseCount++
		case isASCIILetter(r):
			latinCount++
		}
	}

	if chineseCount == 0 && latinCount == 0 {
		return "unknown"
	}
	if chineseCount > 0 && latinCount > 0 {
		if chineseCount > latinCount*2 {
			return "zh"
		}
		if latinCount > chineseCount*2 {
			return "en"
		}
		return "mixed"
	}
	if chineseCount > 0 {
		return "zh"
	}
	return "en"
}

func qualityScore(text string) float64 {
	tokens := simpleTokenize(text)
	if len(tokens) == 0 {
		return 0
	}

	total := 0
	textLike := 0
	punctuation := 0
	maxRepeat := 1
	currentRepeat := 0
	var previous rune

	for _, r := range text {
		if unicode.IsSpace(r) {
			continue
		}
		total++
		if r == previous {
			currentRepeat++
		} else {
			currentRepeat = 1
			previous = r
		}
		if currentRepeat > maxRepeat {
			maxRepeat = currentRepeat
		}

		switch {
		case unicode.Is(unicode.Han, r), isASCIILetter(r), unicode.IsDigit(r):
			textLike++
		case unicode.IsPunct(r):
			punctuation++
		}
	}
	if total == 0 {
		return 0
	}

	textRatio := float64(textLike) / float64(total)
	punctuationRatio := float64(punctuation) / float64(total)
	lengthScore := minFloat(1, float64(len(tokens))/20)
	repeatPenalty := minFloat(0.4, maxFloat(0, float64(maxRepeat-4)*0.06))
	score := 0.55*textRatio + 0.35*lengthScore + 0.10*(1-punctuationRatio) - repeatPenalty
	return clamp(score, 0, 1)
}

func normalizeForDedup(text string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func deduplicate(docs []CleanDocument) ([]CleanDocument, int) {
	seen := make(map[string]struct{}, len(docs))
	unique := make([]CleanDocument, 0, len(docs))
	duplicates := 0

	for _, doc := range docs {
		key := normalizeForDedup(doc.Text)
		if _, ok := seen[key]; ok {
			duplicates++
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, doc)
	}

	return unique, duplicates
}

func filterDocuments(docs []CleanDocument, minQuality float64, minTokens int) ([]CleanDocument, int) {
	kept := make([]CleanDocument, 0, len(docs))
	filtered := 0
	for _, doc := range docs {
		if doc.Language == "unknown" || doc.QualityScore < minQuality || len(simpleTokenize(doc.Text)) < minTokens {
			filtered++
			continue
		}
		kept = append(kept, doc)
	}
	return kept, filtered
}

func simpleTokenize(text string) []string {
	var tokens []string
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(current.String()))
		current.Reset()
	}

	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			flush()
			tokens = append(tokens, string(r))
		case isASCIILetter(r), unicode.IsDigit(r), r == '_':
			current.WriteRune(r)
		case unicode.IsSpace(r):
			flush()
		case unicode.IsPunct(r):
			flush()
			tokens = append(tokens, string(r))
		default:
			flush()
		}
	}
	flush()

	return tokens
}

func chunkDocument(doc CleanDocument, maxTokens int, overlap int) ([]Chunk, error) {
	if maxTokens <= 0 {
		return nil, errors.New("maxTokens must be positive")
	}
	if overlap < 0 || overlap >= maxTokens {
		return nil, errors.New("overlap must be non-negative and smaller than maxTokens")
	}

	tokens := simpleTokenize(doc.Text)
	if len(tokens) == 0 {
		return nil, nil
	}

	step := maxTokens - overlap
	chunks := make([]Chunk, 0, (len(tokens)+step-1)/step)
	for start := 0; start < len(tokens); start += step {
		end := start + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}
		chunkTokens := tokens[start:end]
		chunks = append(chunks, Chunk{
			ID:         fmt.Sprintf("%s-chunk-%02d", doc.ID, len(chunks)+1),
			DocumentID: doc.ID,
			Text:       strings.Join(chunkTokens, " "),
			TokenCount: len(chunkTokens),
		})
		if end == len(tokens) {
			break
		}
	}

	return chunks, nil
}

func splitDataset(chunks []Chunk, trainRatio, validationRatio float64) DatasetSplit {
	if len(chunks) == 0 {
		return DatasetSplit{}
	}

	shuffled := append([]Chunk(nil), chunks...)
	rng := rand.New(rand.NewSource(42))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	trainCount := int(float64(len(shuffled)) * trainRatio)
	validationCount := int(float64(len(shuffled)) * validationRatio)
	if len(shuffled) == 1 {
		trainCount = 1
		validationCount = 0
	} else if len(shuffled) == 2 {
		trainCount = 1
		validationCount = 0
	} else {
		if trainCount == 0 {
			trainCount = 1
		}
		if validationRatio > 0 && validationCount == 0 {
			validationCount = 1
		}
		if trainCount+validationCount >= len(shuffled) {
			trainCount = len(shuffled) - 2
			validationCount = 1
		}
	}

	testStart := trainCount + validationCount
	return DatasetSplit{
		Train:      shuffled[:trainCount],
		Validation: shuffled[trainCount:testStart],
		Test:       shuffled[testStart:],
	}
}

func tokenFrequency(chunks []Chunk) map[string]int {
	frequency := make(map[string]int)
	for _, chunk := range chunks {
		for _, token := range simpleTokenize(chunk.Text) {
			if isWordToken(token) {
				frequency[token]++
			}
		}
	}
	return frequency
}

func topTokens(frequency map[string]int, limit int) []TopToken {
	tokens := make([]TopToken, 0, len(frequency))
	for token, count := range frequency {
		tokens = append(tokens, TopToken{Token: token, Count: count})
	}
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Count == tokens[j].Count {
			return tokens[i].Token < tokens[j].Token
		}
		return tokens[i].Count > tokens[j].Count
	})
	if limit > 0 && len(tokens) > limit {
		return tokens[:limit]
	}
	return tokens
}

func buildManifest(rawCount, cleanedCount, duplicateCount, filteredCount, finalCount int, chunks []Chunk, split DatasetSplit, topTokenLimit int) Manifest {
	totalTokens := 0
	for _, chunk := range chunks {
		totalTokens += chunk.TokenCount
	}

	averageTokens := 0.0
	if len(chunks) > 0 {
		averageTokens = float64(totalTokens) / float64(len(chunks))
	}

	return Manifest{
		RawDocuments:          rawCount,
		CleanedDocuments:      cleanedCount,
		DuplicateDocuments:    duplicateCount,
		FilteredDocuments:     filteredCount,
		FinalDocuments:        finalCount,
		Chunks:                len(chunks),
		TrainChunks:           len(split.Train),
		ValidationChunks:      len(split.Validation),
		TestChunks:            len(split.Test),
		TotalTokens:           totalTokens,
		AverageTokensPerChunk: roundToTwoDecimals(averageTokens),
		TopTokens:             topTokens(tokenFrequency(chunks), topTokenLimit),
	}
}

func isWordToken(token string) bool {
	for _, r := range token {
		if unicode.Is(unicode.Han, r) || isASCIILetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func clamp(value, low, high float64) float64 {
	return minFloat(high, maxFloat(low, value))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
