package constant

type RTSPState int

const (
	RTSP_STATE_NULL RTSPState = iota
	RTSP_STATE_START
	RTSP_STATE_ANNOUNCE
	RTSP_STATE_SETUP
	RTSP_STATE_RECORD
	RTSP_STATE_PAUSE
	RTSP_STATE_DISCONNECT
)

type BriefState int

const (
	BRIEF_NONE BriefState = iota - 1
	BRIEF_FALSE
	BRIEF_TRUE
)

type GroupState int

const (
	GROUP_NONE GroupState = iota - 1
	GROUP_FALSE
	GROUP_TRUE
)

type CallState int

const (
	PJSIP_INV_STATE_NULL         CallState = iota /**< Before INVITE is sent or received  */
	PJSIP_INV_STATE_CALLING                       /**< After INVITE is sent		    */
	PJSIP_INV_STATE_INCOMING                      /**< After INVITE is received.	    */
	PJSIP_INV_STATE_EARLY                         /**< After response with To tag.	    */
	PJSIP_INV_STATE_CONNECTING                    /**< After 2xx is sent/received.	    */
	PJSIP_INV_STATE_CONFIRMED                     /**< After ACK is sent/received.	    */
	PJSIP_INV_STATE_DISCONNECTED                  /**< Session is terminated.		    */
)

type CallMediaState int

const (
	PJSUA_CALL_MEDIA_NONE CallMediaState = iota
	PJSUA_CALL_MEDIA_ACTIVE
	PJSUA_CALL_MEDIA_LOCAL_HOLD
	PJSUA_CALL_MEDIA_REMOTE_HOLD
	PJSUA_CALL_MEDIA_ERROR
)

type RadioButtonState int

const (
	BUTTON_INVALID RadioButtonState = iota
	TX_BUTTON_OFF
	TX_BUTTON_ON
	RX_BUTTON_OFF
	RX_BUTTON_ON
)

type Direction int

const (
	UNKNOWN Direction = iota
	INCOMING
	OUTGOING
)

type Hold int

const (
	HOLD_OFF Hold = iota
	HOLD_CALLING_PARTY
	HOLD_CALLED_PARTY
	HOLD_BOTH
)

type RecorderType int

const (
	RET_INVALID RecorderType = iota - 1
	RET_PHONE
	RET_RADIO_TX
	RET_RADIO_RX
	RET_BRIEF
	RET_AMBIENT
	RET_PHONE_GROUP
	RET_RADIO_GROUP
	RET_BRIEF_GROUP
)

type SetParameterMode string

const (
	NormalMode SetParameterMode = "NORMAL"
	RecordMode SetParameterMode = "RECORD"
	PauseMode  SetParameterMode = "PAUSE"
)

type BlockState int

const (
	NON_BLOCK BlockState = iota
	NORMAL_BLOCK
)
