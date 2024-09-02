package dns

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type DNSMessage struct {
	Header Header;
	Question []Question;
	Answer []Answer;
	Authority Authority;
}


type Header struct {
    ID               uint16 // ID for the DNS query
    QR               uint8  // Query/Response flag (0 for query, 1 for response)
    OpCode           uint8  // Operation code (e.g., standard query = 0)
    AA               uint8  // Authoritative Answer flag
    TC               uint8  // Truncation flag
    RD               uint8  // Recursion Desired flag
    RA               uint8  // Recursion Available flag
    Z                uint8  // Reserved for future use (should be 0)
    RCode            uint8  // Response code (0 for no error)
    QDCount          uint16 // Number of questions in the Question section
    ANCount          uint16 // Number of resource records in the Answer section
    NSCount          uint16 // Number of name server resource records in the Authority Records section
    ARCount          uint16 // Number of resource records in the Additional Records section
}
type Question struct{
	Question string;
	Class int;
	Type int;

}
type Answer struct {
	Domain string;
	Class int;
	Type int;
	TTL int;
	Len int;
	Data string;
}
type Authority struct{

}

func (msg *DNSMessage) ParseMsg() []byte {
	//Header
	buf := make([]byte, 12);
    binary.BigEndian.PutUint16(buf[0:2], msg.Header.ID)
    buf[2] = msg.Header.QR<<7 | (msg.Header.OpCode&0xF)<<3 | msg.Header.AA<<2 | msg.Header.TC<<1 | msg.Header.RD
    buf[3] = msg.Header.RA<<7 | msg.Header.Z<<4 | (msg.Header.RCode & 0xF)
    binary.BigEndian.PutUint16(buf[4:6], msg.Header.QDCount)
    binary.BigEndian.PutUint16(buf[6:8], msg.Header.ANCount)
    binary.BigEndian.PutUint16(buf[8:10], msg.Header.NSCount)
    binary.BigEndian.PutUint16(buf[10:12], msg.Header.ARCount)

	//Question
	// fmt.Println("-----------")
	// fmt.Println(msg.Question)
	for i:=0;i<len(msg.Question);i++{
		stringEncoding := encodeString(msg.Question[i].Question);
		buf = append(buf, stringEncoding...);
		byteArray := make([]byte, 2) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Question[i].Type))
		buf = append(buf, byteArray...)
		byteArray = make([]byte, 2) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Question[i].Class))
		buf = append(buf, byteArray...)
	}

	//Answer
	// fmt.Println("-----------")
	// fmt.Println(msg.Answer)
	for i:=0;i<len(msg.Answer);i++{
		stringEncoding := encodeString(msg.Answer[i].Domain);
		buf = append(buf, stringEncoding...);
		byteArray := make([]byte, 2) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer[i].Type))
		buf = append(buf, byteArray...)
		byteArray = make([]byte, 2) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer[i].Class))
		buf = append(buf, byteArray...)
		byteArray = make([]byte, 4) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer[i].TTL))
		buf = append(buf, byteArray...)
		byteArray = make([]byte, 2) 
		binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer[i].Len))
		buf = append(buf, byteArray...)
		buf = append(buf, encodeData(msg.Answer[i].Data)...)
	}
    return buf
}


func encodeString(domain string) []byte{
	names := strings.Split(domain, ".");
	var res []byte;
	for i:=0;i<len(names);i++{
		val := uint8(len(names[i]));
		res = append(res, val);
		for j:=0;j<len(names[i]);j++{
			res = append(res, byte(names[i][j]));
		}
	}
	res = append(res, byte('\x00'));
	
	return res;
}

func encodeData(domain string) []byte{
	names := strings.Split(domain, ".");
	var res []byte;
	for i:=0;i<len(names);i++{
		for j:=0;j<len(names[i]);j++{
			res = append(res, byte(names[i][j]));
		}
	}
	return res;
}

func parseDNSHeader(data []byte) Header {
    return Header{
        ID:      binary.BigEndian.Uint16(data[0:2]),
        QR:      data[2] >> 7,
        OpCode:  (data[2] >> 3) & 0xF,
        AA:      (data[2] >> 2) & 0x1,
        TC:      (data[2] >> 1) & 0x1,
        RD:      data[2] & 0x1,
        RA:      data[3] >> 7,
        Z:       (data[3] >> 4) & 0x7,
        RCode:   data[3] & 0xF,
        QDCount: binary.BigEndian.Uint16(data[4:6]),
        ANCount: binary.BigEndian.Uint16(data[6:8]),
        NSCount: binary.BigEndian.Uint16(data[8:10]),
        ARCount: binary.BigEndian.Uint16(data[10:12]),
    }
}

func parseDNSQuestion(data []byte, offset int) (Question, int) {
	if(data[offset]>>6 == 3){
		newOffSet := binary.BigEndian.Uint16((data[offset:offset+2])) ^ 49152;
		return parseDNSQuestion(data, int(newOffSet));
	}
    qName, newOffset := parseQName(data, offset)
    qType := int(binary.BigEndian.Uint16(data[newOffset : newOffset+2]))
    qClass := int(binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4]))
    return Question{Question: qName, Type: qType, Class: qClass}, newOffset + 4
}

func parseQName(data []byte, offset int) (string, int) {
    var qName string
    for {
        length := int(data[offset])
        if length == 0 {
            break
        }
        offset++
        qName += string(data[offset:offset+length]) + "."
        offset += length
    }
    return qName[:len(qName)-1], offset + 1
}

func parseDNSAnswer(data []byte, offset int) (Answer, int) {
    name, newOffset := parseQName(data, offset)
    ansType := int(binary.BigEndian.Uint16(data[newOffset : newOffset+2]))
    ansClass := int(binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4]))
    ttl := int(binary.BigEndian.Uint32(data[newOffset+4 : newOffset+8]))
    rdLength := int(binary.BigEndian.Uint16(data[newOffset+8 : newOffset+10]))
    rData := string(data[newOffset+10 : newOffset+10+int(rdLength)])
	ipAddress := fmt.Sprintf("%d.%d.%d.%d", rData[0], rData[1], rData[2], rData[3])
	fmt.Println(ipAddress)
    return Answer{
        Domain:     name,
        Type:     ansType,
        Class:    ansClass,
        TTL:      ttl,
        Len: rdLength,
        Data:    ipAddress,
    }, newOffset + 10 + int(rdLength)
}

func ParseDNSMessage(data []byte) DNSMessage {
    header := parseDNSHeader(data)
    offset := 12

    // Parse questions
    questions := make([]Question, header.QDCount)
    for i := 0; i < int(header.QDCount); i++ {
		questions[i], offset = parseDNSQuestion(data, offset)
    }

    // Parse answers
    answers := make([]Answer, header.ANCount)
    for i := 0; i < int(header.ANCount); i++ {
        answers[i], offset = parseDNSAnswer(data, offset)

    }

    return DNSMessage{
        Header:    header,
        Question: questions,
        Answer:   answers,
    }
}