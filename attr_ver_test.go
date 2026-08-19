package gtp5gnl

import (
	"testing"

	"github.com/khirono/go-nl"
)

func TestDecodeVersionInfo(t *testing.T) {
	b := encodeAttrList(t, nl.AttrList{
		{Type: VERSION_SHARED_MARK_ABI, Value: nl.AttrU8(1)},
		{Type: 15, Value: nl.AttrU32(0xdeadbeef)},
		{Type: VERSION_DRIVER, Value: nl.AttrString("0.10.2")},
	})

	info, err := DecodeVersionInfo(b)
	if err != nil {
		t.Fatalf("DecodeVersionInfo: %v", err)
	}
	if info.DriverVersion != "0.10.2" {
		t.Fatalf("driver version = %q, want %q", info.DriverVersion, "0.10.2")
	}
	if info.SharedMarkABI == nil {
		t.Fatal("shared mark ABI is nil")
	}
	if *info.SharedMarkABI != 1 {
		t.Fatalf("shared mark ABI = %d, want 1", *info.SharedMarkABI)
	}

	version, err := DecodeVersion(b)
	if err != nil {
		t.Fatalf("DecodeVersion: %v", err)
	}
	if version != "0.10.2" {
		t.Fatalf("DecodeVersion = %q, want %q", version, "0.10.2")
	}
}

func TestDecodeLegacyVersionInfo(t *testing.T) {
	b := encodeAttrList(t, nl.AttrList{
		{Type: VERSION_DRIVER, Value: nl.AttrString("0.10.2")},
	})

	info, err := DecodeVersionInfo(b)
	if err != nil {
		t.Fatalf("DecodeVersionInfo: %v", err)
	}
	if info.DriverVersion != "0.10.2" {
		t.Fatalf("driver version = %q, want %q", info.DriverVersion, "0.10.2")
	}
	if info.SharedMarkABI != nil {
		t.Fatalf("shared mark ABI = %d, want nil", *info.SharedMarkABI)
	}
}

func TestDecodeVersionInfoRejectsEmptyCapability(t *testing.T) {
	b := encodeAttrList(t, nl.AttrList{
		{Type: VERSION_DRIVER, Value: nl.AttrString("0.10.2")},
		{Type: VERSION_SHARED_MARK_ABI},
	})

	if _, err := DecodeVersionInfo(b); err == nil {
		t.Fatal("DecodeVersionInfo succeeded, want error")
	}
}
