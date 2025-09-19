package constant

type pjsip_status_code int

const (
	PJSIP_SC_TRYING               pjsip_status_code = 100
	PJSIP_SC_RINGING              pjsip_status_code = 180
	PJSIP_SC_CALL_BEING_FORWARDED pjsip_status_code = 181
	PJSIP_SC_QUEUED               pjsip_status_code = 182
	PJSIP_SC_PROGRESS             pjsip_status_code = 183

	PJSIP_SC_OK       pjsip_status_code = 200
	PJSIP_SC_ACCEPTED pjsip_status_code = 202

	PJSIP_SC_MULTIPLE_CHOICES    pjsip_status_code = 300
	PJSIP_SC_MOVED_PERMANENTLY   pjsip_status_code = 301
	PJSIP_SC_MOVED_TEMPORARILY   pjsip_status_code = 302
	PJSIP_SC_USE_PROXY           pjsip_status_code = 305
	PJSIP_SC_ALTERNATIVE_SERVICE pjsip_status_code = 380

	PJSIP_SC_BAD_REQUEST                   pjsip_status_code = 400
	PJSIP_SC_UNAUTHORIZED                  pjsip_status_code = 401
	PJSIP_SC_PAYMENT_REQUIRED              pjsip_status_code = 402
	PJSIP_SC_FORBIDDEN                     pjsip_status_code = 403
	PJSIP_SC_NOT_FOUND                     pjsip_status_code = 404
	PJSIP_SC_METHOD_NOT_ALLOWED            pjsip_status_code = 405
	PJSIP_SC_NOT_ACCEPTABLE                pjsip_status_code = 406
	PJSIP_SC_PROXY_AUTHENTICATION_REQUIRED pjsip_status_code = 407
	PJSIP_SC_REQUEST_TIMEOUT               pjsip_status_code = 408
	PJSIP_SC_GONE                          pjsip_status_code = 410
	PJSIP_SC_REQUEST_ENTITY_TOO_LARGE      pjsip_status_code = 413
	PJSIP_SC_REQUEST_URI_TOO_LONG          pjsip_status_code = 414
	PJSIP_SC_UNSUPPORTED_MEDIA_TYPE        pjsip_status_code = 415
	PJSIP_SC_UNSUPPORTED_URI_SCHEME        pjsip_status_code = 416
	PJSIP_SC_BAD_EXTENSION                 pjsip_status_code = 420
	PJSIP_SC_EXTENSION_REQUIRED            pjsip_status_code = 421
	PJSIP_SC_SESSION_TIMER_TOO_SMALL       pjsip_status_code = 422
	PJSIP_SC_INTERVAL_TOO_BRIEF            pjsip_status_code = 423
	PJSIP_SC_TEMPORARILY_UNAVAILABLE       pjsip_status_code = 480
	PJSIP_SC_CALL_TSX_DOES_NOT_EXIST       pjsip_status_code = 481
	PJSIP_SC_LOOP_DETECTED                 pjsip_status_code = 482
	PJSIP_SC_TOO_MANY_HOPS                 pjsip_status_code = 483
	PJSIP_SC_ADDRESS_INCOMPLETE            pjsip_status_code = 484
	PJSIP_AC_AMBIGUOUS                     pjsip_status_code = 485
	PJSIP_SC_BUSY_HERE                     pjsip_status_code = 486
	PJSIP_SC_REQUEST_TERMINATED            pjsip_status_code = 487
	PJSIP_SC_NOT_ACCEPTABLE_HERE           pjsip_status_code = 488
	PJSIP_SC_BAD_EVENT                     pjsip_status_code = 489
	PJSIP_SC_REQUEST_UPDATED               pjsip_status_code = 490
	PJSIP_SC_REQUEST_PENDING               pjsip_status_code = 491
	PJSIP_SC_UNDECIPHERABLE                pjsip_status_code = 493

	PJSIP_SC_INTERNAL_SERVER_ERROR pjsip_status_code = 500
	PJSIP_SC_NOT_IMPLEMENTED       pjsip_status_code = 501
	PJSIP_SC_BAD_GATEWAY           pjsip_status_code = 502
	PJSIP_SC_SERVICE_UNAVAILABLE   pjsip_status_code = 503
	PJSIP_SC_SERVER_TIMEOUT        pjsip_status_code = 504
	PJSIP_SC_VERSION_NOT_SUPPORTED pjsip_status_code = 505
	PJSIP_SC_MESSAGE_TOO_LARGE     pjsip_status_code = 513
	PJSIP_SC_PRECONDITION_FAILURE  pjsip_status_code = 580

	PJSIP_SC_BUSY_EVERYWHERE         pjsip_status_code = 600
	PJSIP_SC_DECLINE                 pjsip_status_code = 603
	PJSIP_SC_DOES_NOT_EXIST_ANYWHERE pjsip_status_code = 604
	PJSIP_SC_NOT_ACCEPTABLE_ANYWHERE pjsip_status_code = 606

	// PJSIP_SC_TSX_TIMEOUT = PJSIP_SC_REQUEST_TIMEOUT
	/*PJSIP_SC_TSX_RESOLVE_ERROR = 702,*/
	// PJSIP_SC_TSX_TRANSPORT_ERROR = PJSIP_SC_SERVICE_UNAVAILABLE,

	/* This is not an actual status code, but rather a constant
	 * to force GCC to use 32bit to represent this enum, since
	 * we have a code in PJSUA-LIB that assigns an integer
	 * to this enum (see pjsua_acc_get_info() function).
	 */
	// PJSIP_SC__force_32bit = 0x7FFFFFFF
)

