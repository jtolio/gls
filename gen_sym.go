package context

var (
	symPool = &IdPool{}
)

type ContextKey struct{ id uint }

func GenSym() ContextKey {
	return ContextKey{id: symPool.Acquire()}
}
