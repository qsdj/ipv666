package addressing

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lavalamp-/ipv666/internal"
	"github.com/lavalamp-/ipv666/internal/logging"
	"github.com/lavalamp-/ipv666/internal/zrandom"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

func FilterIPv4FromList(toParse []*net.IP) ([]*net.IP) {
	var toReturn []*net.IP
	for _, curIP := range toParse {
		if curIP.To4() == nil {
			toReturn = append(toReturn, curIP)
		}
	}
	return toReturn
}

func IsAddressIPv4(toCheck *net.IP) (bool) {
	return toCheck.To4() != nil
}

func GetIPsFromStrings(toParse []string) []*net.IP {
	var toReturn []*net.IP
	for _, curParse := range toParse {
		newIP := net.ParseIP(curParse)
		if newIP == nil {
			logging.Warnf("Could not parse IP from string '%s'.", curParse)
		} else {
			toReturn = append(toReturn, &newIP)
		}
	}
	return toReturn
}

func GetIPSet(ips []*net.IP) map[string]*internal.Empty {
	toReturn := make(map[string]*internal.Empty)
	blacklistEntry := &internal.Empty{}
	for _, ip := range ips {
		toReturn[ip.String()] = blacklistEntry
	}
	return toReturn
}

func GetFirst64BitsOfIP(ip *net.IP) uint64 {
	ipBytes := ([]byte)(*ip)
	return binary.LittleEndian.Uint64(ipBytes[:8])
}

func GetUniqueIPs(ips []*net.IP, updateFreq int) []*net.IP {
	checkMap := make(map[string]bool)
	var toReturn []*net.IP
	for i, ip := range ips {
		if i % updateFreq == 0 {
			logging.Debugf("Processing %d out of %d for unique IPs.", i, len(ips))
		}
		if _, ok := checkMap[ip.String()]; !ok {
			checkMap[ip.String()] = true
			toReturn = append(toReturn, ip)
		}
	}
	return toReturn
}

func FatHexStringToIP(toParse string) (*net.IP, error) {
	data, err := hex.DecodeString(toParse)
	if err != nil {
		return nil, err
	}
	ip := net.IP(data)
	return &ip, nil
}

func ReadIPsFromFatHexFileBytes(toParse []byte) []*net.IP {
	parseString := strings.TrimSpace(string(toParse))
	lines := strings.Split(parseString, "\n")
	var toReturn []*net.IP
	for _, line := range lines {
		lineStrip := strings.TrimSpace(line)
		newIp, err := FatHexStringToIP(lineStrip)
		if err != nil {
			logging.Warnf("Error thrown when processing bytes %s as fat hex: %s", line, err.Error())
		} else {
			toReturn = append(toReturn, newIp)
		}
	}
	return toReturn
}

func ReadIPsFromHexFileBytes(toParse []byte) []*net.IP {
	parseString := strings.TrimSpace(string(toParse))
	lines := strings.Split(parseString, "\n")
	var toReturn []*net.IP
	for _, line := range lines {
		newIP := net.ParseIP(strings.TrimSpace(line))
		if newIP == nil {
			logging.Warnf("No IP found from content '%s'.", line)
			continue
		}
		toReturn = append(toReturn, &newIP)
	}
	return toReturn
}

func ReadIPsFromHexFile(filePath string) ([]*net.IP, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ReadIPsFromHexFileBytes(fileContent), nil
}

func WriteIPsToHexFile(filePath string, addrs []*net.IP) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	writer := bufio.NewWriter(file)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, addr := range addrs {
		writer.WriteString(fmt.Sprintf("%s\n", addr.String()))
	}
	writer.Flush()
	return nil
}

func GetTextLinesFromIPs(addrs []*net.IP) (string) {
	var toReturn []string
	for _, addr := range addrs {
		toReturn = append(toReturn, fmt.Sprintf("%s\n", addr.String()))
	}
	return strings.Join(toReturn, "")
}

func ReadIPsFromBinaryFileBytes(toParse []byte) []*net.IP {
	var toReturn []*net.IP
	for i := 0; i < len(toParse); i += 16 {
		ipBytes := make([]byte, 16)
		copy(ipBytes, toParse[i:i+16])
		newIP := net.IP(ipBytes)
		toReturn = append(toReturn, &newIP)
	}
	return toReturn
}