type Q931Cause int

const (
	UNALLOCATED            Q931Cause = 1
	NO_ROUTE_TRANSIT       Q931Cause = 2
	NO_ROUTE_DEST          Q931Cause = 3
	SPECIAL_TONE           Q931Cause = 4
	MISDIALED_PREFIX       Q931Cause = 5
	CHANNEL_UNACCEPT       Q931Cause = 6
	CALL_AWARDED           Q931Cause = 7
	PREFIX_0_RESTRICTED    Q931Cause = 8
	PREFIX_1_RESTRICTED    Q931Cause = 9
	PREFIX_1_REQUIRED      Q931Cause = 10
	MORE_DIGITS            Q931Cause = 11
	NORMAL_CLEAR           Q931Cause = 16
	USER_BUSY              Q931Cause = 17
	NO_USER_RESP           Q931Cause = 18
	USER_ALERT_NO_ANS      Q931Cause = 19
	CALL_REJECTED          Q931Cause = 21
	NUM_CHANGED            Q931Cause = 22
	REV_CHARGE_REJECT      Q931Cause = 23
	CALL_SUSPENDED         Q931Cause = 24
	CALL_RESUMED           Q931Cause = 25
	NON_SELECT_CLEAR       Q931Cause = 26
	DEST_OUT_ORDER         Q931Cause = 27
	INVALID_NUM_FORMAT     Q931Cause = 28
	EKTS_REJECTED          Q931Cause = 29
	STATUS_RESP            Q931Cause = 30
	NORMAL_UNSPEC          Q931Cause = 31
	CIRCUIT_OUT_ORDER      Q931Cause = 33
	NO_CIRCUIT_AVAIL       Q931Cause = 34
	DEST_UNREACHABLE       Q931Cause = 35
	OUT_OF_ORDER           Q931Cause = 36
	DEGRADED_SERVICE       Q931Cause = 37
	NET_OUT_ORDER          Q931Cause = 38
	TRANSIT_DELAY_UNMET    Q931Cause = 39
	THROUGHPUT_UNMET       Q931Cause = 40
	TEMP_FAILURE           Q931Cause = 41
	SWITCH_CONGEST         Q931Cause = 42
	ACCESS_INFO_DISC       Q931Cause = 43
	REQ_CIRCUIT_UNAVAIL    Q931Cause = 44
	PREEMPTED              Q931Cause = 45
	PRIORITY_BLOCKED       Q931Cause = 46
	RESOURCE_UNAVAIL       Q931Cause = 47
	QOS_UNAVAIL            Q931Cause = 49
	REQ_FACILITY_UNSUB     Q931Cause = 50
	REV_CHARGE_DENIED      Q931Cause = 52
	OUTGOING_BARRED        Q931Cause = 53
	OUTGOING_BARRED_CUG    Q931Cause = 54
	INCOMING_BARRED        Q931Cause = 55
	INCOMING_BARRED_CUG    Q931Cause = 56
	CALL_WAIT_UNSUB        Q931Cause = 57
	BEARER_CAP_UNAUTH      Q931Cause = 58
	BEARER_CAP_UNAVAIL     Q931Cause = 59
	SERV_OPT_UNAVAIL       Q931Cause = 63
	BEARER_SVC_UNIMPL      Q931Cause = 65
	CHANNEL_TYPE_UNIMPL    Q931Cause = 66
	TRANSIT_NET_UNIMPL     Q931Cause = 67
	MSG_UNIMPL             Q931Cause = 68
	REQ_FACILITY_UNIMPL    Q931Cause = 69
	RESTRICTED_DIGITAL     Q931Cause = 70
	SERV_OPT_UNIMPL        Q931Cause = 79
	INVALID_CALL_REF       Q931Cause = 81
	ID_CHANNEL_NOT_EXIST   Q931Cause = 82
	SUSPENDED_CALL_EXIST   Q931Cause = 83
	CALL_ID_IN_USE         Q931Cause = 84
	NO_SUSPENDED_CALL      Q931Cause = 85
	CLEARED_CALL_ID        Q931Cause = 86
	USER_NOT_CUG_MEMBER    Q931Cause = 87
	INCOMPATIBLE_DEST      Q931Cause = 88
	NON_EXIST_ABBREV       Q931Cause = 89
	DEST_MISSING           Q931Cause = 90
	INVALID_TRANSIT_SEL    Q931Cause = 91
	INVALID_FACILITY_PARM  Q931Cause = 92
	MANDATORY_INFO_MISSING Q931Cause = 93
	MSG_TYPE_UNIMPL        Q931Cause = 95
	MSG_INCOMPAT_STATE     Q931Cause = 96
	INFO_ELEM_UNIMPL       Q931Cause = 97
	INVALID_INFO_CONTENT   Q931Cause = 98
	MSG_INCOMPAT_CALL      Q931Cause = 99
	RECOVERY_TIMER_EXP     Q931Cause = 100
	PARAM_UNIMPL           Q931Cause = 101
	PROTOCOL_ERROR         Q931Cause = 111
	INTERNETWORKING        Q931Cause = 127
	PROPRIETARY_CODE       Q931Cause = 128
)
