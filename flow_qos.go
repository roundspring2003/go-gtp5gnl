package gtp5gnl

import (
	"fmt"
	"io"

	"github.com/khirono/go-nl"
)

const (
	FLOW_QOS_UNSPEC = iota
	FLOW_QOS_VERSION
	FLOW_QOS_POLICY_ID
	_ // Attribute 3 is reserved for the removed TC class ID.
	FLOW_QOS_FLAGS
	FLOW_QOS_GENERATION
)

const (
	SHARED_MARK_ABI_VERSION  uint8  = 1
	FLOW_QOS_BINDING_VERSION uint8  = 2
	FLOW_QOS_VALID           uint8  = 0x01
	FLOW_QOS_POLICY_ID_MAX   uint32 = 0x00ffffff
)

// FlowQoS is the PDR-local binding used by the gtp5g shared skb mark ABI.
// PolicyID is encoded in the low 24 bits of skb->mark. Egress BPF selects
// the TC class from the policy map.
type FlowQoS struct {
	Version    uint8
	PolicyID   uint32
	Flags      uint8
	Generation uint32
}

// NewFlowQoSAttr builds the nested PDR_FLOW_QOS Generic Netlink attribute.
// A valid binding carries all fields. A binding with Flags == 0 clears an
// existing binding and only publishes the version, flags and optional
// generation fields required by that operation.
func NewFlowQoSAttr(flowQoS FlowQoS) (nl.Attr, error) {
	if flowQoS.Version != FLOW_QOS_BINDING_VERSION {
		return nl.Attr{}, fmt.Errorf(
			"unsupported FlowQoS binding version: got %d, want %d",
			flowQoS.Version,
			FLOW_QOS_BINDING_VERSION,
		)
	}
	if flowQoS.Flags & ^FLOW_QOS_VALID != 0 {
		return nl.Attr{}, fmt.Errorf("unsupported FlowQoS flags: %#x", flowQoS.Flags)
	}
	if flowQoS.Flags&FLOW_QOS_VALID != 0 && flowQoS.PolicyID > FLOW_QOS_POLICY_ID_MAX {
		return nl.Attr{}, fmt.Errorf(
			"FlowQoS policy ID %d exceeds maximum %d",
			flowQoS.PolicyID,
			FLOW_QOS_POLICY_ID_MAX,
		)
	}

	attrs := nl.AttrList{
		{Type: FLOW_QOS_VERSION, Value: nl.AttrU8(flowQoS.Version)},
	}
	if flowQoS.Flags&FLOW_QOS_VALID != 0 {
		attrs = append(attrs, nl.Attr{Type: FLOW_QOS_POLICY_ID, Value: nl.AttrU32(flowQoS.PolicyID)})
	}
	attrs = append(attrs, nl.Attr{Type: FLOW_QOS_FLAGS, Value: nl.AttrU8(flowQoS.Flags)})
	if flowQoS.Generation != 0 {
		attrs = append(attrs, nl.Attr{
			Type:  FLOW_QOS_GENERATION,
			Value: nl.AttrU32(flowQoS.Generation),
		})
	}

	return nl.Attr{Type: PDR_FLOW_QOS, Value: attrs}, nil
}

// DecodeFlowQoS decodes the payload of a nested PDR_FLOW_QOS attribute.
func DecodeFlowQoS(b []byte) (FlowQoS, error) {
	var flowQoS FlowQoS

	for len(b) > 0 {
		hdr, n, err := nl.DecodeAttrHdr(b)
		if err != nil {
			return flowQoS, err
		}
		attrLen := int(hdr.Len)
		if attrLen < n || attrLen > len(b) {
			return flowQoS, io.ErrUnexpectedEOF
		}

		payload := b[n:attrLen]
		switch hdr.MaskedType() {
		case FLOW_QOS_VERSION:
			if len(payload) < 1 {
				return flowQoS, io.ErrUnexpectedEOF
			}
			flowQoS.Version = payload[0]
		case FLOW_QOS_POLICY_ID:
			if len(payload) < 4 {
				return flowQoS, io.ErrUnexpectedEOF
			}
			flowQoS.PolicyID = native.Uint32(payload)
		case FLOW_QOS_FLAGS:
			if len(payload) < 1 {
				return flowQoS, io.ErrUnexpectedEOF
			}
			flowQoS.Flags = payload[0]
		case FLOW_QOS_GENERATION:
			if len(payload) < 4 {
				return flowQoS, io.ErrUnexpectedEOF
			}
			flowQoS.Generation = native.Uint32(payload)
		}

		alignedLen := hdr.Len.Align()
		if alignedLen > len(b) {
			return flowQoS, io.ErrUnexpectedEOF
		}
		b = b[alignedLen:]
	}

	return flowQoS, nil
}
