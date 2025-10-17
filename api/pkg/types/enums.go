package types

type ComicType int

const (
	ConfigTypePreview ComicType = iota
	ConfigTypePublished
)

var AllComicTypes = []struct {
	Value  ComicType
	TSName string
}{
	{ConfigTypePreview, "preview"},
	{ConfigTypePublished, "published"},
}
