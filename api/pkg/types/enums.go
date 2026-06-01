package types

type ExampleType int

const (
	ConfigTypePreview ExampleType = iota
	ConfigTypePublished
)

var AllExampleTypes = []struct {
	Value  ExampleType
	TSName string
}{
	{ConfigTypePreview, "preview"},
	{ConfigTypePublished, "published"},
}
