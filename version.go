package gtp5gnl

import (
	"fmt"
	"syscall"

	"github.com/khirono/go-genl"
	"github.com/khirono/go-nl"
)

func GetVersion(c *Client) (string, error) {
	info, err := GetVersionInfo(c)
	if err != nil {
		return "", err
	}
	return info.DriverVersion, nil
}

// GetVersionInfo gets the driver version and optional ABI capabilities from
// gtp5g. Legacy modules return a VersionInfo with SharedMarkABI == nil.
func GetVersionInfo(c *Client) (*VersionInfo, error) {
	flags := syscall.NLM_F_ACK
	req := nl.NewRequest(c.ID, flags)
	err := req.Append(genl.Header{Cmd: CMD_GET_VERSION})
	if err != nil {
		return nil, err
	}

	rsps, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if len(rsps) != 1 {
		return nil, fmt.Errorf("invalid Version")
	}
	info, err := DecodeVersionInfo(rsps[0].Body[genl.SizeofHeader:])
	if err != nil {
		return nil, err
	}
	return info, nil
}
