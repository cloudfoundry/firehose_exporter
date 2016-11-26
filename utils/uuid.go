package utils

import (
	"encoding/binary"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

func UUIDToString(uuid *events.UUID) string {
	var uuidBytes [16]byte

	if uuid == nil {
		return ""
	}

	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}

func StringToUUID(id string) *events.UUID {
	idHex, err := uuid.ParseHex(id)
	if err != nil {
		return nil
	}

	return &events.UUID{
		Low:  proto.Uint64(binary.LittleEndian.Uint64(idHex[:8])),
		High: proto.Uint64(binary.LittleEndian.Uint64(idHex[8:])),
	}
}
