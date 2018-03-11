package actions

type CompileStart struct{}
type CompileOpen struct{}
type CompileMessage struct{ Message interface{} }
type CompileClose struct{}

type Load struct{}

type SplitChange struct {
	Sizes []float64
}

type EditorTextChangedDebounced struct {
	Text string
}

type Error struct {
	Err error
}
