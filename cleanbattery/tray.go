/*

  Copyright (c) 2012-2014, Muhammed Uluyol <uluyol0@gmail.com>
  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions are met:

   - Redistributions of source code must retain the above copyright notice,
     this list of conditions and the following disclaimer.

   - Redistributions in binary form must reproduce the above copyright notice,
     this list of conditions and the following disclaimer in the documentation
     and/or other materials provided with the distribution.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.

*/

package main

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	timeout            = 3000
	newline            = byte(10)
	statusCharging    = "Charging"
	statusDischarging = "Discharging"
)

var (
	statusIcon *gtk.GtkStatusIcon
	full        int64
)

func main() {
	gtk.Init(&os.Args)
	glib.SetApplicationName("zzcleanbattery")

	buf, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/energy_full")
	if err != nil {
		panic(err)
	}
	str := string(bytes.Split(buf, []byte{newline})[0])
	full, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}

	statusIcon = gtk.StatusIcon()
	statusIcon.SetTitle("zzcleanbattery")

	updateIcon()

	glib.TimeoutAdd(timeout, updateIcon)
	gtk.Main()
}

func updateIcon() bool {

	var (
		hours   int64
		minutes int64
		seconds int64
		pfull   int64
		rate    int64
		now     int64
		status  string
		text    string
	)

	buf, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/energy_now")
	if err != nil {
		panic(err)
	}
	str := string(bytes.Split(buf, []byte{newline})[0])
	now, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}

	buf, err = ioutil.ReadFile("/sys/class/power_supply/BAT0/power_now")
	if err != nil {
		panic(err)
	}
	str = string(bytes.Split(buf, []byte{newline})[0])
	rate, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}

	buf, err = ioutil.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		panic(err)
	}
	status = string(bytes.Split(buf, []byte{newline})[0])

	pfull = now * 100 / full

	if rate > 0 {
		switch status {
		case statusCharging:
			seconds = 3600 * (full - now) / rate
		case statusDischarging:
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

	statusIcon.SetTooltipText(text)
	statusIcon.SetFromIconName(getIconName(status, pfull))
	return true
}

func getIconName(status string, pfull int64) string {
	if status == statusDischarging {
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
			return "battery_third_fouth"
		default:
			return "battery_full"
		}
	} else if status == statusCharging {
		return "battery_charged"
	}
	return "battery_plugged"
}
