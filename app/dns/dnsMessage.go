package dns

import (
	"encoding/binary"
	"strings"
)

type DNSMessage struct {
	Header Header;
	Question Question;
	Answer Answer;
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
	stringEncoding := encodeString(msg.Question.Question);
	buf = append(buf, stringEncoding...);
	byteArray := make([]byte, 2) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Question.Type))
	buf = append(buf, byteArray...)
	byteArray = make([]byte, 2) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Question.Class))
	buf = append(buf, byteArray...)

	//Answer
	stringEncoding = encodeString(msg.Answer.Domain);
	buf = append(buf, stringEncoding...);
	byteArray = make([]byte, 2) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer.Type))
	buf = append(buf, byteArray...)
	byteArray = make([]byte, 2) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer.Class))
	buf = append(buf, byteArray...)
	byteArray = make([]byte, 4) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer.TTL))
	buf = append(buf, byteArray...)
	byteArray = make([]byte, 2) 
	binary.BigEndian.PutUint16(byteArray, uint16(msg.Answer.Len))
	buf = append(buf, byteArray...)
	buf = append(buf, encodeData(msg.Answer.Data)...)
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
