package actions

type Compile struct{}

type Load struct{}

type SplitChange struct {
	Sizes []float64
}

type EditorTextChangedDebounced struct {
	Text string
}
