module dvrs.lib/RTSPClient

go 1.21.5

replace go.uber.org/zap => ./pkg/zap

replace github.com/orcaman/concurrent-map/v2 => ./pkg/concurrent_map

replace github.com/bluenviron/gortsplib/v4 => ./pkg/gortsplib

toolchain go1.21.6

require (
	github.com/bluenviron/gortsplib/v4 v4.0.0-00010101000000-000000000000
	github.com/orcaman/concurrent-map/v2 v2.0.0-00010101000000-000000000000
	github.com/pion/rtp v1.8.3
	go.uber.org/zap v1.0.0
)

require (
	github.com/bluenviron/mediacommon v1.9.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pion/rtcp v1.2.13 // indirect
	github.com/pion/sdp/v3 v3.0.6 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.20.0 // indirect
)

require (
	github.com/pion/randutil v0.1.0 // indirect
	github.com/zaf/g711 v1.4.0
	golang.org/x/sys v0.16.0 // indirect
)
