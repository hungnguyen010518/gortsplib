package handlers

import (
	"encoding/xml"
	"reflect"
	"strconv"
	"strings"

	"dvrs.lib/RTSPClient/constant"
	"dvrs.lib/RTSPClient/models"
	"dvrs.lib/RTSPClient/utils"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/liberrors"
)

type CRD struct {
	XMLName    xml.Name      `xml:"call-record-data"`
	Disabled   bool          `xml:"-"`
	VCSUser    string        `xml:"-"`
	Value      string        `xml:"connref,attr"`
	Properties CRDProperties `xml:"properties"`
	Operations CRDOperations `xml:"operations"`
}

type CRDProperties struct {
	XMLName            xml.Name            `xml:"properties"`
	Disabled           bool                `xml:"-"`
	Vnd                models.CRDAttribute `xml:"property"`
	CallingNr          models.CRDAttribute `xml:"property"`
	CalledNr           models.CRDAttribute `xml:"property"`
	ConnectedNr        models.CRDAttribute `xml:"property"`
	CallType           models.CRDAttribute `xml:"property"`
	ClientId           models.CRDAttribute `xml:"property"`
	ClientType         models.CRDAttribute `xml:"property"`
	ConnectedTime      models.CRDAttribute `xml:"property"`
	ConnectTime        models.CRDAttribute `xml:"property"`
	Direction          models.CRDAttribute `xml:"property"`
	SipDisconnectCause models.CRDAttribute `xml:"property"`
	DisconnectCause    models.CRDAttribute `xml:"property"`
	DisconnectTime     models.CRDAttribute `xml:"property"`
	Priority           models.CRDAttribute `xml:"property"`
	SetupTime          models.CRDAttribute `xml:"property"`
	FrequencyID        models.CRDAttribute `xml:"property"`
	CallRef            models.CRDAttribute `xml:"property"`
	AlertNr            models.CRDAttribute `xml:"property"`
	AlertTime          models.CRDAttribute `xml:"property"`
}

type CRDOperations struct {
	XMLName                   xml.Name            `xml:"operations"`
	Enabled                   bool                `xml:"-"`
	HOLD                      models.SubOperation `xml:"operation"`
	PTT                       models.SubOperation `xml:"operation"`
	SQU                       models.SubOperation `xml:"operation"`
	RadioAccessMode           models.SubOperation `xml:"operation"`
	BSS_Quality_Index         models.CRDAttribute `xml:"operation"`
	Simultaneous_Transmission models.CRDAttribute `xml:"operation"`
	R2S                       models.SubOperation `xml:"operation"`
	R2S_TLV                   models.SubOperation `xml:"operation"`
	VOTING                    models.CRDAttribute `xml:"operation"`
	VcsDicomR2S               models.CRDAttribute `xml:"operation"`
	FrequencyID               models.SubOperation `xml:"operation"`
	PTT_Type                  string              `xml:"-"`
}

func (crd CRD) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Check if disabled
	if crd.Disabled {
		return nil // Skip marshalling if disabled
	}

	// Marshalling logic if not disabled
	start.Name.Local = "call-record-data"
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "connref"}, Value: crd.Value})
	e.EncodeToken(start)

	// Marshal Properties and Operations
	e.EncodeElement(crd.Properties, xml.StartElement{Name: xml.Name{Local: "properties"}})
	e.EncodeElement(crd.Operations, xml.StartElement{Name: xml.Name{Local: "operations"}})

	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func (cp CRDProperties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if cp.Disabled {
		return nil
	}
	start.Name.Local = "properties"
	e.EncodeToken(start)
	marshalCRDAttributes(e, reflect.ValueOf(cp), "property")
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

// Custom marshaling for CRDOperations
func (co CRDOperations) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if !co.Enabled {
		return nil
	}
	start.Name.Local = "operations"
	e.EncodeToken(start)
	marshalCRDAttributes(e, reflect.ValueOf(co), "operation")
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func marshalCRDAttributes(e *xml.Encoder, v reflect.Value, name string) {
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if typ.Field(i).Name == "XMLName" {
			// Skip XMLName field
			continue
		}
		// Special handling for HoldOperation
		if subOp, ok := field.Interface().(models.SubOperation); ok {
			if subOp.Disabled || subOp.Value == "" || subOp.Time == "" {
				continue
			}
			fieldName := typ.Field(i).Name
			startElement := xml.StartElement{
				Name: xml.Name{Local: name},
				Attr: []xml.Attr{
					{Name: xml.Name{Local: "name"}, Value: fieldName},
					{Name: xml.Name{Local: "time"}, Value: subOp.Time},
				},
			}
			e.EncodeElement(subOp.Value, startElement)
		} else if attr, ok := field.Interface().(models.CRDAttribute); ok {
			// Check if the attribute is enabled for marshaling
			if attr.Disabled || attr.Value == "" {
				continue
			}
			attr.Name = strings.Replace(typ.Field(i).Name, "_", " ", -1)
			if attr.Name == "SipDisconnectCause" {
				continue
			} else if attr.Name == "Vnd" {
				attr.Name = "vnd.Dicom"
			}
			e.EncodeElement(attr, xml.StartElement{Name: xml.Name{Local: name}})
		}
	}
}

