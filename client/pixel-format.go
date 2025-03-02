package client

import (
	"bytes"
	"encoding/binary"
	"github.com/borderzero/vncproxy/common"
	"io"
)

func readPixelFormat(r io.Reader, result *common.PixelFormat) error {
	var rawPixelFormat [16]byte
	if _, err := io.ReadFull(r, rawPixelFormat[:]); err != nil {
		return err
	}

	var pfBoolByte uint8
	brPF := bytes.NewReader(rawPixelFormat[:])
	if err := binary.Read(brPF, binary.BigEndian, &result.BPP); err != nil {
		return err
	}

	if err := binary.Read(brPF, binary.BigEndian, &result.Depth); err != nil {
		return err
	}

	if err := binary.Read(brPF, binary.BigEndian, &pfBoolByte); err != nil {
		return err
	}

	if pfBoolByte != 0 {
		// Big endian is true
		result.BigEndian = 1
	}

	if err := binary.Read(brPF, binary.BigEndian, &pfBoolByte); err != nil {
		return err
	}

	if pfBoolByte != 0 {
		// True Color is true. So we also have to read all the color max & shifts.
		result.TrueColor = 1

		if err := binary.Read(brPF, binary.BigEndian, &result.RedMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.GreenMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.BlueMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.RedShift); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.GreenShift); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.BlueShift); err != nil {
			return err
		}
	}

	return nil
}

func writePixelFormat(format *common.PixelFormat) ([]byte, error) {
	var buf bytes.Buffer

	// Byte 1
	if err := binary.Write(&buf, binary.BigEndian, format.BPP); err != nil {
		return nil, err
	}

	// Byte 2
	if err := binary.Write(&buf, binary.BigEndian, format.Depth); err != nil {
		return nil, err
	}

	var boolByte byte
	if format.BigEndian == 1 {
		boolByte = 1
	} else {
		boolByte = 0
	}

	// Byte 3 (BigEndian)
	if err := binary.Write(&buf, binary.BigEndian, boolByte); err != nil {
		return nil, err
	}

	if format.TrueColor == 1 {
		boolByte = 1
	} else {
		boolByte = 0
	}

	// Byte 4 (TrueColor)
	if err := binary.Write(&buf, binary.BigEndian, boolByte); err != nil {
		return nil, err
	}

	// If we have true color enabled then we have to fill in the rest of the
	// structure with the color values.
	if format.TrueColor == 1 {
		if err := binary.Write(&buf, binary.BigEndian, format.RedMax); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, format.GreenMax); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, format.BlueMax); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, format.RedShift); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, format.GreenShift); err != nil {
			return nil, err
		}

		if err := binary.Write(&buf, binary.BigEndian, format.BlueShift); err != nil {
			return nil, err
		}
	}

	return buf.Bytes()[0:16], nil
}
