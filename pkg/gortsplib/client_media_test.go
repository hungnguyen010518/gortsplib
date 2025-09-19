package gortsplib

import (
	"testing"

	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
)

func Test_clientMedia_start(t *testing.T) {
	type fields struct {
		c                      *Client
		media                  *description.Media
		formats                map[uint8]*clientFormat
		tcpChannel             int
		udpRTPListener         *clientUDPListener
		udpRTCPListener        *clientUDPListener
		tcpRTPFrame            *base.InterleavedFrame
		tcpRTCPFrame           *base.InterleavedFrame
		tcpBuffer              []byte
		writePacketRTPInQueue  func([]byte)
		writePacketRTCPInQueue func([]byte)
		onPacketRTCP           OnPacketRTCPFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &clientMedia{
				c:                      tt.fields.c,
				media:                  tt.fields.media,
				formats:                tt.fields.formats,
				tcpChannel:             tt.fields.tcpChannel,
				udpRTPListener:         tt.fields.udpRTPListener,
				udpRTCPListener:        tt.fields.udpRTCPListener,
				tcpRTPFrame:            tt.fields.tcpRTPFrame,
				tcpRTCPFrame:           tt.fields.tcpRTCPFrame,
				tcpBuffer:              tt.fields.tcpBuffer,
				writePacketRTPInQueue:  tt.fields.writePacketRTPInQueue,
				writePacketRTCPInQueue: tt.fields.writePacketRTCPInQueue,
				onPacketRTCP:           tt.fields.onPacketRTCP,
			}
			cm.start()
		})
	}
}
