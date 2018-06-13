package codec

var (
	coders = map[string]Coder{}
)

func RegisterCoder(name string, c Coder) {
	if _, ok := coders[name]; ok {
		panic("duplicate register Coder :" + name)
	}

	coders[name] = c
}

func MustGetCoder(name string) Coder {
	if c, ok := coders[name]; ok {
		return c
	} else {
		panic("Coder not register, name :" + name)
	}
}
