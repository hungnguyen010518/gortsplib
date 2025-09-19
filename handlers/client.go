package handlers

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"

	"dvrs.lib/RTSPClient/constant"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type ClientKey struct {
	CallKey
	groupName string
	ch        int
}

type Client struct {
	ClientKey
	client    *gortsplib.Client
	rtspState constant.RTSPState
}

func (ck ClientKey) Hash() uint32 {
	h := cmap.Fnv32(ck.Name)
	h ^= cmap.Fnv32(strconv.Itoa(int(ck.RecorderType)))
	h ^= cmap.Fnv32(strconv.Itoa(ck.ch))
	h ^= cmap.Fnv32(ck.groupName)
	return h
}

func (c Client) Lock() {
	rtspClient := GetRTSPClient()
	rtspClient.cs.listClient.OuterLock(c.ClientKey)
}

func (c Client) Unlock() {
	rtspClient := GetRTSPClient()
	rtspClient.cs.listClient.OuterUnLock(c.ClientKey)
}

func createClient(c *Client, mediaTransport string, keepAliveTime int, ed137Version string, interleave string) {
	var wg67Version string
	switch ed137Version {
	case "ED137A":
		wg67Version = "recorder.00"
	case "ED137B":
		wg67Version = "recorder.01"
	case "ED137C":
		wg67Version = "recorder.02"
	default:
		wg67Version = "recorder.01"
	}
	var b_interleave bool
	if interleave == "enable" {
		b_interleave = true
	} else {
		b_interleave = false
	}
	if mediaTransport == "tcp" {
		transportTCP := gortsplib.TransportTCP
		c.client = &gortsplib.Client{
			Transport:       &transportTCP,
			KeepAlivePeriod: time.Duration(keepAliveTime * int(time.Second)),
			Wg67Version:     wg67Version,
			UseInterleaved:  b_interleave,
		}
	} else {
		transportUDP := gortsplib.TransportUDP
		c.client = &gortsplib.Client{
			Transport:       &transportUDP,
			KeepAlivePeriod: time.Duration(keepAliveTime * int(time.Second)),
			Wg67Version:     wg67Version,
			UseInterleaved:  b_interleave,
		}
	}
}

func (c *Client) CloseByNormal(crd *CRD) {
	rtspClient := GetRTSPClient()
	if !crd.Disabled && rtspClient.ed137Versions[c.ch] == "ED137C" {
		crd.Properties.DisconnectCause.Value = strconv.Itoa(int(GetDisconnectCause(crd.Properties.SipDisconnectCause.Value, nil)))
		crdByt, _ := xml.MarshalIndent(crd, "", "    ")
		if _, err := c.client.SetParameter(nil, crdByt); err != nil {
			rtspClient.LogDebug("name", c.Name, "recorderType:", "channel:", c.ch, "Error sening SetParameter request:", err)
			return
		}
	}
	c.client.Close()
	c.rtspState = constant.RTSP_STATE_DISCONNECT
}

func (c *Client) Start(crd CRD) (*base.URL, error) {
	if c.rtspState != constant.RTSP_STATE_START || c.client.IsClose() {
		keepAliveTime, _ := strconv.Atoi(rtspClient.keepTimeAlives[c.ch])
		createClient(c, rtspClient.mediaTransports[c.ch], keepAliveTime, rtspClient.ed137Versions[c.ch], rtspClient.interleaves[c.ch])
	}
	var u *base.URL
	var err error
	switch c.RecorderType {
	case constant.RET_PHONE:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/" + strings.ToLower(c.Name))
	case constant.RET_BRIEF:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/" + strings.ToLower(c.Name) + "_brief")
	case constant.RET_AMBIENT:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/ambient")
	case constant.RET_PHONE_GROUP:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/phone")
	case constant.RET_RADIO_GROUP:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/radio")
	case constant.RET_BRIEF_GROUP:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/brief")
	case constant.RET_RADIO_TX:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/" + strings.ToLower(c.Name) + "_ptt")
	default:
		u, err = base.ParseURL("rtsp://" + rtspClient.recAddrs[c.ch] + "/" + strings.ToLower(crd.VCSUser) + "/" + strings.ToLower(c.Name) + "_squ")
	}
	if err != nil {
		return nil, err
	}
	if c.rtspState != constant.RTSP_STATE_START {
		if err = c.client.Start(u.Scheme, u.Host); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (c *Client) AnnounceSetup(u *base.URL) error {
	if _, err := c.client.Announce(u, &rtspClient.desc); err != nil {
		return err
	}
	if err := c.client.SetupAll(u, rtspClient.desc.Medias); err != nil {
		return err
	}
	c.rtspState = constant.RTSP_STATE_SETUP
	return nil
}

func (c *Client) SetParameter(u *base.URL, crd CRD, crdByt []byte) error {
	if !crd.Disabled && (c.RecorderType == constant.RET_PHONE || c.RecorderType == constant.RET_BRIEF || c.RecorderType == constant.RET_AMBIENT || rtspClient.ed137Versions[c.ch] == "ED137C") {
		if _, err := c.client.SetParameter(u, crdByt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Record(crd CRD, crdByt []byte) error {
	if c.RecorderType == constant.RET_PHONE && c.rtspState == constant.RTSP_STATE_PAUSE {
		if _, err := c.client.SetParameter(nil, crdByt); err != nil {
			return err
		}
	} else {
		if c.rtspState == constant.RTSP_STATE_SETUP || c.rtspState == constant.RTSP_STATE_PAUSE {
			if _, err := c.client.Record(crdByt); err != nil {
				return err
			}
		}
	}
	c.rtspState = constant.RTSP_STATE_RECORD
	return nil
}

func (c *Client) Pause(crd CRD, crdByt []byte) error {
	if c.RecorderType == constant.RET_PHONE {
		if _, err := c.client.SetParameter(nil, crdByt); err != nil {
			return err
		}
	} else {
		if _, err := c.client.Pause(crdByt); err != nil {
			return err
		}
	}
	c.rtspState = constant.RTSP_STATE_PAUSE
	return nil
}

func (c *Client) CloseByErr() {
	c.client.Close()
	c.rtspState = constant.RTSP_STATE_DISCONNECT
}
