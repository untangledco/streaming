package pcap

import (
	"fmt"
	"os"
	"testing"
	"time"
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

	encodePcapFile, err := encode(pcapFile)

	if err != nil {
		t.Fatalf("encode file: %v", err)
	}

	fmt.Printf("Encoded PCAP: %x\n", encodePcapFile)

	f, err := os.Create("testing/pcap_file_encode.pcap")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(encodePcapFile)
	if err != nil {
		t.Fatalf("Failed to write to PCAP file")
	}

	filePathTwo := "testing/pcap_file_encode.pcap"

	fileTwo, err := os.Open(filePathTwo)
	if err != nil {
		t.Fatal("Failed to open file")
	}
	defer fileTwo.Close()

	pcap, err := decode(fileTwo)
	if err != nil {
		t.Fatalf("Failed to read PCAP file: %v", err)
	}

	if len(pcap.Packets) == 0 {
		t.Fatal("Should be at least one packet but there is 0")
	}

	fmt.Printf("PCAP Header: %+v\n", pcap.Header)

	for i, packet := range pcap.Packets {
		fmt.Printf("Packet %d: %+v\n", i, packet.Header)

		fmt.Printf("Data %d: %x\n", i, packet.Data)
	}

}

func TestTimestamp(t *testing.T) {
	want := [2]uint32{1, 100}
	when := time.Unix(1, 100)
	sec, nsec := timestamp(when)
	got := [2]uint32{sec, nsec}
	if got != want {
		t.Errorf("timestamp(%s) = %v, want %v", when, got, want)
	}
}
