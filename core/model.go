package core

type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() string
}
