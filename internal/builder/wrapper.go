package builder

// Builder 建筑引擎包装器。
type Builder struct{}

// NewBuilder 创建建筑引擎。
func NewBuilder() *Builder {
	return &Builder{}
}

// Build 调用 Build 函数。
func (b *Builder) Build(req BuildRequest) (*BuildResponse, error) {
	return Build(req)
}
