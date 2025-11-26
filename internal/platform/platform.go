package platform

type Window interface {
	Show()
	Close()

	SetTitle(title string)
	SetSize(width, height int)

	OnClose(callback func())
	OnMouseMove(callback func(x, y int))
	OnClick(callback func(x, y int, button MouseButton))

	Run()
}
