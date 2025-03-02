package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var TightMinToCompress = 12

const (
	SegmentBytes SegmentType = iota
	SegmentMessageStart
	SegmentRectSeparator
	SegmentFullyParsedClientMessage
	SegmentFullyParsedServerMessage
	SegmentServerInitMessage
	SegmentConnectionClosed
	SegmentMessageEnd
)

type SegmentType int

func (seg SegmentType) String() string {
	switch seg {
	case SegmentBytes:
		return "SegmentBytes"
	case SegmentMessageStart:
		return "SegmentMessageStart"
	case SegmentMessageEnd:
		return "SegmentMessageEnd"
	case SegmentRectSeparator:
		return "SegmentRectSeparator"
	case SegmentFullyParsedClientMessage:
		return "SegmentFullyParsedClientMessage"
	case SegmentFullyParsedServerMessage:
		return "SegmentFullyParsedServerMessage"
	case SegmentServerInitMessage:
		return "SegmentServerInitMessage"
	case SegmentConnectionClosed:
		return "SegmentConnectionClosed"
	}

	return ""
}

type RfbSegment struct {
	Bytes              []byte
	SegmentType        SegmentType
	UpcomingObjectType int
	Message            interface{}
}

type SegmentConsumer interface {
	Consume(segment *RfbSegment) error
}

type RfbReadHelper struct {
	io.Reader
	Listeners  *MultiListener
	savedBytes *bytes.Buffer
}

func NewRfbReadHelper(r io.Reader) *RfbReadHelper {
	return &RfbReadHelper{Reader: r, Listeners: &MultiListener{}}
}

func (r *RfbReadHelper) StartByteCollection() {
	r.savedBytes = &bytes.Buffer{}
}

func (r *RfbReadHelper) EndByteCollection() []byte {
	bts := r.savedBytes.Bytes()
	r.savedBytes = nil
	return bts
}

func (r *RfbReadHelper) ReadDiscrete(p []byte) (int, error) {
	return r.Read(p)
}

func (r *RfbReadHelper) SendRectSeparator(upcomingRectType int) error {
	seg := &RfbSegment{SegmentType: SegmentRectSeparator, UpcomingObjectType: upcomingRectType}
	return r.Listeners.Consume(seg)
}

func (r *RfbReadHelper) SendMessageStart(upcomingMessageType ServerMessageType) error {
	seg := &RfbSegment{SegmentType: SegmentMessageStart, UpcomingObjectType: int(upcomingMessageType)}
	return r.Listeners.Consume(seg)
}

func (r *RfbReadHelper) SendMessageEnd(messageType ServerMessageType) error {
	seg := &RfbSegment{SegmentType: SegmentMessageEnd, UpcomingObjectType: int(messageType)}
	return r.Listeners.Consume(seg)
}

func (r *RfbReadHelper) PublishBytes(p []byte) error {
	seg := &RfbSegment{Bytes: p, SegmentType: SegmentBytes}
	return r.Listeners.Consume(seg)
}

func (r *RfbReadHelper) Read(p []byte) (n int, err error) {
	readLen, err := r.Reader.Read(p)
	if err != nil {
		return 0, fmt.Errorf("failed to read RFB bytes onto buffer")
	}
	if r.savedBytes != nil {
		_, err := r.savedBytes.Write(p)
		if err != nil {
			return 0, fmt.Errorf("failed to save bytes in memory buffer: %v", err)
		}
	}
	// write the bytes to the Listener for further processing
	seg := &RfbSegment{Bytes: p[:readLen], SegmentType: SegmentBytes}
	err = r.Listeners.Consume(seg)
	if err != nil {
		return 0, err
	}
	return readLen, err
}

func (r *RfbReadHelper) ReadBytes(count int) ([]byte, error) {
	buff := make([]byte, count)

	lengthRead, err := io.ReadFull(r, buff)
	if lengthRead != count {
		return nil, fmt.Errorf("unable to read bytes: lengthRead=%d, countExpected=%d", lengthRead, count)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes: %v", err)
	}
	return buff, nil
}

func (r *RfbReadHelper) ReadUint8() (uint8, error) {
	var myUint uint8
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}
func (r *RfbReadHelper) ReadUint16() (uint16, error) {
	var myUint uint16
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}
func (r *RfbReadHelper) ReadUint32() (uint32, error) {
	var myUint uint32
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}
func (r *RfbReadHelper) ReadCompactLen() (int, error) {
	var err error
	part, err := r.ReadUint8()
	//byteCount := 1
	len := uint32(part & 0x7F)
	if (part & 0x80) != 0 {
		part, err = r.ReadUint8()
		//byteCount++
		len |= uint32(part&0x7F) << 7
		if (part & 0x80) != 0 {
			part, err = r.ReadUint8()
			//byteCount++
			len |= uint32(part&0xFF) << 14
		}
	}

	return int(len), err
}

func (r *RfbReadHelper) ReadTightData(dataSize int) ([]byte, error) {
	if int(dataSize) < TightMinToCompress {
		return r.ReadBytes(int(dataSize))
	}
	zlibDataLen, err := r.ReadCompactLen()
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %v", err)
	}
	return r.ReadBytes(zlibDataLen)
}
