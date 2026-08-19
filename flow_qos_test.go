package gtp5gnl

import (
	"testing"

	"github.com/khirono/go-nl"
)

func encodeAttr(t *testing.T, attr nl.Attr) []byte {
	t.Helper()

	b := make([]byte, attr.Len())
	n, err := attr.Encode(b)
	if err != nil {
		t.Fatalf("encode attribute: %v", err)
	}
	return b[:n]
}

func encodeAttrList(t *testing.T, attrs nl.AttrList) []byte {
	t.Helper()

	b := make([]byte, attrs.Len())
	n, err := attrs.Encode(b)
	if err != nil {
		t.Fatalf("encode attribute list: %v", err)
	}
	return b[:n]
}

func TestPDRABIAttributeNumbers(t *testing.T) {
	if PDR_PDN_TYPE != 13 {
		t.Fatalf("PDR_PDN_TYPE = %d, want 13", PDR_PDN_TYPE)
	}
	if PDR_FLOW_QOS != 14 {
		t.Fatalf("PDR_FLOW_QOS = %d, want 14", PDR_FLOW_QOS)
	}
}

func TestFlowQoSAttrRoundTrip(t *testing.T) {
	want := FlowQoS{
		Version:    SHARED_MARK_ABI_VERSION,
		PolicyID:   1001,
		TCClassID:  0x00010020,
		Flags:      FLOW_QOS_VALID,
		Generation: 7,
	}

	attr, err := NewFlowQoSAttr(want)
	if err != nil {
		t.Fatalf("NewFlowQoSAttr: %v", err)
	}
	b := encodeAttr(t, attr)

	hdr, n, err := nl.DecodeAttrHdr(b)
	if err != nil {
		t.Fatalf("DecodeAttrHdr: %v", err)
	}
	if hdr.MaskedType() != PDR_FLOW_QOS {
		t.Fatalf("attribute type = %d, want %d", hdr.MaskedType(), PDR_FLOW_QOS)
	}
	if !hdr.Nested() {
		t.Fatal("PDR_FLOW_QOS attribute is not marked nested")
	}

	got, err := DecodeFlowQoS(b[n:int(hdr.Len)])
	if err != nil {
		t.Fatalf("DecodeFlowQoS: %v", err)
	}
	if got != want {
		t.Fatalf("FlowQoS = %+v, want %+v", got, want)
	}
}

func TestDecodePDRFlowQoS(t *testing.T) {
	want := FlowQoS{
		Version:    SHARED_MARK_ABI_VERSION,
		PolicyID:   0x00ffffff,
		TCClassID:  0x00010030,
		Flags:      FLOW_QOS_VALID,
		Generation: 11,
	}
	flowQoSAttr, err := NewFlowQoSAttr(want)
	if err != nil {
		t.Fatalf("NewFlowQoSAttr: %v", err)
	}

	b := encodeAttrList(t, nl.AttrList{
		{Type: PDR_ID, Value: nl.AttrU16(9)},
		flowQoSAttr,
	})
	pdr, err := DecodePDR(b)
	if err != nil {
		t.Fatalf("DecodePDR: %v", err)
	}
	if pdr.FlowQoS == nil {
		t.Fatal("PDR FlowQoS binding is nil")
	}
	if *pdr.FlowQoS != want {
		t.Fatalf("PDR FlowQoS = %+v, want %+v", *pdr.FlowQoS, want)
	}
}

func TestFlowQoSClearAttr(t *testing.T) {
	clear := FlowQoS{
		Version:    SHARED_MARK_ABI_VERSION,
		Generation: 8,
	}
	attr, err := NewFlowQoSAttr(clear)
	if err != nil {
		t.Fatalf("NewFlowQoSAttr: %v", err)
	}

	nested, ok := attr.Value.(nl.AttrList)
	if !ok {
		t.Fatalf("attribute value type = %T, want nl.AttrList", attr.Value)
	}
	if len(nested) != 3 {
		t.Fatalf("clear attribute has %d nested fields, want 3", len(nested))
	}
	if nested[0].Type != FLOW_QOS_VERSION ||
		nested[1].Type != FLOW_QOS_FLAGS ||
		nested[2].Type != FLOW_QOS_GENERATION {
		t.Fatalf("unexpected clear attribute fields: %+v", nested)
	}
}

func TestNewFlowQoSAttrValidation(t *testing.T) {
	tests := []struct {
		name    string
		flowQoS FlowQoS
	}{
		{name: "zero version", flowQoS: FlowQoS{}},
		{name: "unknown flags", flowQoS: FlowQoS{Version: 1, Flags: 0x02}},
		{
			name: "policy ID overflow",
			flowQoS: FlowQoS{
				Version:  1,
				Flags:    FLOW_QOS_VALID,
				PolicyID: FLOW_QOS_POLICY_ID_MAX + 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewFlowQoSAttr(tt.flowQoS); err == nil {
				t.Fatal("NewFlowQoSAttr succeeded, want error")
			}
		})
	}
}
