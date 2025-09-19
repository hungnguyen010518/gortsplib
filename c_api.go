package main

/*
#cgo CXXFLAGS: -std=c++11
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"dvrs.lib/RTSPClient/constant"
	"dvrs.lib/RTSPClient/handlers"
	"fmt"
	"os"
)

var GitVersion string

//export GetGitVersion
func GetGitVersion() *C.char {
	return C.CString(GitVersion)
}

//export OnBriefState
func OnBriefState(statusC C.int, nameC *C.char, crdMsgC *C.char, crdMsgIdC *C.char, listenPortC C.int, nameSize C.int, crdMsgSize C.int, crdMsgIdSize C.int) {
	rtspClient := handlers.GetRTSPClient()
	if rtspClient.GetReloadState() != constant.NON_RELOAD || rtspClient.NumNonGroupCh == 0 {
		return
	}
	name := C.GoStringN(nameC, nameSize)
	key := handlers.CallKey{
		Name:         name,
		RecorderType: constant.RET_BRIEF,
	}
	callInfo, blockState := rtspClient.GetCallInfo(key)
	if blockState != constant.NON_BLOCK {
		return
	}
	crdMsg := C.GoStringN(crdMsgC, crdMsgSize)
	crdMsgId := C.GoStringN(crdMsgIdC, crdMsgIdSize)
	listenPort := int(listenPortC)
	status := int(statusC)

	if listenPort != 0 && listenPort != callInfo.ListenPort {
		callInfo.UpdatelistenPort(listenPort)
	}
	callInfo.HandleBriefState(constant.BriefState(status), crdMsg, crdMsgId)
}

//export OnGroupState
func OnGroupState(recorderTypeC C.int, statusC C.int, crdMsgC *C.char, crdMsgIdC *C.char, listenPortC C.int, crdMsgSize C.int, crdMsgIdSize C.int) {
	rtspClient := handlers.GetRTSPClient()
	if rtspClient.GetReloadState() != constant.NON_RELOAD {
		return
	}
	recorderType := int(recorderTypeC)
	if recorderType != int(constant.RET_AMBIENT) && rtspClient.NumGroupCh == 0 {
		return
	}
	key := handlers.CallKey{
		RecorderType: constant.RecorderType(recorderType),
	}
	callInfo, blockState := rtspClient.GetCallInfo(key)
	if blockState != constant.NON_BLOCK {
		return
	}
	crdMsg := C.GoStringN(crdMsgC, crdMsgSize)
	crdMsgId := C.GoStringN(crdMsgIdC, crdMsgIdSize)
	listenPort := int(listenPortC)
	status := int(statusC)

	if listenPort != 0 && listenPort != callInfo.ListenPort {
		callInfo.UpdatelistenPort(listenPort)
	}
	callInfo.HandleGroupState(constant.GroupState(status), crdMsg, crdMsgId)
}

//export OnRadioState
func OnRadioState(sipTypeC, radioButtonStateC C.int, nameC *C.char,
	crdMsgC *C.char, crdMsgIdC *C.char, listenPortC C.int,
	nameSize C.int, crdMsgSize C.int, crdMsgIdSize C.int) {
	rtspClient := handlers.GetRTSPClient()
	if rtspClient.GetReloadState() != constant.NON_RELOAD || rtspClient.NumNonGroupCh == 0 {
		return
	}
	name := C.GoStringN(nameC, nameSize)
	recorderType := int(sipTypeC)
	key := handlers.CallKey{
		Name:         name,
		RecorderType: constant.RecorderType(recorderType),
	}
	callInfo, blockState := rtspClient.GetCallInfo(key)
	if blockState != constant.NON_BLOCK {
		return
	}
	radioButtonState := int(radioButtonStateC)
	crdMsg := C.GoStringN(crdMsgC, crdMsgSize)
	crdMsgId := C.GoStringN(crdMsgIdC, crdMsgIdSize)
	listenPort := int(listenPortC)

	if listenPort != 0 && listenPort != callInfo.ListenPort {
		callInfo.UpdatelistenPort(listenPort)
	}
	callInfo.HandleRadioButtonState(constant.RadioButtonState(radioButtonState), crdMsg, crdMsgId)
}

//export OnCallState
func OnCallState(callStateC C.int, sipTypeC C.int, nameC *C.char,
	crdMsgC *C.char, crdMsgIdC *C.char, listenPortC C.int,
	nameSizeC C.int, crdMsgSizeC C.int, crdMsgIdSizeC C.int) {
	rtspClient := handlers.GetRTSPClient()
	if rtspClient.GetReloadState() != constant.NON_RELOAD || rtspClient.NumNonGroupCh == 0 {
		return
	}
	recorderType := int(sipTypeC)
	name := C.GoStringN(nameC, nameSizeC)
	key := handlers.CallKey{
		Name:         name,
		RecorderType: constant.RecorderType(recorderType),
	}
	callInfo, blockState := rtspClient.GetCallInfo(key)
	if blockState != constant.NON_BLOCK {
		return
	}
	callState := int(callStateC)
	crdMsg := C.GoStringN(crdMsgC, crdMsgSizeC)
	crdMsgId := C.GoStringN(crdMsgIdC, crdMsgIdSizeC)
	listenPort := int(listenPortC)

	if listenPort != 0 && listenPort != callInfo.ListenPort {
		callInfo.UpdatelistenPort(listenPort)
	}
	callInfo.HandleCallState(constant.CallState(callState), crdMsg, crdMsgId)
}

//export OnCallMediaState
func OnCallMediaState(mediaStateC C.int, nameC *C.char, crdMsgC *C.char, crdMsgIdC *C.char,
	nameSize C.int, crdMsgSizeC C.int, crdMsgIdSizeC C.int) {
	rtspClient := handlers.GetRTSPClient()
	if rtspClient.GetReloadState() != constant.NON_RELOAD || rtspClient.NumNonGroupCh == 0 {
		return
	}
	name := C.GoStringN(nameC, nameSize)
	recorderType := constant.RET_PHONE
	key := handlers.CallKey{
		Name:         name,
		RecorderType: recorderType,
	}
	callInfo, blockState := rtspClient.GetCallInfo(key)
	if blockState != constant.NON_BLOCK {
		return
	}
	mediaState := int(mediaStateC)
	crdMsg := C.GoStringN(crdMsgC, crdMsgSizeC)
	crdMsgId := C.GoStringN(crdMsgIdC, crdMsgIdSizeC)
	callInfo.HandleCallMediaState(constant.CallMediaState(mediaState), crdMsg, crdMsgId)
}

//export LoadRecConfig
func LoadRecConfig() {
	rtspClient := handlers.GetRTSPClient()

	// read rec.cfg
	recData, err := os.ReadFile(string(constant.REC_CFG_FILE))
	if err != nil {
		rtspClient.LogInfo(fmt.Sprintf("Could not load %s: %v", constant.REC_CFG_FILE, err))
		recData, err = os.ReadFile(string(constant.ALT_REC_CFG_FILE))
		if err != nil {
			rtspClient.LogInfo(fmt.Sprintf("Could not load %s: %v", constant.ALT_REC_CFG_FILE, err))
		}
	}

	// read device_system.cfg
	devSysData, err := os.ReadFile(string(constant.DEV_SYS_CFG_FILE))
	if err != nil {
		rtspClient.LogInfo(fmt.Sprintf("Could not load %s: %v", constant.DEV_SYS_CFG_FILE, err))
		devSysData, err = os.ReadFile(string(constant.ALT_DEV_SYS_CFG_FILE))
		if err != nil {
			rtspClient.LogInfo(fmt.Sprintf("Could not load %s: %v", constant.ALT_DEV_SYS_CFG_FILE, err))
		}
	}
	rtspClient.ReloadMutex.Lock()
	defer rtspClient.ReloadMutex.Unlock()
	if rtspClient.ReloadState != constant.NON_RELOAD {
		saveRecCfg := handlers.GetSaveCfg()
		saveRecCfg.Reset()
		saveRecCfg.LoadRecFileConfig(recData)
		saveRecCfg.LoadDevSysFileConfig(devSysData)
		saveRecCfg.CheckDupConfig()
		rtspClient.LogInfo(saveRecCfg.String())
	} else {
		rtspClient.LoadRecFileConfig(recData)
		rtspClient.LoadDevSysFileConfig(devSysData)
		rtspClient.CheckDupConfig()
		rtspClient.LogInfo(rtspClient.String())
	}
	rtspClient.LogDebug("Load rec.cfg successfully")
}

//export StopAllCall
func StopAllCall() {
	rtspClient := handlers.GetRTSPClient()
	rtspClient.StopAllCall()
}

func main() {}
