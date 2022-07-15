package main

func compress(buf []byte, w Writer, ver uint8, compression int) {
	// *compression*
	for i := 0; i < len(buf); i++ {
		w.writeByte(buf[i])
	}
}
