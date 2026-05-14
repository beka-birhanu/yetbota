package processors

// PostFanOutData holds the data needed to start a fan-out for a post.
type PostFanOutData struct {
	Score    float64
	AuthorID string
}
