package syncmsg

// BUG(doug4j@gmail.com): Move this test file into a separate package, see https://medium.com/@matryer/5-simple-tips-and-tricks-for-writing-unit-tests-in-golang-619653f90742#.o8nxf7z53
import (
	"crypto/sha256"
	//"datasynctools/syncutil"
	//"datasynctools/syncdao"
	"encoding/hex"
	//"fmt"
	"log"
	"testing"

	proto "github.com/golang/protobuf/proto"
	//"time"
)

func unwrapProtoBytesString(fieldValue []byte) string {
	answer := &ProtoFieldTypeString{}
	err := proto.Unmarshal(fieldValue, answer)
	if err != nil {
		//syncutil.Error("Error: " + err.Error())
		return "<not type ProtoFieldTypeString>"
	}
	return *answer.FieldValue
}

func TestSimpleHash(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestSimpleHash")
	log.Println("")

	creator := NewCreator()

	record := &ProtoRecord{
		Fields: []*ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateBoolProtoField("mandatoryBool", true),
			creator.CreateTimeProtoField("mandatoryDate", creator.FormatTimeFromString("2009-10-02 16:05:03.230")),
			creator.CreateDoubleProtoField("mandatoryDouble", 203.2435),
			creator.CreateFloatProtoField("mandatoryFloat", 432.50),
			creator.CreateInt64ProtoField("mandatoryInt", 101),
			creator.CreateStringProtoField("mandatoryString", "Hello"),
		},
	}

	for _, field := range record.GetFields() {
		log.Printf("name:%s, type:%s, bytes:%v, valueAsString:%v", field.GetFieldName(), field.GetEncodedFieldType(), field.GetFieldValue(), unwrapProtoBytesString(field.GetFieldValue()))
	}

	recordBytes, err := proto.Marshal(record)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return
		//return err
	}

	recordSha256Hex := Hash256Bytes(recordBytes)
	log.Println("recordSha256Hex:", recordSha256Hex)

	expectedHex := "62249f849e11b39ec5cfaef37e5ba03b106f0d0a76285086ca0519f6310d637d"
	if recordSha256Hex != expectedHex {
		t.Errorf("hex value is not what is expected, should be '%s' and it's '%s'", expectedHex, recordSha256Hex)
	}

	log.Println("")
	log.Println("***END TestSimpleHash")
	log.Println("")
	log.Println("")
}

func TestDateHash(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestDateHash")
	log.Println("")

	creator := NewCreator()

	record := &ProtoRecord{
		Fields: []*ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateTimeProtoField("mandatoryDate", creator.FormatTimeFromString("2009-10-02 16:05:03.230")),
		},
	}

	for _, field := range record.GetFields() {
		log.Printf("name:%s, type:%s, bytes:%v, valueAsString:%v", field.GetFieldName(), field.GetEncodedFieldType(), field.GetFieldValue(), unwrapProtoBytesString(field.GetFieldValue()))
	}

	recordBytes, err := proto.Marshal(record)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return
		//return err
	}

	recordSha256Hex := Hash256Bytes(recordBytes)
	log.Println("recordSha256Hex:", recordSha256Hex)

	expectedHex := "740edf09ec644a2e62256b286fe465c0782c43b326e3330c9eb5a757483b938f"
	if recordSha256Hex != expectedHex {
		t.Errorf("hex value is not what is expected, should be '%s' and it's '%s'", expectedHex, recordSha256Hex)
	}

	log.Println("")
	log.Println("***END TestDateHash")
	log.Println("")
	log.Println("")
}

func TestContactHash(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestContactHash")
	log.Println("")

	creator := NewCreator()

	record := &ProtoRecord{
		Fields: []*ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", "911DD745-8916-41C4-9973-F8B38A501602"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Henry"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 5.5),
			creator.CreateStringProtoField("lastName", "Adins"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	recordLastKnownBytes, err := proto.Marshal(record)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		t.Error(err)
	}
	actual := Hash256Bytes(recordLastKnownBytes)
	log.Println("Hash256Bytes: ", Hash256Bytes(recordLastKnownBytes))

	expected := "75648c9b32dba111fc51ffd533e04b0f2555b20ce492ec558327e495bf326bc2"
	if actual != expected {
		t.Error("Value is not as expected")
	}

	log.Println("")
	log.Println("***END TestContactHash")
	log.Println("")
	log.Println("")
}

type AllTypes struct {
	mandatoryString string
	/*
	   let mandatoryInt: Int
	   let mandatoryDate: NSDate
	   let mandatoryBool: Bool
	   let mandatoryDouble: Double
	   let mandatoryFloat: Float
	   let mandatoryBinary: NSData

	   let optionalButPresentString: String?
	   let optionalButPresentInt: Int?
	   let optionalButPresentDate: NSDate?
	   let optionalButPresentBool: Bool?
	   let optionalButPresentDouble: Double?
	   let optionalButPresentFloat: Float?
	   let optionalButPresentBinary: NSData?

	   let optionalNilString: String?
	   let optionalNilInt: Int?
	   let optionalNilDate: NSDate?
	   let optionalNilBool: Bool?
	   let optionalNilDouble: Double?
	   let optionalNilFloat: Float?
	   let optionalNilBinary: NSData?
	*/
}

func Hash256Bytes(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)
	msgDigest := hasher.Sum(nil)
	sha256Hex := hex.EncodeToString(msgDigest)
	return sha256Hex
}
