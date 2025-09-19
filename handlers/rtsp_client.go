package handlers

import (
	"sync"
	"time"

	"dvrs.lib/RTSPClient/constant"
	"dvrs.lib/RTSPClient/utils"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type CheckReload struct {
	ReloadState constant.ReloadState
	ReloadMutex *sync.RWMutex
}

type CRDModel struct {
	listCRD cmap.ConcurrentMap[ClientKey, CRD]
}

type ClientModel struct {
	listClient cmap.ConcurrentMap[ClientKey, Client]
}

type CallModel struct {
	listCallInfo cmap.ConcurrentMap[CallKey, CallInfo]
}

type RTSPClient struct {
	*Config
	CheckReload
	callModel CallModel
	cs        ClientModel
	crds      CRDModel
	utils.Logger
}

var rtspClient *RTSPClient

func GetRTSPClient() *RTSPClient {
	if rtspClient == nil {
		saveCfg = NewCfg()
		rtspClient = &RTSPClient{
			callModel: CallModel{
				listCallInfo: cmap.NewWithCustomShardingFunction[CallKey, CallInfo](CallKey.Hash),
			},
			Config: NewCfg(),
			CheckReload: CheckReload{
				ReloadState: constant.NON_RELOAD,
				ReloadMutex: &sync.RWMutex{},
			},
			cs: ClientModel{
				listClient: cmap.NewWithCustomShardingFunction[ClientKey, Client](ClientKey.Hash),
			},
			crds:   CRDModel{listCRD: cmap.NewWithCustomShardingFunction[ClientKey, CRD](ClientKey.Hash)},
			Logger: utils.CreateZapLogger(),
		}

	}
	return rtspClient
}
func (rtspClient *RTSPClient) GetCallInfo(Key CallKey) (CallInfo, constant.BlockState) {
	if rtspClient.callModel.listCallInfo.Has(Key) {
		callInfo, _ := rtspClient.callModel.listCallInfo.Get(Key)
		return callInfo, callInfo.blockState
	} else {
		callInfo := CallInfo{
			CallKey: Key,
			RTPClient: RTPClient{
				chUpdateListenConn: make(chan int, 2),
				chRecordRTP:        make(chan bool, 1),
			},
			ThreadHandle: ThreadHandle{
				chDone: make(chan bool),
				wg:     &sync.WaitGroup{},
			},
			EventQueue: EventQueue{
				chBriefStateInfo:           make(chan BriefStateInfo, 10),
				chGroupStateInfo:         make(chan GroupStateInfo, 10),
				chCallMediaStateInfo:       make(chan CallMediaStateInfo, 10),
				chCallStateInfo:            make(chan CallStateInfo, 10),
				chRadioButtonStateInfo:     make(chan RadioButtonStateInfo, 20),
				chLastCallMediaStateInfo:   make(chan CallMediaStateInfo, 1),
				chLastRadioButtonStateInfo: make(chan RadioButtonStateInfo, 1),
			},
			SleepHandle: SleepHandle{
				goSleep: make(chan bool, 1),
				sleep:   false,
			},
		}
		callInfo.Lock()
		rtspClient.callModel.listCallInfo.Set(Key, callInfo)
		callInfo.Unlock()
		go callInfo.runInner()
		return callInfo, constant.NON_BLOCK
	}
}

func (rtspClient *RTSPClient) GetCallInfoIfExist(Key CallKey) (CallInfo, bool) {
	if rtspClient.callModel.listCallInfo.Has(Key) {
		callInfo, _ := rtspClient.callModel.listCallInfo.Get(Key)
		return callInfo, true
	}
	return CallInfo{
		CallKey: Key,
	}, false
}

func (rtspClient *RTSPClient) updateCallInfoSkipLock(Key CallKey, callInfo *CallInfo) {
	rtspClient.callModel.listCallInfo.Set(Key, *callInfo)
}

func (rtspClient *RTSPClient) GetReloadState() constant.ReloadState {
	rtspClient.ReloadMutex.RLock()
	defer rtspClient.ReloadMutex.RUnlock()
	return rtspClient.ReloadState
}

func (rtspClient *RTSPClient) SetReloadState(ReloadState constant.ReloadState) {
	rtspClient.ReloadMutex.Lock()
	defer rtspClient.ReloadMutex.Unlock()
	rtspClient.ReloadState = ReloadState
}

func (rtspClient *RTSPClient) waitForRealeaseCall() {
	maxWaitDuration := 4 * time.Second
	intervalDuration := 50 * time.Millisecond
	maxWait := time.After(maxWaitDuration)
	interval := time.After(0)
	defer rtspClient.updateConfigAfterReload()
	for {
		select {
		case <-maxWait:
			rtspClient.LogWarn("Time waiting for all callInfo to release has been expired, pottentally have future errors")
			return
		case <-interval:
			if rtspClient.GetReloadState() == constant.NORMAL_RELOAD && rtspClient.cs.listClient.IsEmpty() && rtspClient.crds.listCRD.IsEmpty() && rtspClient.callModel.listCallInfo.IsEmpty() {
				rtspClient.LogDebug("All calls have been release successfully")
				return
			}
			interval = time.After(intervalDuration)
		}
	}
}

func (rtspClient *RTSPClient) updateConfigAfterReload() {
	rtspClient.ReloadMutex.Lock()
	defer rtspClient.ReloadMutex.Unlock()
	rtspClient.Reset()
	rtspClient.Config = saveCfg.Copy()
	rtspClient.CheckDupConfig()
	saveCfg.Reset()
	rtspClient.ReloadState = constant.NON_RELOAD
}

func (rtspClient *RTSPClient) StopAllCall() {
	rtspClient.SetReloadState(constant.NORMAL_RELOAD)
	for it := range rtspClient.callModel.listCallInfo.IterBuffered() {
		callInfo := it.Val
		if callInfo.blockState != constant.NON_BLOCK {
			continue
		}
		switch callInfo.RecorderType {
		case constant.RET_PHONE:
			select {
			case callInfo.chCallStateInfo <- CallStateInfo{state: constant.PJSIP_INV_STATE_DISCONNECTED}:
				callInfo.setBlockState(constant.NORMAL_BLOCK)
			default:
			}
		case constant.RET_RADIO_TX, constant.RET_RADIO_RX:
			select {
			case callInfo.chRadioButtonStateInfo <- RadioButtonStateInfo{state: constant.BUTTON_INVALID}:
				callInfo.setBlockState(constant.NORMAL_BLOCK)
			default:
			}
		case constant.RET_BRIEF:
			select {
			case callInfo.chBriefStateInfo <- BriefStateInfo{state: constant.BRIEF_FALSE}:
				callInfo.setBlockState(constant.NORMAL_BLOCK)
			default:
			}
		case constant.RET_AMBIENT:
			select {
			case callInfo.chGroupStateInfo <- GroupStateInfo{state: constant.GROUP_FALSE}:
				callInfo.setBlockState(constant.NORMAL_BLOCK)
			default:
			}
		}
	}
	go rtspClient.waitForRealeaseCall()
}
