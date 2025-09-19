package utils

import (
	"github.com/pion/rtp"
	"github.com/zaf/g711"
)

func ConvertCodec(pkt *rtp.Packet, codec string) bool {
	if codec != "g711alaw" && codec != "g711ulaw" || pkt.PayloadType != 8 && pkt.PayloadType != 0 && pkt.PayloadType != 96 {
		return false
	}
	if codec == "g711ulaw" && pkt.PayloadType == 0 || codec == "g711alaw" && pkt.PayloadType == 8 {
		return true
	}
	payLoad := pkt.Payload
	switch codec {
	case "g711alaw":
		if pkt.PayloadType == 96 {
			payLoad = g711.EncodeAlaw(payLoad)
		} else {
			payLoad = g711.Ulaw2Alaw(payLoad)
		}
		pkt.Header.PayloadType = 8
	case "g711ulaw":
		if pkt.PayloadType == 96 {
			payLoad = g711.EncodeUlaw(payLoad)
		} else {
			payLoad = g711.Alaw2Ulaw(payLoad)
		}
		pkt.Header.PayloadType = 0
	default:
	}
	pkt.Payload = payLoad
	return true
}
