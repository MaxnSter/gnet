package codec

type Coder interface {

	//encode msg to []byte, err != nil if not succeed
	Encode(v interface{}) (data []byte, err error)

	// decode []byte to msg, err != nil if not succeed
	// v must be a pointer
	Decode(data []byte, v interface{}) (err error)

	// name of the coder
	TypeName() string
}
