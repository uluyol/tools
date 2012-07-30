/*
 *  Copyright (c) 2012, Muhammed Uluyol <uluyol0@gmail.com>
 *  All rights reserved.
 *
 *  Redistribution and use in source and binary forms, with or without
 *  modification, are permitted provided that the following conditions are met:
 *
 *   - Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *
 *   - Redistributions in binary form must reproduce the above copyright notice,
 *     this list of conditions and the following disclaimer in the documentation
 *     and/or other materials provided with the distribution.
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 *  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 *  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 *  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 *  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 *  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *  POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"bytes"
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/glib"
	"os"
)

const (
	TIMEOUT = 3000
	NEWLINE = byte(10)
	STATUS_CHARGING = "Charging"
	STATUS_DISCHARGING = "Discharging"
)

var (
	status_icon *gtk.GtkStatusIcon
	full int64
)

func main() {
	gtk.Init(&os.Args)
	glib.SetApplicationName("zzcleanbattery")

	buf, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/energy_full")
	if err != nil { panic(err) }
	str := string(bytes.Split(buf, []byte{NEWLINE})[0])
	full, err = strconv.ParseInt(str, 10, 64)
	if err != nil { panic(err) }

	status_icon = gtk.StatusIcon()
	status_icon.SetTitle("zzcleanbattery")

	update_icon()

	glib.TimeoutAdd(TIMEOUT, update_icon)
	gtk.Main()
}

func update_icon() bool {

	var (
		hours int64
		minutes int64
		seconds int64
		pfull int64
		rate int64
		now int64
		status string
		text string
	)

	buf, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/energy_now")
	if err != nil { panic(err) }
	str := string(bytes.Split(buf, []byte{NEWLINE})[0])
	now, err = strconv.ParseInt(str, 10, 64)
	if err != nil { panic(err) }

	buf, err = ioutil.ReadFile("/sys/class/power_supply/BAT0/power_now")
	if err != nil { panic(err) }
	str = string(bytes.Split(buf, []byte{NEWLINE})[0])
	rate, err = strconv.ParseInt(str, 10, 64)
	if err != nil { panic(err) }

	buf, err = ioutil.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil { panic(err) }
	status = string(bytes.Split(buf, []byte{NEWLINE})[0])

	pfull = now * 100 / full

	if rate > 0 {
		switch status {
		case STATUS_CHARGING:
			seconds = 3600 * (full - now) / rate
		case STATUS_DISCHARGING:
			seconds = 3600 * now / rate
		default:
			seconds = 0
		}
	} else {
		seconds = 0
	}
	hours = seconds / 3600
	seconds -= hours * 3600
	minutes = seconds / 60
	seconds -= minutes * 60
	if seconds == 0 {
		text = fmt.Sprintf("%s, %d%%", status, pfull)
	} else {
		text = fmt.Sprintf("%s, %d%%, %d:%02d remaining",
		           status,
		           pfull,
		           hours,
		           minutes)
	}

	status_icon.SetTooltipText(text)
	status_icon.SetFromIconName(get_icon_name(status, pfull))
	return true
}

func get_icon_name(status string, pfull int64) string {
	if status == STATUS_DISCHARGING {
		switch {
		case pfull < 10:
			return "battery_empty"
		case pfull < 20:
			return "battery_caution"
		case pfull < 40:
			return "battery_low"
		case pfull < 60:
			return "battery_two_thirds"
		case pfull < 75:
			return "battery_third_fourth"
		default:
			return "battery_full"
		}
	} else if status == STATUS_CHARGING {
		return "battery_charged"
	}
	return "battery_plugged"
}