package handlers

/*
#cgo CXXFLAGS: -std=c++11
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/xml"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"dvrs.lib/RTSPClient/constant"
	"dvrs.lib/RTSPClient/models"
	"dvrs.lib/RTSPClient/utils"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pion/rtp"
)

type CallKey struct {
	Name         string
	RecorderType constant.RecorderType
}

type RTPInfo struct {
	// for Split and Merge RTP Packet
	lastSSRC       uint32
	lastActSeq     uint16
	lastDisTs      uint32
	lastActTs      uint32
	SegPloadLength int
}

type BriefStateInfo struct {
	state    constant.BriefState
	crdMsg   string
	crdMsgId string
}

type GroupStateInfo struct {
	state    constant.GroupState
	crdMsg   string
	crdMsgId string
}

type RadioButtonStateInfo struct {
	state    constant.RadioButtonState
	crdMsg   string
	crdMsgId string
}

type CallStateInfo struct {
	state    constant.CallState
	crdMsg   string
	crdMsgId string
}

type CallMediaStateInfo struct {
	state    constant.CallMediaState
	crdMsg   string
	crdMsgId string
}

type Done struct {
	start  chan bool
	finish chan bool
}

func (ck CallKey) Hash() uint32 {
	h := cmap.Fnv32(ck.Name)
	h ^= cmap.Fnv32(strconv.Itoa(int(ck.RecorderType)))
	return h
}

func (callInfo CallInfo) Lock() {
	rtspClient.callModel.listCallInfo.OuterLock(callInfo.CallKey)
}

func (callInfo CallInfo) Unlock() {
	rtspClient.callModel.listCallInfo.OuterUnLock(callInfo.CallKey)
}

func (callInfo CallInfo) getClient(ch int) Client {
	key := ClientKey{
		CallKey: callInfo.CallKey,
		ch:      ch,
	}
	rtspClient := GetRTSPClient()
	if rtspClient.cs.listClient.Has(key) {
		c, _ := rtspClient.cs.listClient.Get(key)
		return c
	} else {
		c := Client{
			client:    &gortsplib.Client{},
			rtspState: constant.RTSP_STATE_NULL,
			ClientKey: key,
		}
		c.Lock()
		rtspClient.cs.listClient.Set(key, c)
		c.Unlock()
		return c
	}
}

func (callInfo CallInfo) updateClient(ch int, c *Client) {
	key := ClientKey{
		CallKey: callInfo.CallKey,
		ch:      ch,
	}
	c.Lock()
	rtspClient.cs.listClient.Set(key, *c)
	c.Unlock()
}

func (callInfo CallInfo) getClientIfExist(ch int) Client {
	key := ClientKey{
		CallKey: callInfo.CallKey,
		ch:      ch,
	}
	if rtspClient.cs.listClient.Has(key) {
		c, _ := rtspClient.cs.listClient.Get(key)
		return c
	}
	return Client{
		client:    &gortsplib.Client{},
		rtspState: constant.RTSP_STATE_NULL,
		ClientKey: key,
	}

}

type EventQueue struct {
	chBriefStateInfo           chan BriefStateInfo
	chGroupStateInfo           chan GroupStateInfo
	chCallStateInfo            chan CallStateInfo
	chCallMediaStateInfo       chan CallMediaStateInfo
	chRadioButtonStateInfo     chan RadioButtonStateInfo
	chLastCallMediaStateInfo   chan CallMediaStateInfo
	chLastRadioButtonStateInfo chan RadioButtonStateInfo
}

type SleepHandle struct {
	sleep   bool
	goSleep chan bool
}

type ThreadHandle struct {
	chDone chan bool
	wg     *sync.WaitGroup
}

type RTPClient struct {
	ListenPort         int
	chUpdateListenConn chan int
	chRecordRTP        chan bool
}

type CallInfo struct {
	CallKey
	RTPClient
	EventQueue
	ThreadHandle
	SleepHandle
	blockState constant.BlockState
}

func (callInfo CallInfo) getCRD(ch int) CRD {
	key := ClientKey{
		CallKey: callInfo.CallKey,
		ch:      ch,
	}
	if rtspClient.crds.listCRD.Has(key) {
		crd, _ := rtspClient.crds.listCRD.Get(key)
		return crd
	} else {
		return CRD{}
	}
}

func (callInfo CallInfo) updateCRD(ch int, crd *CRD) {
	key := ClientKey{
		CallKey: callInfo.CallKey,
		ch:      ch,
	}
	rtspClient.crds.listCRD.Set(key, *crd)
}

func (callInfo *CallInfo) updateListenConn(listenPort int) {
	select {
	case callInfo.chUpdateListenConn <- listenPort:
	default:
	}
}

func (callInfo *CallInfo) doRecordRTP(flag bool) {
	callInfo.chRecordRTP <- flag
}

func (callInfo CallInfo) atLeastChannelRecord() bool {
	for i := 0; i < rtspClient.MaxCh; i++ {
		c := callInfo.getClientIfExist(i)
		if c.rtspState == constant.RTSP_STATE_RECORD || c.RecorderType == constant.RET_PHONE && c.rtspState == constant.RTSP_STATE_PAUSE {
			return true
		}
	}
	return false
}

func (callInfo CallInfo) UpdatelistenPort(listenPort int) {
	defer callInfo.updateListenConn(listenPort)
	callInfo.Lock()
	defer callInfo.Unlock()
	oldCallInfo, ok := rtspClient.GetCallInfoIfExist(callInfo.CallKey)
	if !ok {
		return
	}
	defer rtspClient.updateCallInfoSkipLock(callInfo.CallKey, &oldCallInfo)
	oldCallInfo.ListenPort = listenPort
}

func (callInfo CallInfo) isSleep() bool {
	oldCallInfo, ok := rtspClient.GetCallInfoIfExist(callInfo.CallKey)
	if !ok {
		return false
	}
	return oldCallInfo.sleep
}

func (callInfo CallInfo) setSleep(sleep bool) {
	callInfo.Lock()
	defer callInfo.Unlock()
	oldCallInfo, ok := rtspClient.GetCallInfoIfExist(callInfo.CallKey)
	if ok {
		oldCallInfo.sleep = sleep
		rtspClient.updateCallInfoSkipLock(callInfo.CallKey, &oldCallInfo)
	}
}

func (callInfo CallInfo) setBlockState(blockState constant.BlockState) {
	callInfo.Lock()
	defer callInfo.Unlock()
	oldCallInfo, ok := rtspClient.GetCallInfoIfExist(callInfo.CallKey)
	if ok {
		oldCallInfo.blockState = blockState
		rtspClient.updateCallInfoSkipLock(callInfo.CallKey, &oldCallInfo)
	}
}

func (callInfo CallInfo) HandleCallMediaState(callMediaState constant.CallMediaState, crdMsg string, crdMsgId string) {
	if !callInfo.isSleep() && callMediaState == constant.PJSUA_CALL_MEDIA_ACTIVE {
		select {
		case callInfo.chCallMediaStateInfo <- CallMediaStateInfo{state: callMediaState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
		default:
		}
	} else {
		select {
		case <-callInfo.chLastCallMediaStateInfo:
		default:
		}
		callInfo.chLastCallMediaStateInfo <- CallMediaStateInfo{state: callMediaState, crdMsg: crdMsg, crdMsgId: crdMsgId}
	}
}

func (callInfo CallInfo) HandleRadioButtonState(radioButtonState constant.RadioButtonState, crdMsg string, crdMsgId string) {
	if radioButtonState == constant.BUTTON_INVALID {
		select {
		case callInfo.chRadioButtonStateInfo <- RadioButtonStateInfo{state: radioButtonState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
			callInfo.setBlockState(constant.NORMAL_BLOCK)
		default:
		}
	} else if !callInfo.isSleep() && (radioButtonState == constant.TX_BUTTON_ON || radioButtonState == constant.RX_BUTTON_ON) {
		select {
		case callInfo.chRadioButtonStateInfo <- RadioButtonStateInfo{state: radioButtonState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
		default:
		}
	} else {
		select {
		case <-callInfo.chLastRadioButtonStateInfo:
		default:
		}
		callInfo.chLastRadioButtonStateInfo <- RadioButtonStateInfo{state: radioButtonState, crdMsg: crdMsg, crdMsgId: crdMsgId}
	}
}

func (callInfo CallInfo) HandleBriefState(briefState constant.BriefState, crdMsg string, crdMsgId string) {
	select {
	case callInfo.chBriefStateInfo <- BriefStateInfo{state: briefState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
		if briefState == constant.BRIEF_FALSE {
			callInfo.setBlockState(constant.NORMAL_BLOCK)
		}
	default:
	}
}

func (callInfo CallInfo) HandleGroupState(groupState constant.GroupState, crdMsg string, crdMsgId string) {
	select {
	case callInfo.chGroupStateInfo <- GroupStateInfo{state: groupState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
		if groupState == constant.GROUP_FALSE {
			callInfo.setBlockState(constant.NORMAL_BLOCK)
		}
	default:
	}
}

func (callInfo CallInfo) HandleCallState(callState constant.CallState, crdMsg string, crdMsgId string) {
	select {
	case callInfo.chCallStateInfo <- CallStateInfo{state: callState, crdMsg: crdMsg, crdMsgId: crdMsgId}:
		if callState == constant.PJSIP_INV_STATE_DISCONNECTED {
			callInfo.setBlockState(constant.NORMAL_BLOCK)
		}
	default:
	}
}

func (callInfo *CallInfo) SetCRD(crdMsg string, crdMsgId string) {
	MaxCh := rtspClient.MaxCh
	var wg sync.WaitGroup
	for i := 0; i < MaxCh; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := callInfo.getClient(i)
			crd := callInfo.getCRD(i)
			defer callInfo.updateCRD(i, &crd)
			if c.rtspState == constant.RTSP_STATE_DISCONNECT || c.rtspState == constant.RTSP_STATE_NULL {
				crd.Value = ""
			}
			crd.SetCRDInner(crdMsg, crdMsgId, rtspClient.ed137Versions[i], callInfo.RecorderType)
		}(i)
	}
	wg.Wait()

}

func (callInfo *CallInfo) doOnBriefState(briefState constant.BriefState) {
	MaxCh := rtspClient.MaxCh
	var wg sync.WaitGroup
	once := sync.Once{}
	for j := 0; j < MaxCh; j++ {
		if rtspClient.recGroups[j] {
			continue
		}
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			crd := callInfo.getCRD(j)
			defer callInfo.updateCRD(j, &crd)
			c := callInfo.getClient(j)
			defer callInfo.updateClient(j, &c)
			rtspState := c.rtspState
			if briefState == constant.BRIEF_TRUE && (rtspState == constant.RTSP_STATE_NULL || rtspState == constant.RTSP_STATE_DISCONNECT) {
				defer once.Do(func() {
					callInfo.doRecordRTP(true)
				})
				crd.EnableConnectBrief()
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")
				u, err := c.Start(crd)
				if err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error Starting:", err)
					c.rtspState = constant.RTSP_STATE_DISCONNECT
					return
				}
				if err = c.AnnounceSetup(u); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending Announce or SETUP request:", err)
					c.CloseByErr()
					return
				}
				if err = c.Record(crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending RECORD request:", err)
					c.CloseByErr()
					return
				}
			} else if briefState == constant.BRIEF_FALSE && rtspState == constant.RTSP_STATE_RECORD {
				defer once.Do(func() {
					callInfo.doRecordRTP(false)
				})
				crd.EnableDisconnectBrief()
				c.CloseByNormal(&crd)
			}
		}(j)
	}
	wg.Wait()
}

func (callInfo *CallInfo) doOnGroupState(groupState constant.GroupState) {
	MaxCh := rtspClient.MaxCh
	var wg sync.WaitGroup
	once := sync.Once{}
	for j := 0; j < MaxCh; j++ {
		if !rtspClient.recGroups[j] && callInfo.RecorderType != constant.RET_AMBIENT {
			continue
		}
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			crd := callInfo.getCRD(j)
			defer callInfo.updateCRD(j, &crd)
			c := callInfo.getClient(j)
			defer callInfo.updateClient(j, &c)
			rtspState := c.rtspState
			if groupState == constant.GROUP_TRUE && (rtspState == constant.RTSP_STATE_NULL || rtspState == constant.RTSP_STATE_DISCONNECT) {
				defer once.Do(func() {
					callInfo.doRecordRTP(true)
				})
				crd.EnableConnectGroup()
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")
				u, err := c.Start(crd)
				if err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error Starting:", err)
					c.rtspState = constant.RTSP_STATE_DISCONNECT
					return
				}
				if err = c.AnnounceSetup(u); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending Announce or SETUP request:", err)
					c.CloseByErr()
					return
				}
				if err = c.Record(crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending RECORD request:", err)
					c.CloseByErr()
					return
				}
			} else if groupState == constant.GROUP_FALSE && rtspState == constant.RTSP_STATE_RECORD {
				defer once.Do(func() {
					callInfo.doRecordRTP(false)
				})
				crd.EnableDisconnectGroup()
				c.CloseByNormal(&crd)
			}
		}(j)
	}
	wg.Wait()

}

func (callInfo *CallInfo) doOnRadioState(radioButtonState constant.RadioButtonState) {
	MaxCh := rtspClient.MaxCh
	recorderType := callInfo.RecorderType
	var wg sync.WaitGroup
	once := sync.Once{}
	for j := 0; j < MaxCh; j++ {
		if rtspClient.recGroups[j] {
			continue
		}
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			crd := callInfo.getCRD(j)
			defer callInfo.updateCRD(j, &crd)
			c := callInfo.getClient(j)
			defer callInfo.updateClient(j, &c)
			rtspState := c.rtspState
			if (radioButtonState == constant.TX_BUTTON_OFF && recorderType == constant.RET_RADIO_TX || radioButtonState == constant.RX_BUTTON_OFF && recorderType == constant.RET_RADIO_RX) && (int(rtspState) <= int(constant.RTSP_STATE_START) || rtspState == constant.RTSP_STATE_DISCONNECT || c.client.IsClose()) {
				if c.client.IsClose() {
					fmt.Println("\n\n\n\nDebug 1000: Client is closed, need to restart\n\n\n")
				}
				if rtspClient.ed137Versions[j] == "ED137C" {
					if recorderType == constant.RET_RADIO_TX {
						crd.Properties.CallRef.Value += ("_PTT_" + utils.CreateRand4Digits())
					} else {
						crd.Properties.CallRef.Value += ("_SQU_" + utils.CreateRand4Digits())
					}
				}
				crd.EnableSetupRadio()
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")
				u, err := c.Start(crd)
				if err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", "channel:", c.ch, "Error Starting:", err)
					c.rtspState = constant.RTSP_STATE_DISCONNECT
					return
				}
				if err := c.AnnounceSetup(u); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending Announce or SETUP request:", err)
					c.CloseByErr()
					return
				}
				if err := c.SetParameter(u, crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending SET_PARAMETER request:", err)
					c.CloseByErr()
					return
				}

			} else if (radioButtonState == constant.TX_BUTTON_ON && recorderType == constant.RET_RADIO_TX || radioButtonState == constant.RX_BUTTON_ON && recorderType == constant.RET_RADIO_RX) && (rtspState == constant.RTSP_STATE_SETUP || rtspState == constant.RTSP_STATE_PAUSE) && !c.client.IsClose() {
				defer once.Do(func() {
					callInfo.doRecordRTP(true)
				})
				if recorderType == constant.RET_RADIO_TX {
					if crd.Operations.PTT_Type == "" || crd.Operations.PTT_Type == "0" {
						crd.Operations.PTT.Value = "1"
					} else {
						crd.Operations.PTT.Value = crd.Operations.PTT_Type
					}
				} else {
					crd.Operations.SQU.Value = "1"
				}
				crd.EnableConfirmRadio()
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")
				if err := c.Record(crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending RECORD request:", err)
					c.CloseByErr()
					return
				}

			} else if (radioButtonState == constant.TX_BUTTON_OFF && recorderType == constant.RET_RADIO_TX || radioButtonState == constant.RX_BUTTON_OFF && recorderType == constant.RET_RADIO_RX) && rtspState == constant.RTSP_STATE_RECORD && !c.client.IsClose() {
				defer once.Do(func() {
					callInfo.doRecordRTP(false)
				})
				if recorderType == constant.RET_RADIO_TX {
					crd.Operations.PTT.Value = "0"
				} else {
					crd.Operations.SQU.Value = "0"
				}
				crd.EnablePauseRadio()
				crd.Properties.DisconnectCause = models.CRDAttribute{Value: strconv.Itoa(int(GetDisconnectCause(crd.Properties.SipDisconnectCause.Value, nil)))}
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")
				if err := c.Pause(crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", "channel:", c.ch, "Error sending PAUSE request:", err)
					c.CloseByErr()
					return
				}

			} else if (radioButtonState == constant.TX_BUTTON_ON && recorderType == constant.RET_RADIO_TX || radioButtonState == constant.RX_BUTTON_ON && recorderType == constant.RET_RADIO_RX) && (int(rtspState) <= int(constant.RTSP_STATE_START) || rtspState == constant.RTSP_STATE_DISCONNECT || c.client.IsClose()) {
				defer once.Do(func() {
					callInfo.doRecordRTP(true)
				})
				if rtspClient.ed137Versions[j] == "ED137C" {
					if recorderType == constant.RET_RADIO_TX {
						crd.Properties.CallRef.Value += ("_PTT_" + utils.CreateRand4Digits())
					} else {
						crd.Properties.CallRef.Value += ("_SQU_" + utils.CreateRand4Digits())
					}
				}

				crd.EnableSetupRadio()
				crdByt, _ := xml.MarshalIndent(crd, "", "    ")

				u, err := c.Start(crd)
				if err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error Starting:", err)
					c.rtspState = constant.RTSP_STATE_DISCONNECT
					return
				}
				if err := c.AnnounceSetup(u); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending ANNOUNCE or SETUP request:", err)
					c.CloseByErr()
					return
				}
				if err := c.SetParameter(u, crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending SET_PARAMETER request:", err)
					c.CloseByErr()
					return
				}
				if recorderType == constant.RET_RADIO_TX {
					if crd.Operations.PTT_Type == "" || crd.Operations.PTT_Type == "0" {
						crd.Operations.PTT.Value = "1"
					} else {
						crd.Operations.PTT.Value = crd.Operations.PTT_Type
					}
				} else {
					crd.Operations.SQU.Value = "1"
				}

				crd.EnableConfirmRadio()
				crdByt, _ = xml.MarshalIndent(crd, "", "    ")
				if err = c.Record(crd, crdByt); err != nil {
					rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending RECORD request:", err)
					c.CloseByErr()
					return
				}

			} else if (radioButtonState == constant.BUTTON_INVALID && recorderType == constant.RET_RADIO_TX || radioButtonState == constant.BUTTON_INVALID && recorderType == constant.RET_RADIO_RX) && int(rtspState) > int(constant.RTSP_STATE_NULL) && int(rtspState) < int(constant.RTSP_STATE_DISCONNECT) && !c.client.IsClose() {
				defer once.Do(func() {
					callInfo.doRecordRTP(false)
				})
				crd.EnableDisconnectRadio()
				c.CloseByNormal(&crd)
			}

		}(j)
	}
	wg.Wait()
}

func (callInfo *CallInfo) doOnCallState(callState constant.CallState) {
	rtspClient := GetRTSPClient()
	MaxCh := rtspClient.MaxCh
	var wg sync.WaitGroup
	once := sync.Once{}
	if callInfo.RecorderType == constant.RET_PHONE {
		for j := 0; j < MaxCh; j++ {
			if rtspClient.recGroups[j] {
				continue
			}
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				crd := callInfo.getCRD(j)
				defer callInfo.updateCRD(j, &crd)
				c := callInfo.getClient(j)
				defer callInfo.updateClient(j, &c)
				rtspState := c.rtspState
				switch callState {
				case constant.PJSIP_INV_STATE_CALLING, constant.PJSIP_INV_STATE_INCOMING:
					if rtspState == constant.RTSP_STATE_NULL {
						crd.EnableSetupPhone()
						crdByt, _ := xml.MarshalIndent(crd, "", "    ")
						u, err := c.Start(crd)
						if err != nil {
							rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error Starting:", err)
							c.rtspState = constant.RTSP_STATE_DISCONNECT
							return
						}
						if err := c.AnnounceSetup(u); err != nil {
							rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending ANNOUNCE or SETUP request:", err)
							c.CloseByErr()
							return
						}
						if err := c.SetParameter(u, crd, crdByt); err != nil {
							rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending SET_PARAMETER request:", err)
							c.CloseByErr()
							return
						}
					}
				case constant.PJSIP_INV_STATE_CONFIRMED:
					defer once.Do(func() {
						callInfo.doRecordRTP(true)
					})
					if rtspState == constant.RTSP_STATE_SETUP {
						crd.EnableConfirmPhone()
						crdByt, _ := xml.MarshalIndent(crd, "", "    ")

						if err := c.Record(crd, crdByt); err != nil {
							rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending Record request:", err)
							c.CloseByErr()
							return
						}

					}
				case constant.PJSIP_INV_STATE_DISCONNECTED:
					defer once.Do(func() {
						callInfo.doRecordRTP(false)
					})
					if rtspState != constant.RTSP_STATE_DISCONNECT && rtspState != constant.RTSP_STATE_NULL {
						crd.Operations.Enabled = false
						crd.EnableDisconnectPhone()
						c.CloseByNormal(&crd)
					}
				}
			}(j)
		}
		wg.Wait()
	}
}

func (callInfo *CallInfo) doOnCallMediaState(mediaState constant.CallMediaState) {
	rtspClient := GetRTSPClient()
	MaxCh := rtspClient.MaxCh

	var wg sync.WaitGroup
	once := sync.Once{}
	for j := 0; j < MaxCh; j++ {
		if rtspClient.recGroups[j] {
			continue
		}
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			crd := callInfo.getCRD(j)
			defer callInfo.updateCRD(j, &crd)
			c := callInfo.getClient(j)
			defer callInfo.updateClient(j, &c)
			rtspState := c.rtspState
			switch mediaState {
			case constant.PJSUA_CALL_MEDIA_LOCAL_HOLD, constant.PJSUA_CALL_MEDIA_REMOTE_HOLD:
				if rtspState == constant.RTSP_STATE_RECORD {
					// defer once.Do(func() {
					// 	callInfo.doRecordRTP(false)
					// })
					crd.Operations.Enabled = true
					crd.EnablePausePhone()
					direction, _ := strconv.Atoi(crd.Properties.Direction.Value)
					if mediaState == constant.PJSUA_CALL_MEDIA_LOCAL_HOLD && direction == int(constant.INCOMING) || mediaState == constant.PJSUA_CALL_MEDIA_REMOTE_HOLD && direction == int(constant.OUTGOING) {
						crd.Operations.HOLD.Value = strconv.Itoa(int(constant.HOLD_CALLED_PARTY))
					} else if mediaState == constant.PJSUA_CALL_MEDIA_LOCAL_HOLD && direction == int(constant.OUTGOING) || mediaState == constant.PJSUA_CALL_MEDIA_REMOTE_HOLD && direction == int(constant.INCOMING) {
						crd.Operations.HOLD.Value = strconv.Itoa(int(constant.HOLD_CALLING_PARTY))
					}
					crdByt, _ := xml.MarshalIndent(crd, "", "    ")
					if err := c.Pause(crd, crdByt); err != nil {
						rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending PAUSE request:", err)
						c.CloseByErr()
						return
					}
				} else {
					direction, _ := strconv.Atoi(crd.Properties.Direction.Value)
					if mediaState == constant.PJSUA_CALL_MEDIA_LOCAL_HOLD && direction == int(constant.INCOMING) || mediaState == constant.PJSUA_CALL_MEDIA_REMOTE_HOLD && direction == int(constant.OUTGOING) {
						crd.Operations.HOLD.Value = strconv.Itoa(int(constant.HOLD_CALLED_PARTY))
					} else if mediaState == constant.PJSUA_CALL_MEDIA_LOCAL_HOLD && direction == int(constant.OUTGOING) || mediaState == constant.PJSUA_CALL_MEDIA_REMOTE_HOLD && direction == int(constant.INCOMING) {
						crd.Operations.HOLD.Value = strconv.Itoa(int(constant.HOLD_CALLING_PARTY))
					}
					crdByt, _ := xml.MarshalIndent(crd, "", "    ")
					if err := c.SetParameter(nil, crd, crdByt); err != nil {
						rtspClient.LogDebug("name", c.Name, "recorderType:", "channel:", c.ch, "Error sending SET_PARAMETER request:", err)
						c.CloseByErr()
						return
					}

				}

			case constant.PJSUA_CALL_MEDIA_ACTIVE:
				if rtspState == constant.RTSP_STATE_PAUSE {
					defer once.Do(func() {
						callInfo.doRecordRTP(true)
					})
					crd.EnablePausePhone()
					crd.Operations.Enabled = true
					crd.Operations.HOLD.Value = strconv.Itoa(int(constant.HOLD_OFF))
					crdByt, _ := xml.MarshalIndent(crd, "", "    ")
					if err := c.Record(crd, crdByt); err != nil {
						rtspClient.LogDebug("name", c.Name, "recorderType:", int(c.RecorderType), "channel:", c.ch, "Error sending Record request:", err)
						c.CloseByErr()
						return
					}

				}
			}
		}(j)
	}
	wg.Wait()
}

func (callInfo *CallInfo) sendRTPInner() {
	defer callInfo.wg.Done() // Signal completion when the function exits
	rtspClient.LogDebug("Starting sendRTPInner for name:", callInfo.Name)

	intervalDuration := 200 * time.Millisecond
	interval := gortsplib.EmptyTimer()
	readTimeDuration := 30 * time.Millisecond
	var listenConn *net.UDPConn = nil
	listenPort := 0
	isRecord := false
	rtpInfo := RTPInfo{}

	defer func() {
		if listenConn != nil {
			_ = listenConn.Close()
			rtspClient.LogDebug("Closed listenConn for name:", callInfo.Name)
		}
	}()

	for {
		select {
		case <-callInfo.chDone:
			rtspClient.LogDebug("Received done signal in sendRTPInner for name:", callInfo.Name)
			return

		case newlistenPort := <-callInfo.chUpdateListenConn:
			if listenConn != nil {
				_ = listenConn.Close()
			}
			var err error
			listenConn, err = utils.CreateListenServer("127.0.0.1:" + strconv.Itoa(newlistenPort))
			if err != nil {
				rtspClient.LogDebug("Failed to create listen server for name:", callInfo.Name, "Error:", err)
				return
			}
			listenPort = newlistenPort
			rtspClient.LogDebug("Updated listen port to", listenPort, "for name:", callInfo.Name)

		case newIsRecord := <-callInfo.chRecordRTP:
			if newIsRecord {
				interval = time.NewTimer(intervalDuration)
				rtpInfo = RTPInfo{}
			} else {
				interval = gortsplib.EmptyTimer()
			}
			isRecord = newIsRecord
			rtspClient.LogDebug("Recording state changed to", isRecord, "for name:", callInfo.Name)

		case <-interval.C:
			if !callInfo.atLeastChannelRecord() {
				break
			}
			listPkt := []rtp.Packet{}
			callInfo.readRTPPacket(&listPkt, readTimeDuration, &listenConn, listenPort)
			callInfo.sendRTPPacket(&listPkt, &rtpInfo)
			interval = time.NewTimer(intervalDuration)
		}
	}
}

func (callInfo *CallInfo) readRTPPacket(listPkt *[]rtp.Packet, readTimeDuration time.Duration, listenConn **net.UDPConn, listenPort int) {
	startTime := time.Now()
	if *listenConn == nil {
		if listenPort == 0 {
			return
		}
		if *listenConn, _ = utils.CreateListenServer("127.0.0.1:" + strconv.Itoa(listenPort)); *listenConn == nil {
			return
		}
	}
	if err := (*listenConn).SetReadDeadline(startTime.Add(readTimeDuration)); err != nil {
		return
	}
	for {
		buf := make([]byte, 1024)
		n, _, err := (*listenConn).ReadFromUDP(buf)
		if err != nil {
			return
		}
		if !utils.IsRTPPacket(buf) {
			break
		}
		var pkt rtp.Packet
		if err = pkt.Unmarshal(buf[:n]); err != nil {
			break
		}
		if pkt.Timestamp == uint32(0) || len(pkt.Payload) == 0 {
			continue
		}
		*listPkt = append(*listPkt, pkt)
	}
}

func (callInfo *CallInfo) sendRTPPacket(listPkt *[]rtp.Packet, rtpInfo *RTPInfo) {
	if len(*listPkt) == 0 {
		return
	}
	MaxCh := rtspClient.MaxCh
	newListPkt := []rtp.Packet{}

	*listPkt = callInfo.mergePktWithSameSsrc(*listPkt, rtpInfo)
	for _, pkt := range *listPkt {
		if utils.ConvertCodec(&pkt, rtspClient.codec) {
			newListPkt = append(newListPkt, pkt)
		}
	}
	*listPkt = newListPkt
	var wg sync.WaitGroup
	for j := 0; j < MaxCh; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			c := callInfo.getClientIfExist(j)
			if c.rtspState != constant.RTSP_STATE_RECORD && (c.RecorderType != constant.RET_PHONE || c.rtspState != constant.RTSP_STATE_PAUSE) {
				return
			}
			for _, pkt := range *listPkt {
				if err := c.SendRTPPacket(rtspClient.desc.Medias[0], pkt); err != nil {
					return
				}
			}
		}(j)
	}
	wg.Wait()
}

func (callInfo *CallInfo) mergePktWithSameSsrc(listPkt []rtp.Packet, rtpInfo *RTPInfo) []rtp.Packet {
	if callInfo.RecorderType != constant.RET_RADIO_RX {
		return listPkt
	}
	curSSRC := listPkt[0].SSRC
	var lastActTs, lastDisTs uint32
	var segPloadLength int
	var lastActSeq uint16
	if curSSRC == rtpInfo.lastSSRC {
		lastActTs = rtpInfo.lastActTs
		lastDisTs = rtpInfo.lastDisTs
		segPloadLength = rtpInfo.SegPloadLength
		lastActSeq = rtpInfo.lastActSeq
	} else {
		lastActTs = 0
		lastDisTs = 0
		for _, pkt := range listPkt {
			if len(pkt.Payload) != 0 {
				segPloadLength = len(pkt.Payload)
				break
			}
		}
		lastActSeq = 0
	}
	if segPloadLength == 0 {
		return []rtp.Packet{}
	}
	newListPkt := []rtp.Packet{}
	for _, pkt := range listPkt {
		if lastDisTs == 0 {
			newListPkt = append(newListPkt, pkt)
			lastDisTs = pkt.Timestamp
		} else if pkt.Timestamp == lastDisTs {
			pkt.Timestamp = lastActTs + uint32(segPloadLength)/2
			pkt.SequenceNumber = lastActSeq + 1
			newListPkt = append(newListPkt, pkt)
		} else {
			lastDisTs = pkt.Timestamp
			if pkt.SequenceNumber != lastActSeq+1 {
				pkt.SequenceNumber = lastActSeq + uint16((pkt.Timestamp-lastActTs)*2/uint32(segPloadLength))
			}
			newListPkt = append(newListPkt, pkt)
		}
		lastActTs = pkt.Timestamp
		lastActSeq = pkt.SequenceNumber
		if segPloadLength != len(pkt.Payload) && len(pkt.Payload) != 0 {
			segPloadLength = len(pkt.Payload)
		}
	}
	rtpInfo.lastSSRC = curSSRC
	rtpInfo.lastActTs = lastActTs
	rtpInfo.lastDisTs = lastDisTs
	rtpInfo.SegPloadLength = segPloadLength
	rtpInfo.lastActSeq = lastActSeq
	return newListPkt
}

func (c *Client) SendRTPPacket(media *description.Media, pkt rtp.Packet) error {
	return c.client.WritePacketRTP(rtspClient.desc.Medias[0], &pkt)
}

func (callInfo *CallInfo) runInner() {
	once := sync.Once{}
	rtspClient.LogDebug("Starting runInner for call name:", callInfo.Name)

	// Start handleInner and sendRTPInner
	callInfo.wg.Add(2)
	go callInfo.handleInner()
	go callInfo.sendRTPInner()

	// Ensure cleanup when the function exits
	defer func() {
		rtspClient.LogDebug("Stopping runInner for call name:", callInfo.Name)
		callInfo.cleanupResources()
	}()

	for {
		select {
		case <-callInfo.chDone:
			rtspClient.LogDebug("Received done signal for call name:", callInfo.Name)
			return

		case callStateInfo := <-callInfo.chCallStateInfo:
			callInfo.SetCRD(callStateInfo.crdMsg, callStateInfo.crdMsgId)
			callInfo.doOnCallState(callStateInfo.state)
			if callStateInfo.state == constant.PJSIP_INV_STATE_DISCONNECTED {
				once.Do(func() {
					close(callInfo.chDone)
				})
			}
		case radioButtonStateInfo := <-callInfo.chRadioButtonStateInfo:
			callInfo.SetCRD(radioButtonStateInfo.crdMsg, radioButtonStateInfo.crdMsgId)
			callInfo.doOnRadioState(radioButtonStateInfo.state)
			if radioButtonStateInfo.state == constant.BUTTON_INVALID {
				once.Do(func() {
					close(callInfo.chDone)
				})
			}
		case mediaStateInfo := <-callInfo.chCallMediaStateInfo:
			callInfo.SetCRD(mediaStateInfo.crdMsg, mediaStateInfo.crdMsgId)
			callInfo.doOnCallMediaState(mediaStateInfo.state)
		case briefStateInfo := <-callInfo.chBriefStateInfo:
			callInfo.SetCRD(briefStateInfo.crdMsg, briefStateInfo.crdMsgId)
			callInfo.doOnBriefState(briefStateInfo.state)
			if briefStateInfo.state == constant.BRIEF_FALSE {
				once.Do(func() {
					close(callInfo.chDone)
				})
			}
		case groupStateInfo := <-callInfo.chGroupStateInfo:
			callInfo.SetCRD(groupStateInfo.crdMsg, groupStateInfo.crdMsgId)
			callInfo.doOnGroupState(groupStateInfo.state)
			if groupStateInfo.state == constant.GROUP_FALSE {
				once.Do(func() {
					close(callInfo.chDone)
				})
			}
		}

	}
}

func (callInfo *CallInfo) cleanupResources() {
	rtspClient := GetRTSPClient()

	// Wait for handleInner and sendRTPInner to finish
	rtspClient.LogDebug("Waiting for handleInner and sendRTPInner to finish for call name", callInfo.Name)
	callInfo.wg.Wait()
	rtspClient.LogDebug("handleInner and sendRTPInner finished for call name:", callInfo.Name)

	// Remove clients and CRDs
	for i := 0; i < rtspClient.MaxCh; i++ {
		cKey := ClientKey{
			CallKey: callInfo.CallKey,
			ch:      i,
		}
		rtspClient.cs.listClient.Remove(cKey)
		rtspClient.crds.listCRD.Remove(cKey)
	}

	// Remove call info
	callInfo.Lock()
	rtspClient.callModel.listCallInfo.Remove(callInfo.CallKey)
	callInfo.Unlock()

	rtspClient.LogDebug("Cleanup completed for call name:", callInfo.Name)
}


func (callInfo *CallInfo) handleInner() {
	defer callInfo.wg.Done() // Signal completion when the function exits
	rtspClient := GetRTSPClient()
	rtspClient.LogDebug("Starting handleInner for call name:", callInfo.Name)

	wakeupTime := time.After(1000 * time.Second)
	lastRadioButtonStateInfo := RadioButtonStateInfo{state: constant.BUTTON_INVALID}
	lastCallMediaStateInfo := CallMediaStateInfo{state: constant.PJSUA_CALL_MEDIA_NONE}
	lastPutRadioButtonState := constant.BUTTON_INVALID

	for {
		select {
		case <-callInfo.chDone:
			rtspClient.LogDebug("Received done signal in handleInner for call name:", callInfo.Name)
			return

		case <-callInfo.goSleep:
			switch callInfo.RecorderType {
			case constant.RET_PHONE:
				wakeupTime = time.After(5 * time.Second)
			case constant.RET_RADIO_TX, constant.RET_RADIO_RX:
				wakeupTime = time.After(5 * time.Second)
			}

		case <-wakeupTime:
			callInfo.setSleep(false)
			if lastRadioButtonStateInfo.state != constant.BUTTON_INVALID {
				callInfo.chRadioButtonStateInfo <- lastRadioButtonStateInfo
				lastPutRadioButtonState = lastRadioButtonStateInfo.state
				lastRadioButtonStateInfo = RadioButtonStateInfo{state: constant.BUTTON_INVALID}
			} else if lastCallMediaStateInfo.state != constant.PJSUA_CALL_MEDIA_NONE {
				callInfo.chCallMediaStateInfo <- lastCallMediaStateInfo
				lastCallMediaStateInfo = CallMediaStateInfo{state: constant.PJSUA_CALL_MEDIA_NONE}
			}

		case lastCallMediaStateInfo = <-callInfo.chLastCallMediaStateInfo:
			if !callInfo.isSleep() && (lastCallMediaStateInfo.state == constant.PJSUA_CALL_MEDIA_LOCAL_HOLD || lastCallMediaStateInfo.state == constant.PJSUA_CALL_MEDIA_REMOTE_HOLD) {
				select {
				case callInfo.goSleep <- true:
					callInfo.setSleep(true)
				default:
				}
			}

		case lastRadioButtonStateInfo = <-callInfo.chLastRadioButtonStateInfo:
			if !callInfo.isSleep() && (lastRadioButtonStateInfo.state == constant.TX_BUTTON_OFF || lastRadioButtonStateInfo.state == constant.RX_BUTTON_OFF) {
				if lastPutRadioButtonState == constant.BUTTON_INVALID {
					callInfo.chRadioButtonStateInfo <- lastRadioButtonStateInfo
					lastPutRadioButtonState = lastRadioButtonStateInfo.state
				} else {
					select {
					case callInfo.goSleep <- true:
						callInfo.setSleep(true)
					default:
					}
				}
			}
		}
	}
}
