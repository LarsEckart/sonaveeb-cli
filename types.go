package main

// API response types from Ekilex

type WordSearchResult struct {
	Words []WordMatch `json:"words"`
}

type WordMatch struct {
	WordID    int64  `json:"wordId"`
	WordValue string `json:"wordValue"`
	Lang      string `json:"lang"`
}

type WordDetails struct {
	WordClass string     `json:"wordClass"`
	Paradigms []Paradigm `json:"paradigms"`
	Lexemes   []Lexeme   `json:"lexemes"`
}

type Lexeme struct {
	Pos              []PosInfo        `json:"pos"`
	SynonymLangGroups []SynonymLangGroup `json:"synonymLangGroups"`
}

type SynonymLangGroup struct {
	Lang     string    `json:"lang"`
	Synonyms []Synonym `json:"synonyms"`
}

type Synonym struct {
	Words []SynonymWord `json:"words"`
}

type SynonymWord struct {
	WordValue string `json:"wordValue"`
	Lang      string `json:"lang"`
}

type PosInfo struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

type Paradigm struct {
	Title            string `json:"title"`
	InflectionTypeNr string `json:"inflectionTypeNr"`
	InflectionType   string `json:"inflectionType"`
	WordClass        string `json:"wordClass"`
	Forms            []Form `json:"paradigmForms"`
}

type Form struct {
	Value     string `json:"value"`
	MorphCode string `json:"morphCode"`
}