func (crd *CRD) SetCRDInner(crdMsg string, crdMsgId string, ed137Version string, sipType constant.RecorderType) {
	// if ed137Version == "ED137B" && sipType == constant.RET_RADIO_RX {
	// 	return
	// }
	if sipType == constant.RET_RADIO_TX || sipType == constant.RET_RADIO_RX {
		crd.Operations.Enabled = true
	}

	crdPara := strings.Split(crdMsg, ",")
	crdParaId := strings.Split(crdMsgId, ",")

	for i, v := range crdParaId {
		switch vInt, _ := strconv.Atoi(v); vInt {
		case int(constant.VCS_USER_ID):
			if crd.VCSUser == "" {
				crd.VCSUser = crdPara[i]
			}
			if crd.Properties.Vnd.Value == "" {
				switch sipType {
				case constant.RET_RADIO_TX:
					crd.Properties.Vnd.Value += "Radio Selection = TX"
				case constant.RET_RADIO_RX:
					crd.Properties.Vnd.Value += "Radio Selection = RX"
				}
			} else {
				if sipType == constant.RET_RADIO_TX && !strings.Contains(crd.Properties.Vnd.Value, "Radio Selection") {
					crd.Properties.Vnd.Value += ", Radio Selection = TX"
				} else if sipType == constant.RET_RADIO_RX && !strings.Contains(crd.Properties.Vnd.Value, "Radio Selection") {
					crd.Properties.Vnd.Value += ", Radio Selection = RX"
				}
			}
		case int(constant.ENDPT_ID_ID):
			if crd.Properties.Vnd.Value == "" {
				crd.Properties.Vnd.Value += ("TGW Port = " + crdPara[i])
			} else {
				crd.Properties.Vnd.Value += (", TGW Port = " + crdPara[i])
			}
		case int(constant.CLIENT_TYPE_ID):
			crd.Properties.ClientType.Value = crdPara[i]
		case int(constant.DESC_ID):
			if crdPara[i] == "" {
				continue
			} else {
				if crd.Properties.Vnd.Value == "" {
					crd.Properties.Vnd.Value += crdPara[i]
				} else if strings.Contains(crd.Properties.Vnd.Value, "desc") {
					continue
				} else {
					crd.Properties.Vnd.Value += (", desc = " + crdPara[i])
				}
			}
		case int(constant.GROUP_NAME_ID):
			if crdPara[i] == "" {
				continue
			} else {
				if crd.Properties.Vnd.Value == "" {
					crd.Properties.Vnd.Value += crdPara[i]
				} else if strings.Contains(crd.Properties.Vnd.Value, "group name") {
					continue
				} else {
					crd.Properties.Vnd.Value += (", group name = " + crdPara[i])
				}
			}
		case int(constant.ALERT_NR_ID):
			crd.Properties.AlertNr.Value = strings.TrimSuffix(crdPara[i], ";ob")
		case int(constant.ALERT_TIME_ID):
			crd.Properties.AlertTime.Value = crdPara[i]
		case int(constant.CALLING_NR_ID):
			crd.Properties.CallingNr = models.CRDAttribute{Value: strings.TrimSuffix(crdPara[i], ";ob")}
		case int(constant.CALLED_NR_ID):
			crd.Properties.CalledNr = models.CRDAttribute{Value: strings.TrimSuffix(crdPara[i], ";ob")}
			crd.Properties.ConnectedNr = models.CRDAttribute{Value: crd.Properties.CalledNr.Value}
		case int(constant.CLIENT_ID_ID):
			crd.Properties.ClientId = models.CRDAttribute{Value: strings.TrimSuffix(crdPara[i], ";ob")}
			if ed137Version == "ED137C" {
				crd.Properties.ClientType = models.CRDAttribute{Value: "CWP"}
			}
		case int(constant.CALL_REF_ID):
			if ed137Version == "ED137C" || sipType == constant.RET_PHONE {
				crd.Properties.CallRef = models.CRDAttribute{Value: crdPara[i]}
			}
		case int(constant.CONNECT_TIME_ID):
			switch sipType {
			case constant.RET_PHONE:
				crd.Properties.ConnectTime = models.CRDAttribute{Value: crdPara[i]}
				crd.Properties.ConnectedTime = models.CRDAttribute{Value: crdPara[i]}
			case constant.RET_RADIO_TX:
				crd.Operations.PTT.Time = crdPara[i]
				crd.Operations.FrequencyID.Time = crdPara[i]
			case constant.RET_RADIO_RX:
				crd.Properties.ConnectTime = models.CRDAttribute{Value: crdPara[i]}
				crd.Operations.SQU.Time = crdPara[i]
				crd.Operations.FrequencyID.Time = crdPara[i]
			default:
				crd.Properties.ConnectTime.Value = crdPara[i]
			}
		case int(constant.SETUP_TIME_ID):
			crd.Properties.SetupTime = models.CRDAttribute{Value: crdPara[i]}
			if sipType == constant.RET_PHONE {
				continue
			} else if ed137Version == "ED137C" {
				crd.Operations.RadioAccessMode.Time = crdPara[i]
				crd.Properties.ConnectTime.Value = crdPara[i]
				crd.Operations.R2S.Time = crdPara[i]
				crd.Operations.FrequencyID.Time = crdPara[i]
			}
		case int(constant.HOLD_TIME_ID):
			switch sipType {
			case constant.RET_PHONE:
				crd.Operations.HOLD.Time = crdPara[i]
			case constant.RET_RADIO_TX:
				crd.Operations.PTT.Time = crdPara[i]
			case constant.RET_RADIO_RX:
				crd.Operations.SQU.Time = crdPara[i]
			}
		case int(constant.DISCONNECT_TIME_ID):
			crd.Properties.DisconnectTime = models.CRDAttribute{Value: crdPara[i]}
		case int(constant.CALL_TYPE_ID):
			if ed137Version == "ED137C" || sipType == constant.RET_PHONE {
				crd.Properties.CallType = models.CRDAttribute{Value: crdPara[i]}
				if strings.Contains(crdPara[i], "monitoring") {
					crd.Disabled = true
				}
			}
		case int(constant.DIRECTION_ID):
			crd.Properties.Direction = models.CRDAttribute{Value: crdPara[i]}
		case int(constant.SIP_DISCONNECT_CAUSE_ID):
			crd.Properties.SipDisconnectCause = models.CRDAttribute{Value: crdPara[i]}
		case int(constant.PRIORITY_ID):
			if strings.ToUpper(crdPara[i]) == "NORMAL" {
				crd.Properties.Priority = models.CRDAttribute{Value: "3"}
			} else if strings.ToUpper(crdPara[i]) == "EMERGENCY" {
				crd.Properties.Priority = models.CRDAttribute{Value: "1"}
			} else if strings.ToUpper(crdPara[i]) == "URGENT" {
				crd.Properties.Priority = models.CRDAttribute{Value: "2"}
			} else {
				crd.Properties.Priority = models.CRDAttribute{Value: "4"}
			}
		case int(constant.FREQUENCY_ID_ID):
			switch ed137Version {
			case "ED137B":
				crd.Properties.FrequencyID.Value = crdPara[i]
			case "ED137C":
				crd.Operations.FrequencyID.Value = crdPara[i]
			}
		case int(constant.RADIO_ACCESS_MODE_ID):
			if sipType == constant.RET_RADIO_RX {
				switch paraInt, _ := strconv.Atoi(crdPara[i]); paraInt {
				case 1:
					crd.Operations.RadioAccessMode.Value = "1"
				case 2:
					crd.Operations.RadioAccessMode.Value = "2"
				case 3:
					crd.Operations.RadioAccessMode.Value = "3"
				default:
					crd.Operations.RadioAccessMode.Value = "0"
				}
			}
		case int(constant.R2S_ID):
			if ed137Version == "ED137B" || sipType == constant.RET_RADIO_TX {
				continue
			} else {
				crd.Operations.R2S.Value = "Rx=" + crdPara[i]
			}
		case int(constant.PTT_TYPE_ID):
			if sipType == constant.RET_RADIO_TX {
				crd.Operations.PTT_Type = crdPara[i]
			}
		}
	}
	if crd.Value == "" {
		switch crd.Properties.Direction.Value {
		case "", "0":
			crd.Value = utils.CreateRandConref()
		case "1":
			split := strings.Split(crd.Properties.CalledNr.Value, "@")
			if len(split) > 1 {
				crd.Value = utils.CreateRandConref() + "@" + split[1]
			} else {
				crd.Value = utils.CreateRandConref() + "@" + split[0]
			}
		default:
			split := strings.Split(crd.Properties.CallingNr.Value, "@")
			if len(split) > 1 {
				crd.Value = utils.CreateRandConref() + "@" + split[1]
			} else {
				crd.Value = utils.CreateRandConref() + "@" + split[0]
			}
		}
	}
}

