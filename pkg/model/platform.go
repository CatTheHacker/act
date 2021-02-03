package model

type PlatformEngine string

type Platform struct {
	Platform  string
	UseHost   bool
	Image     string
	Supported bool
}
