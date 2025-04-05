package protocol

import (
	"crypto/sha256"
	"encoding/base32"
	"strings"
)

const (
	DeviceIDLength = 32
	ShortIDLength  = 4
)

type (
	DeviceID [DeviceIDLength]byte
	ShortID  [ShortIDLength]byte
)

func NewDeviceID(rawCert []byte) (DeviceID, error) {
	return DeviceID(sha256.Sum256(rawCert)), nil
}

func (d DeviceID) String() string {
	id := base32.StdEncoding.EncodeToString(d[:])
	id = strings.Trim(id, "=")
	return id
}

func (d DeviceID) GoString() string {
	return d.String()
}
