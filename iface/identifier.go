package iface

// Identifier表示一个有唯一标识的对象
type Identifier interface {
	// ID返回当前对象的唯一标识
	ID() int64
}
