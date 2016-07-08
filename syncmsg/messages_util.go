package syncmsg

import (
	"data-sync-tools-go/syncutil"
	"errors"
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
)

//Value returns the undelrying int32 value of SentSyncStateEnum
func (sentState SentSyncStateEnum) Value() int32 {
	return int32(sentState)
	//return proto.enum .EnumName(SentSyncStateEnum_name, int32(x))
}

//IsFirstTimeSend determines if the value is in the first time state
func (sentState SentSyncStateEnum) IsFirstTimeSend() bool {
	return sentState == SentSyncStateEnum_PersistedNeverSentToPeer
	//return proto.enum .EnumName(SentSyncStateEnum_name, int32(x))
}

//IsNotFirstTimeSend determines if the value is not in the first time state
func (sentState SentSyncStateEnum) IsNotFirstTimeSend() bool {
	return sentState == SentSyncStateEnum_PersistedStandardSentToPeer
	//return proto.enum .EnumName(SentSyncStateEnum_name, int32(x))
}

//NewCreatorError creates an error for the list of errors from the Creator.
func NewCreatorError(errs []error) error {
	if len(errs) <= 0 {
		msg := "Cannot create common CreatorError because there are no errors. Returing this message instead."
		syncutil.Error(msg)
		return errors.New(msg)
	}
	return fmt.Errorf("%s. %v other errors (not shown)", errs[0].Error(), len(errs)-1)
}

//Creator represents a means for creating sync record fields collecting error along the way and able to report on them after numerous successive calls.
type Creator struct {
	Errors []error
}

//NewCreator creates a new instance of SyncMsgCreater.
func NewCreator() *Creator {
	return &Creator{
		Errors: make([]error, 0, 5),
	}
}

func createPlaceholderProtoField(name string) *ProtoField {
	return &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_STRING).Enum(),
		FieldName:        proto.String(name + "-placeholder-"),
		FieldValue:       nil,
	}
}

//CreateBoolProtoField creates a sync record field of type bool.
func (c *Creator) CreateBoolProtoField(name string, value bool) *ProtoField {
	field := &ProtoFieldTypeBool{
		FieldValue: proto.Bool(value),
	}
	fieldBytes, err := proto.Marshal(field)
	if err != nil {
		c.Errors = append(c.Errors, err)
		return createPlaceholderProtoField(name)
	}
	answer := &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_BOOL).Enum(),
		FieldName:        proto.String(name),
		FieldValue:       fieldBytes,
	}
	return answer
}

//CreateStringProtoField creates a sync record field of type string.
func (c *Creator) CreateStringProtoField(name string, value string) *ProtoField {
	field := &ProtoFieldTypeString{
		FieldValue: proto.String(value),
	}
	fieldBytes, err := proto.Marshal(field)
	if err != nil {
		c.Errors = append(c.Errors, err)
		return createPlaceholderProtoField(name)
	}
	answer := &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_STRING).Enum(),
		FieldName:        proto.String(name),
		FieldValue:       fieldBytes,
	}
	return answer
}

//CreateInt64ProtoField creates a sync record field of type int
func (c *Creator) CreateInt64ProtoField(name string, value int) *ProtoField {
	field := &ProtoFieldTypeSint64{
		FieldValue: proto.Int64(int64(value)),
	}
	fieldBytes, err := proto.Marshal(field)
	if err != nil {
		c.Errors = append(c.Errors, err)
		return createPlaceholderProtoField(name)
	}
	answer := &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_SINT64).Enum(),
		FieldName:        proto.String(name),
		FieldValue:       fieldBytes,
	}
	return answer
}

//CreateDoubleProtoField creates a sync record field of type double.
func (c *Creator) CreateDoubleProtoField(name string, value float64) *ProtoField {
	field := &ProtoFieldTypeDouble{
		FieldValue: proto.Float64(float64(value)),
	}
	fieldBytes, err := proto.Marshal(field)
	if err != nil {
		c.Errors = append(c.Errors, err)
		return createPlaceholderProtoField(name)
	}
	answer := &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_DOUBLE).Enum(),
		FieldName:        proto.String(name),
		FieldValue:       fieldBytes,
	}
	return answer
}

//CreateFloatProtoField creates a sync record field of type float.
func (c *Creator) CreateFloatProtoField(name string, value float32) *ProtoField {
	field := &ProtoFieldTypeFloat{
		FieldValue: proto.Float32(float32(value)),
	}
	fieldBytes, err := proto.Marshal(field)
	if err != nil {
		c.Errors = append(c.Errors, err)
		return createPlaceholderProtoField(name)
	}
	answer := &ProtoField{
		EncodedFieldType: ProtoEncodedFieldType(ProtoEncodedFieldType_FLOAT).Enum(),
		FieldName:        proto.String(name),
		FieldValue:       fieldBytes,
	}
	return answer
}

//CreateTimeProtoField creates a sync record field of type time.Time.
func (c *Creator) CreateTimeProtoField(name string, utcTime time.Time) *ProtoField {
	//utcTime := value.UTC()
	stringValue := utcTime.Format(SyncStandardDateFormat)
	return c.CreateStringProtoField(name, stringValue)
}

//SyncStandardDateFormat is the standard date format used for all sync of date data.
var SyncStandardDateFormat = "2006-01-02 15:04:05.000"

//FormatTimeFromString is used to create time.Time using a unified string format for all sync record fields of type time.Time.
//The format is up to millisecond precision in following pattern by example: '2006-01-02 15:04:05.001'.
//A non-patterned version of sample date is Jan 2nd 2006 at 3:04pm 5 seconds and 1 millisecond.
func (c *Creator) FormatTimeFromString(dateString string) time.Time {
	//syncutil.Debug(dateString)
	answer, err := time.Parse(SyncStandardDateFormat, dateString)
	if err != nil {
		syncutil.Warn("Error creating date from string", dateString, "error:", err)
		c.Errors = append(c.Errors, err)
		return time.Now()
	}
	return answer
}

//FormatStringFromTime is used to create time.Time using a unified string format for all sync record fields of type time.Time.
//The format is up to millisecond precision in following pattern by example: '2006-01-02 15:04:05.001'.
//A non-patterned version of sample date is Jan 2nd 2006 at 3:04pm 5 seconds and 1 millisecond.
func (c *Creator) FormatStringFromTime(utcTime time.Time) string {
	//answertime.Parse(SyncStandardDateFormat, utcT
	// TODO(doug4j@gmail.com): Finish my implementation
	return ""
}

//CreateSentSyncState transforms a int32 into a SentSyncStateEnum.
func CreateSentSyncState(val int32) (SentSyncStateEnum, error) {
	theVal := SentSyncStateEnum(val)
	switch theVal {
	case SentSyncStateEnum_PersistedNeverSentToPeer:
		return SentSyncStateEnum_PersistedNeverSentToPeer, nil
	case SentSyncStateEnum_PersistedFirstTimeSentToPeer:
		return SentSyncStateEnum_PersistedFirstTimeSentToPeer, nil
	case SentSyncStateEnum_PersistedStandardSentToPeer:
		return SentSyncStateEnum_PersistedStandardSentToPeer, nil
	case SentSyncStateEnum_PersistedFastDeleted:
		return SentSyncStateEnum_PersistedFastDeleted, nil
	default:
		return SentSyncStateEnum_PersistedNeverSentToPeer, fmt.Errorf("Value %v is not a valid enum", val)
	}
}