func GetDisconnectCause(sipDisconnectCause string, clientErr error) constant.Q931Cause {
	if vInt, _ := strconv.Atoi(sipDisconnectCause); sipDisconnectCause == "" || vInt == int(constant.PJSIP_SC_OK) {
		if clientErr == nil {
			return constant.NORMAL_CLEAR
		} else if v, ok := clientErr.(liberrors.ErrClientBadStatusCode); ok {
			switch v.Code {
			case base.StatusNotFound:
				return constant.UNALLOCATED
			case base.StatusBadGateway:
				return constant.NO_ROUTE_DEST
			case base.StatusNotAcceptable:
				return constant.CHANNEL_UNACCEPT
			case base.StatusRequestTimeout:
				return constant.NO_USER_RESP
			case base.StatusServiceUnavailable:
				return constant.SERV_OPT_UNAVAIL
			case base.StatusOK:
				return constant.NORMAL_CLEAR
			default:
				return constant.NORMAL_UNSPEC
			}
		} else {
			return constant.NORMAL_UNSPEC
		}
	} else if vInt, _ := strconv.Atoi(sipDisconnectCause); vInt != int(constant.PJSIP_SC_OK) {
		switch vInt {
		case int(constant.PJSIP_SC_NOT_FOUND):
			return constant.UNALLOCATED
		case int(constant.PJSIP_SC_BAD_GATEWAY):
			return constant.NO_ROUTE_DEST
		case int(constant.PJSIP_SC_NOT_ACCEPTABLE):
			return constant.CHANNEL_UNACCEPT
		case int(constant.PJSIP_SC_BUSY_HERE):
			return constant.USER_BUSY
		case int(constant.PJSIP_SC_REQUEST_TIMEOUT):
			return constant.NO_USER_RESP
		case int(constant.PJSIP_SC_SERVICE_UNAVAILABLE):
			return constant.SERV_OPT_UNAVAIL
		default:
			return constant.NORMAL_UNSPEC
		}
	} else {
		return constant.NORMAL_UNSPEC
	}
}

