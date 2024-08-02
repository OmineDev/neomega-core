package underlay_conn

import (
	"bytes"
	"fmt"
)

func byteSlicesToBytes(packets [][]byte) []byte {
	buf := bytes.NewBuffer(nil)
	l := make([]byte, 5)
	for _, packet := range packets {
		// Each packet is prefixed with a varuint32 specifying the length of the packet.
		writeVaruint32(buf, uint32(len(packet)), l)
		buf.Write(packet)
	}
	return buf.Bytes()
}

func bytesToBytesSlices(data []byte) (packets [][]byte) {
	packets = make([][]byte, 0)
	b := bytes.NewBuffer(data)
	for b.Len() != 0 {
		var length uint32
		if err := readVaruint32(b, &length); err != nil {
			panic(fmt.Errorf("error reading packet length: %v", err))
		}
		packets = append(packets, b.Next(int(length)))
	}
	return packets
}
