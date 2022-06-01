package commands

type SearchResult struct {
	Name  string
	Score int
	Url   string
}

type TaggedSearchResult struct {
	Docset string
	Result SearchResult
}

type SearchResponse struct {
	Results map[string][]SearchResult
}
