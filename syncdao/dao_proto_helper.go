package syncdao

// TODO(doug4j@gmail.com): Move the contents in this go file to the syncmsg package.

import (
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"encoding/hex"
	//"errors"
	"github.com/golang/protobuf/proto"
	//"log"
	//"time"
)

// TODO(doug4j@gmail.com): All of these Calculate* methods can be put into a creator class that reports errors on errors separately similar to syncmsg.NewSyncMsgCreator

//CalculateStringHexOfBinValue creates a hex value from the binary contents of syncmsg.ProtoFieldTypeBytes returning an empty string and log.Println an error should the proto.Unmarshal fail.
func CalculateStringHexOfBinValue(fieldValue []byte) string {
	answer := &syncmsg.ProtoFieldTypeBytes{}
	err := proto.Unmarshal(fieldValue, answer)
	if err != nil {
		// BUG(doug4j@gmail.com): Finish my error handling. This is not a good long-term solution... just hacked it in for now.
		syncutil.Error("Error: " + err.Error() + ". Returning empty string")
		return ""
	}
	return hex.EncodeToString(answer.FieldValue)
}

//CalculateInt64Value creates an int64 from syncmsg.ProtoFieldTypeSint64 returning 0 and log.Println an error should the proto.Unmarshal fail.
/*
func CalculateInt64Value(fieldValue []byte) int64 {
	answer := &syncmsg.ProtoFieldTypeSint64{}
	err := proto.Unmarshal(fieldValue, answer)
	if err != nil {
		// BUG(doug4j@gmail.com): Finish my error handling. This is not a good long-term solution... just hacked it in for now.
		syncutil.Error("Error: " + err.Error() + ". Returning 0")
		return 0
	}
	return *answer.FieldValue
}
*/

//CalculateValue creates an string, sint64, or double based on valueType (syncmsg.ProtoEncodedFieldType) from supplied bytes (fieldValue []byte). Returns nil and log.Println on error. Returns nil and log.Println msg if valueType (syncmsg.ProtoEncodedFieldType) is not syncmsg.ProtoEncodedFieldType_STRING, ProtoEncodedFieldType_SINT64, or ProtoEncodedFieldType_DOUBLE.
func CalculateValue(valueType syncmsg.ProtoEncodedFieldType, fieldValue []byte) interface{} {
	switch valueType {
	case syncmsg.ProtoEncodedFieldType_STRING:
		answer := &syncmsg.ProtoFieldTypeString{}
		err := proto.Unmarshal(fieldValue, answer)
		if err != nil {
			syncutil.Error("Error: " + err.Error())
			return nil
		}
		return *answer.FieldValue
	case syncmsg.ProtoEncodedFieldType_SINT64:
		answer := &syncmsg.ProtoFieldTypeSint64{}
		err := proto.Unmarshal(fieldValue, answer)
		if err != nil {
			syncutil.Error("Error: " + err.Error())
			return nil
		}
		return *answer.FieldValue
	case syncmsg.ProtoEncodedFieldType_DOUBLE:
		answer := &syncmsg.ProtoFieldTypeDouble{}
		err := proto.Unmarshal(fieldValue, answer)
		if err != nil {
			syncutil.Error("Error: " + err.Error())
			return nil
		}
		return *answer.FieldValue
	// TODO(doug4j@gmail.com): Finish value extraction implementation for all types
	case syncmsg.ProtoEncodedFieldType_FLOAT:
		answer := &syncmsg.ProtoFieldTypeFloat{}
		err := proto.Unmarshal(fieldValue, answer)
		if err != nil {
			syncutil.Error("Error: " + err.Error())
			return nil
		}
		return *answer.FieldValue
	default:
		syncutil.Error("Error: supplied type is not ProtoEncodedFieldType_STRING, ProtoEncodedFieldType_SINT64, ProtoEncodedFieldType_DOUBLE, or ProtoEncodedFieldType_FLOAT")
		return nil
	}
}

//SyncFieldDefinition defines a field definition for syncing
type SyncFieldDefinition struct {
	FieldName    string
	FieldType    SyncFieldTypeEnum
	IsPrimaryKey bool
}

//SyncFieldTypeEnum represents the types of available sync fields
type SyncFieldTypeEnum int32

const (
	//SyncFieldTypeEnumUndefined is an undefined SyncFieldTypeEnum.
	SyncFieldTypeEnumUndefined SyncFieldTypeEnum = 0
	//SyncFieldTypeEnumString is a string SyncFieldTypeEnum.
	SyncFieldTypeEnumString SyncFieldTypeEnum = 1
	//SyncFieldTypeEnumInt is an int SyncFieldTypeEnum.
	SyncFieldTypeEnumInt SyncFieldTypeEnum = 2
	//SyncFieldTypeEnumFloat is a float SyncFieldTypeEnum.
	SyncFieldTypeEnumFloat SyncFieldTypeEnum = 3
	//SyncFieldTypeEnumBool is a bool SyncFieldTypeEnum.
	SyncFieldTypeEnumBool SyncFieldTypeEnum = 4
	//SyncFieldTypeEnumDate is an time SyncFieldTypeEnum.
	SyncFieldTypeEnumDate SyncFieldTypeEnum = 5
	//SyncFieldTypeEnumBinary is an []byte SyncFieldTypeEnum.
	SyncFieldTypeEnumBinary SyncFieldTypeEnum = 6
)

//SyncFieldTypeEnumName gives the string name of SyncFieldTypeEnum.
var SyncFieldTypeEnumName = map[int32]string{
	0: "Undefined",
	1: "String",
	2: "Int",
	3: "Float",
	4: "Bool",
	5: "Date",
	6: "Binary",
}

//SyncFieldTypeEnumValue gives the int32 value of SyncFieldTypeEnum.
var SyncFieldTypeEnumValue = map[string]int32{
	"Undefined": 0,
	"String":    1,
	"Int":       2,
	"Float":     3,
	"Bool":      4,
	"Date":      5,
	"Binary":    6,
}

//Enum gives the SyncFieldTypeEnum given one of the constants from SyncFieldTypeEnum*.
func (x SyncFieldTypeEnum) Enum() *SyncFieldTypeEnum {
	p := new(SyncFieldTypeEnum)
	*p = x
	return p
}
