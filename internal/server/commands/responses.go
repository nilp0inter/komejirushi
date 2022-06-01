package commands

type SearchResult struct {
	Name  string `json:"n"`
	Score int    `json:"s"`
	Url   string `json:"u"`
}

type TaggedSearchResult struct {
	Docset string       `json:"ds"`
	Result SearchResult `json:"rs"`
}

type SearchResponse struct {
	Results map[string][]SearchResult `json:"results"`
}