func ReadIPsFromBinaryFile(filePath string) ([]*net.IP, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	if fileSize % 16 != 0 {
		return nil, errors.New(fmt.Sprintf("Expected file size to be a multiple of 16 (got %d).", fileSize))
	}
	buffer := make([]byte, 16)
	var toReturn []*net.IP
	for {
		_, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		ipBytes := make([]byte, 16)
		copy(ipBytes, buffer)
		newIP := net.IP(ipBytes)
		toReturn = append(toReturn, &newIP)
	}
	return toReturn, nil
}

func WriteIPsToBinaryFile(filePath string, addrs []*net.IP) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	writer := bufio.NewWriter(file)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, addr := range addrs {
		writer.Write(*addr)
	}
	writer.Flush()
	return nil
}

func WriteIPsToFatHexFile(filePath string, addrs []*net.IP) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	writer := bufio.NewWriter(file)
	if err != nil {
		return err
	}
	defer file.Close()
	buffer := make([]byte, 32)
	for _, addr := range addrs {
		hex.Encode(buffer, *addr)
		writer.Write(buffer)
		writer.Write([]byte("\n"))
	}
	writer.Flush()
	return nil
}

func ReadIPsFromFile(filePath string) ([]*net.IP, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParseIPsFromBytes(bytes)
}

func ParseIPsFromBytes(toParse []byte) ([]*net.IP, error) {
	split := strings.Split(string(toParse), "\n")
	toCheck := split[0]
	if strings.Contains(toCheck, ":") {  // Standard ASCII hex with colons
		return ReadIPsFromHexFileBytes(toParse), nil
	} else if len(toCheck) == 32 {  // ASCII hex without colons
		return ReadIPsFromFatHexFileBytes(toParse), nil
	} else if len(toParse) % 16 != 0 {
		return nil, errors.New("could not determine the format of IPv6 address bytes")
	} else {  // Binary representation
		return ReadIPsFromBinaryFileBytes(toParse), nil
	}
}

func GetNybbleFromIP(ip *net.IP, index int) (uint8) {
	// TODO fatal error if index > 31
	byteIndex := index / 2
	addrBytes := ([]byte)(*ip)
	addrByte := addrBytes[byteIndex]
	if index % 2 == 0 {
		return addrByte >> 4
	} else {
		return addrByte & 0xf
	}
}

func GetNybblesFromIP(ip *net.IP, nybbleCount int) ([]uint8) {
	var toReturn []uint8
	for i := 0; i < nybbleCount; i++ {
		toReturn = append(toReturn, GetNybbleFromIP(ip, i))
	}
	return toReturn
}

func GenerateRandomAddress() (*net.IP) {
	bytes := zrandom.GenerateHostBits(128)
	toReturn := net.IP(bytes)
	return &toReturn
}

func FlipBitsInAddress(toFlip *net.IP, startIndex uint8, endIndex uint8) (*net.IP) {
	toFlipBytes := *toFlip
	endIndex++
	startByte := startIndex / 8
	startOffset := startIndex % 8
	endByte := endIndex / 8
	endOffset := endIndex % 8
	var maskBytes []byte
	var flipBytes []byte
	var i uint8

	if startByte == endByte {
		for i = 0; i < 16; i++ {
			if i == startByte {
				firstHalf := byte(^(0xff >> startOffset))
				secondHalf := byte(0xff >> endOffset)
				maskBytes = append(maskBytes, firstHalf | secondHalf)
			} else {
				maskBytes = append(maskBytes, 0xff)
			}
		}
	} else {
		for i = 0; i < 16; i++ {
			if i < startByte {
				maskBytes = append(maskBytes, 0xff)
			} else if i == startByte {
				maskBytes = append(maskBytes, byte(^(0xff >> startOffset)))
			} else if i < endByte {
				maskBytes = append(maskBytes, 0x00)
			} else if i == endByte {
				maskBytes = append(maskBytes, byte(0xff >> endOffset))
			} else {
				maskBytes = append(maskBytes, 0xff)
			}
		}
	}

	for i = 0; i < 16; i++ {
		flippedBits := ^toFlipBytes[i] & ^maskBytes[i]
		flipBytes = append(flipBytes, toFlipBytes[i] & maskBytes[i] | flippedBits)
	}

	toReturn := net.IP(flipBytes)
	return &toReturn

}
