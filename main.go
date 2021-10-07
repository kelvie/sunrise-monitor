package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

func runcmd(cmdStr string) {
	if cmdStr == "" {
		return
	}
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("Running", cmdStr)
	cmd.Run()
}

func main() {
	// Default is downtown Vancouver
	lat := flag.Float64("lat", 49.28307, "Latitude")
	long := flag.Float64("long", -123.12015, "Longitude")
	onSunrise := flag.String("onSunrise", "echo Sun has risen!", "shell command to run on sunrise. Will also be run on startup if the sun is up.")
	onSunset := flag.String("onSunset", "echo Sun has set!", "shell command to run on sunset. Will also be run onstartup if the sun is down")
	offset := flag.Int("offset", 0, "Offset in minutes to delay the commands -- negative means run before sunset/rise")

	flag.Parse()

	getRiseSet := func(t time.Time) (time.Time, time.Time) {
		srise, sset :=  sunrise.SunriseSunset(*lat, *long, t.Year(), t.Month(), t.Day())
		off := time.Duration(*offset) * time.Minute
		return srise.Add(off).Local(), sset.Add(off).Local()
	}

	now := time.Now()
	srise, sset := getRiseSet(now)

	// It's dark before sunrise and after sunset
	if now.Before(srise) {
		log.Printf("Sun is down (sunrise will be at %v), running sunset command.", srise)
		runcmd(*onSunset)
	} else if now.After(sset) {
		log.Printf("Sun is down (sun set at %v), running sunrise command.", sset)
		runcmd(*onSunset)
	} else {
		log.Printf("Sun is up (sun rose at %v), running sunrise command.", srise)
		runcmd(*onSunrise)
	}

	// This assumes sun doesn't set past midnight, and we don't do wacky shit like change the timezone.
	for {
		now = time.Now()

		// After sunset? We set next sunrise/sunset to tomorrow
		if now.After(sset) {
			tomorrow := now.Add(24 * time.Hour)
			srise, sset = getRiseSet(tomorrow)
		}

		// If it's before sunrise, wait for sunrise, then run the sunrise command
		if now.Before(srise) {
			log.Printf("Waiting for next sunrise at %v.", srise)
			time.Sleep(time.Until(srise))
			log.Printf("Sun has risen, running command.")
			runcmd(*onSunrise)
		}

		// If it's before sunset, wait for sunset and run the sunset command
		if now.Before(sset) {
			log.Printf("Waiting for next sunset at %v.", sset)
			time.Sleep(time.Until(sset))
			log.Printf("Sun has set, running command.")
			runcmd(*onSunset)
		}

		// In case the next loop happens so fast that it's *still* before sunset
		time.Sleep(1 * time.Second)
	}
}
