package gtp5gnl

import (
	"bytes"
	"io"

	"github.com/khirono/go-nl"
)

const (
	VERSION_DRIVER = iota
	VERSION_SHARED_MARK_ABI
)

// VersionInfo contains the gtp5g driver version and optional ABI
// capabilities returned by CMD_GET_VERSION. SharedMarkABI is nil for legacy
// modules that only return the driver version attribute.
type VersionInfo struct {
	DriverVersion string
	SharedMarkABI *uint8
}

func DecodeVersion(b []byte) (string, error) {
	info, err := DecodeVersionInfo(b)
	if err != nil {
		return "", err
	}
	return info.DriverVersion, nil
}

// DecodeVersionInfo decodes all CMD_GET_VERSION attributes without relying
// on their order, while remaining compatible with legacy version-only replies.
func DecodeVersionInfo(b []byte) (*VersionInfo, error) {
	info := new(VersionInfo)

	for len(b) > 0 {
		hdr, n, err := nl.DecodeAttrHdr(b)
		if err != nil {
			return nil, err
		}
		attrLen := int(hdr.Len)
		if attrLen < n || attrLen > len(b) {
			return nil, io.ErrUnexpectedEOF
		}

		payload := b[n:attrLen]
		switch hdr.MaskedType() {
		case VERSION_DRIVER:
			info.DriverVersion = string(bytes.Trim(payload, "\x00"))
		case VERSION_SHARED_MARK_ABI:
			if len(payload) < 1 {
				return nil, io.ErrUnexpectedEOF
			}
			version := payload[0]
			info.SharedMarkABI = &version
		}

		alignedLen := hdr.Len.Align()
		if alignedLen > len(b) {
			return nil, io.ErrUnexpectedEOF
		}
		b = b[alignedLen:]
	}

	return info, nil
}