func (crd *CRD) EnableDisconnectPhone() {
	crd.DisableAllProperty()
	crd.Properties.Vnd.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Properties.DisconnectCause.Disabled = false
	crd.Properties.DisconnectTime.Disabled = false
	crd.Properties.ClientId.Disabled = false
}

func (crd *CRD) EnableSetupPhone() {
	crd.DisableAllProperty()
	crd.Properties.Vnd.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Properties.CallingNr.Disabled = false
	crd.Properties.CalledNr.Disabled = false
	crd.Properties.SetupTime.Disabled = false
	crd.Properties.ClientId.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Properties.Direction.Disabled = false
	crd.Properties.CallRef.Disabled = false
	crd.Properties.CallType.Disabled = false
	crd.Properties.Priority.Disabled = false
}

func (crd *CRD) EnableConfirmPhone() {
	crd.DisableAllProperty()
	crd.Properties.Vnd.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Properties.ConnectedNr.Disabled = false
	crd.Properties.ConnectTime.Disabled = false
	crd.Properties.ClientId.Disabled = false
}

func (crd *CRD) EnablePausePhone() {
	crd.DisableAllProperty()
	crd.Properties.Vnd.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Properties.ConnectedNr.Disabled = false
	crd.Properties.ConnectedTime.Disabled = false
	crd.Properties.ClientId.Disabled = false
}

