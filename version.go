package main

// These variables are injected at build time via -ldflags:
//
//	-X main.AppVersion=0.0.3
//	-X main.BuildDate=29052026
//	-X main.BuildTime=1433
//	-X main.BuildNumber=67a4f8c2
var (
	AppVersion  = "dev"
	BuildDate   = "unknown"
	BuildTime   = "unknown"
	BuildNumber = "0"
)
