package jxs

const (
	codestreamStart uint16 = 0xff10
	codestreamEnd   uint16 = 0xff11
)

type codestream struct {
	start        uint16
	capabilities capabilities
}

type capabilities uint8

const (
	starTetrixTransform capabilities = 1<<1 + iota
	quadraticTransform
	extendedTransform
	waveletDecomposition
	losslessDecode
	rawMode
)

type header struct {
	marker uint16
	length uint16
}
