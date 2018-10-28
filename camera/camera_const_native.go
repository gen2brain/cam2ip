// +build native

package camera

import (
	"github.com/korandiz/v4l"
)

// Property identifiers.
const (
	PropBrightness    = v4l.CtrlBrightness
	PropContrast      = v4l.CtrlContrast
	PropSaturation    = v4l.CtrlSaturation
	PropHue           = v4l.CtrlHue
	PropGain          = v4l.CtrlGain
	PropExposure      = v4l.CtrlExposure
	PropWhiteBalanceU = v4l.CtrlWhiteBalance
	PropSharpness     = v4l.CtrlSharpness
	PropWhiteBalanceV = v4l.CtrlDoWhiteBalance
	PropBacklight     = v4l.CtrlBacklightCompensation
)
