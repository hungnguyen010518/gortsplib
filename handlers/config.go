package handlers

import (
	"regexp"
	"strconv"
	"strings"

	"dvrs.lib/RTSPClient/utils"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

var saveCfg *Config

type Config struct {
	MaxCh           int
	NumNonGroupCh   int
	NumGroupCh      int
	recAddrs        []string
	mediaTransports []string
	keepTimeAlives  []string
	interleaves     []string
	ed137Versions   []string
	recGroups       []bool
	desc            description.Session
	codec           string
}

func NewCfg() *Config {
	cfg := &Config{}
	cfg.desc = description.Session{
		Medias: []*description.Media{{
			Type: description.MediaTypeAudio,
			Formats: []format.Format{&format.G711{
				MULaw: false,
			},
			},
		}},
	}
	cfg.codec = "g711alaw"
	return cfg
}

func GetSaveCfg() *Config {
	if saveCfg == nil {
		saveCfg = NewCfg()
	}
	return saveCfg
}

func (cfg *Config) LoadRecFileConfig(data []byte) {
	var reTrue = regexp.MustCompile(`(true)|(false)`)
	var reEnable = regexp.MustCompile(`(yes)|(no)|(enable)|(disable)|(true)|(false)`)
	var reIP = regexp.MustCompile(`([0-9]+\.){3}[0-9]+`)
	var rePort = regexp.MustCompile(`[0-9]+`)
	var reTime = regexp.MustCompile(`[0-9]+`)
	var reTransport = regexp.MustCompile(`(tcp)|(udp)`)
	var reEd = regexp.MustCompile(`ED137[A-C]{1}`)
	var reCodec = regexp.MustCompile(`g[0-9a-z]+`)

	var wEnable = regexp.MustCompile(`Enable`)
	var wIPAddress = regexp.MustCompile(`rec_ip`)
	var wRemoteRTSPPort = regexp.MustCompile(`rec_port`)
	var wMediaTransport = regexp.MustCompile(`media_transport`)
	var wInterleaved = regexp.MustCompile(`interleaved`)
	var wKeepAliveInterval = regexp.MustCompile(`keep_alive_interval`)
	var wEd137Version = regexp.MustCompile(`ed137_version`)
	var wCodec = regexp.MustCompile(`codec`)
	var wRecGroup = regexp.MustCompile(`rec_group`)

	var enables, recIPs, recPorts, mediaTransports, keepTimeAlives, ed137Versions, interleaves, recGroups []string
	var codec string
	dataStr := utils.RemoveComments(string(data))
	lines := strings.Split(dataStr, "\n")
	for _, line := range lines {
		if wEnable.MatchString(line) {
			// If 'true' or 'false' is explicitly mentioned, use that.
			// Otherwise, assume 'true' (enabled) by default.
			matches := reTrue.FindStringSubmatch(line)
			if len(matches) > 0 {
				enables = append(enables, matches[0])
			} else {
				enables = append(enables, "true")
			}
		} else if wIPAddress.MatchString(line) {
			// Extract IP address. If not found, use default "127.0.0.1"
			matches := reIP.FindStringSubmatch(line)
			if len(matches) > 0 {
				recIPs = append(recIPs, matches[0])
			} else {
				recIPs = append(recIPs, "127.0.0.1")
			}
		} else if wRemoteRTSPPort.MatchString(line) {
			// Extract port number. If not found, use default "8554"
			matches := rePort.FindStringSubmatch(line)
			if len(matches) > 0 {
				recPorts = append(recPorts, matches[0])
			} else {
				recPorts = append(recPorts, "8554")
			}
		} else if wMediaTransport.MatchString(line) {
			// Extract media transport. If not found, use default "udp"
			matches := reTransport.FindStringSubmatch(line)
			if len(matches) > 0 {
				mediaTransports = append(mediaTransports, matches[0])
			} else {
				mediaTransports = append(mediaTransports, "udp")
			}
		} else if wInterleaved.MatchString(line) {
			// Extract interleaved setting. If not found, use default "enable"
			matches := reEnable.FindStringSubmatch(line)
			if len(matches) > 0 {
				interleaves = append(interleaves, matches[0])
			} else {
				interleaves = append(interleaves, "enable")
			}
		} else if wKeepAliveInterval.MatchString(line) {
			// Extract keep-alive interval. If not found, use default "20"
			matches := reTime.FindStringSubmatch(line)
			if len(matches) > 0 {
				keepTimeAlives = append(keepTimeAlives, matches[0])
			} else {
				keepTimeAlives = append(keepTimeAlives, "20")
			}
		} else if wEd137Version.MatchString(line) {
			// Extract ED137 version. If not found, use default "ED137B"
			matches := reEd.FindStringSubmatch(line)
			if len(matches) > 0 {
				ed137Versions = append(ed137Versions, matches[0])
			} else {
				ed137Versions = append(ed137Versions, "ED137B")
			}
		} else if wCodec.MatchString(line) {
			// Extract codec. If not found, use default "g711alaw"
			matches := reCodec.FindStringSubmatch(line)
			if len(matches) > 0 {
				codec = matches[0]
			} else {
				codec = "g711alaw"
			}
		} else if wRecGroup.MatchString(line) {
			// Extract recording group. If not found, use default "default"
			matches := reTrue.FindStringSubmatch(line)
			if len(matches) > 0 {
				recGroups = append(recGroups, matches[0])
			} else {
				recGroups = append(recGroups, "false")
			}
		}
		// Ignore lines that don't match any of the patterns
	}

	if enables == nil {
		// If 'Enable' is not specified, assume all channels are enabled
		MaxCh := len(recIPs)
		for j := 0; j < MaxCh; j++ {
			cfg.recAddrs = append(cfg.recAddrs, recIPs[j]+":"+recPorts[j])
			cfg.mediaTransports = append(cfg.mediaTransports, mediaTransports...)
			cfg.keepTimeAlives = append(cfg.keepTimeAlives, keepTimeAlives...)
			cfg.ed137Versions = append(cfg.ed137Versions, ed137Versions...)
			cfg.interleaves = append(cfg.interleaves, interleaves...)
			cfg.MaxCh++
			if recGroups[j] == "true" {
				cfg.recGroups = append(cfg.recGroups, true)
				cfg.NumGroupCh++
			} else {
				cfg.recGroups = append(cfg.recGroups, false)
				cfg.NumNonGroupCh++
			}

		}
	} else {
		// If 'Enable' is specified, only process channels where it's 'true'
		for j, v := range enables {
			if v == "true" {
				cfg.recAddrs = append(cfg.recAddrs, recIPs[j]+":"+recPorts[j])
				cfg.mediaTransports = append(cfg.mediaTransports, mediaTransports[j])
				cfg.keepTimeAlives = append(cfg.keepTimeAlives, keepTimeAlives[j])
				cfg.ed137Versions = append(cfg.ed137Versions, ed137Versions[j])
				cfg.interleaves = append(cfg.interleaves, interleaves[j])
				cfg.MaxCh++
				if recGroups[j] == "true" {
					cfg.recGroups = append(cfg.recGroups, true)
					cfg.NumGroupCh++
				} else {
					cfg.recGroups = append(cfg.recGroups, false)
					cfg.NumNonGroupCh++
				}
			}
		}
	}
	cfg.codec = codec
	switch codec {
	case "g711alaw":
		cfg.desc = description.Session{
			Medias: []*description.Media{{
				Type: description.MediaTypeAudio,
				Formats: []format.Format{&format.G711{
					MULaw: false,
				},
				},
			}},
		}
	case "g711ulaw":
		cfg.desc = description.Session{
			Medias: []*description.Media{{
				Type: description.MediaTypeAudio,
				Formats: []format.Format{&format.G711{
					MULaw: true,
				},
				},
			}},
		}
	default:
	}
}

func (cfg *Config) LoadDevSysFileConfig(data []byte) {

	listIp := []string{}
	listPort := []string{}

	parsingTMCSServer := false

	dataStr := utils.RemoveComments(string(data))

	// Split the data into lines
	lines := strings.Split(dataStr, "\n")

	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for section start (assuming "tmcs_server {")
		if strings.Contains(line, "tmcs_server") {
			parsingTMCSServer = true
			continue
		}

		if strings.Contains(line, ")") {
			if parsingTMCSServer {
				break
			} else {
				continue
			}
		}

		if strings.Contains(line, "ip_address") {
			ip := utils.ExtractIpAddr(line)
			listIp = append(listIp, ip)
		}

		if strings.Contains(line, "port") {
			port := utils.ExtractPort(line)
			listPort = append(listPort, port)
		}
	}

	if len(listIp) != len(listPort) {
		return
	}
	for i := range listIp {
		ip := listIp[i]
		port := listPort[i]
		if ip == "" || port == "" {
			continue
		}
		cfg.MaxCh++
		cfg.recAddrs = append(cfg.recAddrs, ip+":"+port)
		cfg.mediaTransports = append(cfg.mediaTransports, "udp")
		cfg.ed137Versions = append(cfg.ed137Versions, "ED137B")
		cfg.interleaves = append(cfg.interleaves, "disable")
		cfg.keepTimeAlives = append(cfg.keepTimeAlives, "10")
		cfg.recGroups = append(cfg.recGroups, false)
		cfg.NumNonGroupCh++
	}
}

