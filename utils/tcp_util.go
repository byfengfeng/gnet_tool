package utils

func DecodeRound(bytes []byte) (data []byte,remainingByte []byte) {
	length := Length(bytes)
	data = bytes[2:]
	if len(bytes) > int(length) {
		remainingByte =  bytes[length:]
	}
	return
}

func Decode(bytes []byte) (code uint16,data []byte) {
	code = uint16(bytes[0]) << 8 | uint16(bytes[1])
	data = bytes[2:]
	return
}

func Encode(code uint16, data []byte) (bytes []byte) {
	length := uint32(len(data)+2+2)
	bytes = make([]byte,0)
	bytes = append(bytes,byte(length >> 8),byte(length))
	bytes = append(bytes,byte(code >> 8),byte(code))
	bytes = append(bytes,data...)
	return
}

func Length(bytes []byte) (length uint32) {
	length = uint32(bytes[0]) << 8 | uint32(bytes[1])
	return
}
