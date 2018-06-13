package codec

//the coder
type Coder interface {
	//encode v to []byte, err != nil if not succeed
	Encode(v interface{}) (data []byte, err error)

	// decode []byte to Msg, err != nil if not succeed
	// v must be a pointer
	Decode(data []byte, v interface{}) (err error)

	// name of the coder
	TypeName() string
}