func (cfg *Config) Copy() *Config {
	return &Config{
		MaxCh:           cfg.MaxCh,
		recAddrs:        append([]string{}, cfg.recAddrs...),
		mediaTransports: append([]string{}, cfg.mediaTransports...),
		keepTimeAlives:  append([]string{}, cfg.keepTimeAlives...),
		interleaves:     append([]string{}, cfg.interleaves...),
		ed137Versions:   append([]string{}, cfg.ed137Versions...),
		recGroups:       append([]bool{}, cfg.recGroups...),
		NumGroupCh:      cfg.NumGroupCh,
		NumNonGroupCh:   cfg.NumNonGroupCh,
		desc:            cfg.desc,
		codec:           cfg.codec,
	}
}

func (cfg *Config) Reset() {
	cfg.MaxCh = 0
	cfg.recAddrs = []string{}
	cfg.mediaTransports = []string{}
	cfg.keepTimeAlives = []string{}
	cfg.interleaves = []string{}
	cfg.ed137Versions = []string{}
	cfg.recGroups = []bool{}
	cfg.NumGroupCh = 0
	cfg.NumNonGroupCh = 0
}

func (cfg *Config) CheckDupConfig() {
	if cfg.MaxCh == 0 {
		return
	}
	type SubConfig struct {
		recAddr        string
		mediaTransport string
		keepTimeAlive  string
		interleave     string
		ed137Version   string
		recGroup       bool
	}
	dupCfgAttrs := make(map[string][]SubConfig)
	for i, addr := range cfg.recAddrs {
		subCfgAttrs := dupCfgAttrs[addr]
		subCfgAttrs = append(subCfgAttrs, SubConfig{
			recAddr:        addr,
			mediaTransport: cfg.mediaTransports[i],
			keepTimeAlive:  cfg.keepTimeAlives[i],
			interleave:     cfg.interleaves[i],
			ed137Version:   cfg.ed137Versions[i],
			recGroup:       cfg.recGroups[i],
		})
		dupCfgAttrs[addr] = subCfgAttrs
	}
	cfg.Reset()
	for _, v := range dupCfgAttrs {
		if len(v) != 0 {
			added := false
			for _, attr := range v {
				if attr.mediaTransport == "udp" {
					cfg.recAddrs = append(cfg.recAddrs, attr.recAddr)
					cfg.mediaTransports = append(cfg.mediaTransports, attr.mediaTransport)
					cfg.keepTimeAlives = append(cfg.keepTimeAlives, attr.keepTimeAlive)
					cfg.interleaves = append(cfg.interleaves, attr.interleave)
					cfg.ed137Versions = append(cfg.ed137Versions, attr.ed137Version)
					cfg.MaxCh++
					if attr.recGroup {
						cfg.recGroups = append(cfg.recGroups, true)
						cfg.NumGroupCh++
					} else {
						cfg.recGroups = append(cfg.recGroups, false)
						cfg.NumNonGroupCh++
					}
					added = true
					break
				}
			}
			if !added {
				cfg.recAddrs = append(cfg.recAddrs, v[0].recAddr)
				cfg.mediaTransports = append(cfg.mediaTransports, v[0].mediaTransport)
				cfg.keepTimeAlives = append(cfg.keepTimeAlives, v[0].keepTimeAlive)
				cfg.interleaves = append(cfg.interleaves, v[0].interleave)
				cfg.ed137Versions = append(cfg.ed137Versions, v[0].ed137Version)
				cfg.MaxCh++
				if v[0].recGroup {
					cfg.recGroups = append(cfg.recGroups, true)
					cfg.NumGroupCh++
				} else {
					cfg.recGroups = append(cfg.recGroups, false)
					cfg.NumNonGroupCh++
				}
			}
		}
	}
}

func (cfg *Config) String() string {
	str := "RTSP Client Config:\n"
	str += "Max Channels: " + strconv.Itoa(cfg.MaxCh) + "\n"
	for i := 0; i < cfg.MaxCh; i++ {
		str += "Channel " + strconv.Itoa(i) + ":\n"
		str += "  Address: " + cfg.recAddrs[i] + "\n"
		str += "  Media Transport: " + cfg.mediaTransports[i] + "\n"
		str += "  Keep Alive Time: " + cfg.keepTimeAlives[i] + "\n"
		str += "  Interleave: " + cfg.interleaves[i] + "\n"
		str += "  ED137 Version: " + cfg.ed137Versions[i] + "\n"
		str += "  Recorder Group: " + strconv.FormatBool(cfg.recGroups[i]) + "\n"
	}
	str += "Number of Group Channels: " + strconv.Itoa(cfg.NumGroupCh) + "\n"
	str += "Number of Non-Group Channels: " + strconv.Itoa(cfg.NumNonGroupCh) + "\n"
	return str
}
