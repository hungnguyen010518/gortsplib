package constant

type CfgFile string

const (
	REC_CFG_FILE         CfgFile = "config/rec-config/rec.cfg"
	DEV_SYS_CFG_FILE     CfgFile = "config/system/device_system.cfg"
	ALT_REC_CFG_FILE     CfgFile = "/home/cwp/opconsole/config/rec-config/rec.cfg"
	ALT_DEV_SYS_CFG_FILE CfgFile = "/home/cwp/opconsole/config/system/device_system.cfg"
)

type RecCfg int

const (
	ENABLE_ID RecCfg = iota
	REC_IP_ID
	REC_PORT_ID
	RTSP_TRANSPORT_ID
	MEDIA_TRANSPORT_ID
	INTERLEAVE_ID
	KEEP_ALIVE_TIME_ID
	ED137_VERSION_ID
	CODEC_ID
)

type ReloadState int

const (
	NON_RELOAD ReloadState = iota
	NORMAL_RELOAD
)