func (crd *CRD) DisableAllProperty() {
	v := reflect.ValueOf(&crd.Properties).Elem()
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if fieldName == "XMLName" {
			continue
		}
		fieldValue := v.Field(i)

		// Handle both models.CRDAttribute and HoldOperation
		if fieldValue.CanSet() && fieldValue.Kind() == reflect.Struct {
			if attr, ok := fieldValue.Interface().(models.CRDAttribute); ok {
				attr.Disabled = true
				fieldValue.Set(reflect.ValueOf(attr))
			}
		}
	}
}

func (crd *CRD) EnableSetupRadio() {
	crd.DisableAllProperty()
	crd.Properties.Vnd.Disabled = false
	crd.Properties.ClientType.Disabled = false
	crd.Operations.Enabled = true
	crd.Properties.CallingNr.Disabled = false
	crd.Properties.CalledNr.Disabled = false
	crd.Properties.ClientId.Disabled = false
	crd.Properties.Direction.Disabled = false
	crd.Properties.Priority.Disabled = false
	crd.Properties.SetupTime.Disabled = false
	crd.Properties.ConnectTime.Disabled = false
	crd.Operations.PTT.Disabled = true
	crd.Operations.SQU.Disabled = true
	crd.Operations.RadioAccessMode.Disabled = false
	crd.Operations.R2S.Disabled = false
}

func (crd *CRD) EnableConfirmRadio() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Operations.Enabled = true
	crd.Properties.Vnd.Disabled = false
	crd.Properties.CallingNr.Disabled = false
	crd.Properties.CalledNr.Disabled = false
	crd.Properties.ClientId.Disabled = false
	crd.Properties.Direction.Disabled = false
	crd.Properties.Priority.Disabled = false
	crd.Properties.SetupTime.Disabled = false
	crd.Properties.ConnectTime.Disabled = false
	crd.Properties.FrequencyID.Disabled = false
	crd.Operations.PTT.Disabled = false
	crd.Operations.SQU.Disabled = false
	crd.Operations.RadioAccessMode.Disabled = true
	crd.Operations.R2S.Disabled = true
}

func (crd *CRD) EnablePauseRadio() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.Vnd.Disabled = false
	crd.Properties.Priority.Disabled = false
	crd.Properties.ClientId.Disabled = false
	crd.Properties.FrequencyID.Disabled = false
	crd.Operations.Enabled = true
	crd.Operations.PTT.Disabled = false
	crd.Operations.SQU.Disabled = false
	crd.Operations.RadioAccessMode.Disabled = true
	crd.Operations.R2S.Disabled = true
}

func (crd *CRD) EnableDisconnectRadio() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.Vnd.Disabled = false
	crd.Operations.Enabled = false
	crd.Properties.DisconnectCause.Disabled = false
	crd.Properties.DisconnectTime.Disabled = false
	crd.Properties.ClientId.Disabled = false
}

func (crd *CRD) EnableStartRadio() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Operations.Enabled = false
	crd.Properties.AlertNr.Disabled = false
	crd.Properties.AlertTime.Disabled = false
}

func (crd *CRD) EnableEarlyDisconnectRadio() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Operations.Enabled = true
	crd.Operations.FrequencyID.Disabled = false
	crd.Properties.SetupTime.Disabled = false
	crd.Properties.Direction.Disabled = false
	crd.Properties.CallingNr.Disabled = false
	crd.Properties.CalledNr.Disabled = false
	crd.Properties.ClientId.Disabled = false
	crd.Properties.DisconnectCause.Disabled = false
}

func (crd *CRD) EnableConnectBrief() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.ConnectTime.Disabled = false
}

func (crd *CRD) EnableDisconnectBrief() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.DisconnectTime.Disabled = false
	crd.Properties.DisconnectCause.Disabled = false
}

func (crd *CRD) EnableConnectGroup() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.ConnectTime.Disabled = false
}

func (crd *CRD) EnableDisconnectGroup() {
	crd.DisableAllProperty()
	crd.Properties.ClientType.Disabled = false
	crd.Properties.DisconnectTime.Disabled = false
	crd.Properties.DisconnectCause.Disabled = false
}
