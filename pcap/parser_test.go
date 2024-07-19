package pcap

import (
	"fmt"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	filePath := "testing/srt_capture_test.pcap"

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal("Failed to open file")
	}
	defer file.Close()

	pcapFile, err := decode(file)
	if err != nil {
		t.Fatalf("Failed to read PCAP file: %v", err)
	}

	if len(pcapFile.Packets) == 0 {
		t.Fatal("Should be at least one packet but there is 0")
	}

	fmt.Printf("PCAP Header: %+v\n", pcapFile.Header)

	for i, packet := range pcapFile.Packets {
		fmt.Printf("Packet %d: %+v\n", i, packet.Header)

		fmt.Printf("Data %d: %x\n", i, packet.Data)
	}
}
