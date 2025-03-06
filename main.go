package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/screen" // This is required to register screen adapter
	"github.com/pion/mediadevices/pkg/prop"

	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/webrtc/v3"
)

func main() {
	screen := flag.Bool("s", false, "share screen instead of camera and mic")
	videoBitrate := flag.Int("b", 1_000_000, "video bitrate in bits per second")
	iceServer := flag.String("i", "stun:stun.l.google.com:19302", "ice server")
	token := flag.String("t", "", "publishing token")
	microphone := flag.String("m", "", "id of the microphone to use (see --list-microphones)")
	camera := flag.String("c", "", "id of the camera to use (see --list-cameras)")
	videoCodec := flag.String("vc", "h264", "video codec vp8|h264")
	flag.BoolFunc("list-cameras", "list available cameras", func(string) error {
		fmt.Print("\nAvailable cameras:\n\n")
		var devices = mediadevices.EnumerateDevices()
		for _, device := range devices {
			if(device.DeviceType == "camera" && device.Label != ""){
				fmt.Println(device.Label)
			}
		}
		os.Exit(0)
		return nil
	})
	flag.BoolFunc("list-microphones", "list available microphones", func(string) error {
		fmt.Print("\nAvailable microphones:\n\n")
		var devices = mediadevices.EnumerateDevices()
		for _, device := range devices {
			if(device.DeviceType == "microphone" && device.Label != ""){
				fmt.Println(device.Label)
			}
		}
		os.Exit(0)
		return nil
	})
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("Invalid number of arguments, pass the publishing url as the first argument")
	}
	var cameraId string
	if *camera != ""{
		var devices = mediadevices.EnumerateDevices()
		log.Println(devices)
		for _, device := range devices {
			if(*camera == device.Label){
				cameraId = device.DeviceID
			}
		}
		if(cameraId == ""){
			log.Fatal("Invalid camera id (use --list-cameras to obtain a valid device)")
		}
	}
	var microphoneId string
	if *microphone != ""{
		var devices = mediadevices.EnumerateDevices()
		for _, device := range devices {
			if(*microphone == device.Label){
				microphoneId = device.DeviceID
			}
		}
		if(microphoneId == ""){
			log.Fatal("Invalid microphone id (use --list-microphones to obtain a valid device)")
		}
	}
	log.Println(microphoneId)
	
	// create a new peer connection
	mediaEngine := webrtc.MediaEngine{}
	whip := NewWHIPClient(flag.Args()[0], *token)

	// configure codec specific parameters
	vpxParams, err := vpx.NewVP8Params()
	if err != nil {
		panic(err)
	}
	vpxParams.BitRate = *videoBitrate

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}

	x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	x264Params.BitRate = *videoBitrate
	x264Params.Preset = x264.PresetUltrafast

	var videoCodecSelector mediadevices.CodecSelectorOption
	if *videoCodec == "vp8" {
		videoCodecSelector = mediadevices.WithVideoEncoders(&vpxParams)
	} else {
		videoCodecSelector = mediadevices.WithVideoEncoders(&x264Params)
	}
	var stream mediadevices.MediaStream

	if *screen {
		codecSelector := mediadevices.NewCodecSelector(videoCodecSelector)
		codecSelector.Populate(&mediaEngine)

		stream, err = mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
			Video: func(constraint *mediadevices.MediaTrackConstraints) {},
			Codec: codecSelector,
		})
		if err != nil {
			log.Fatal("Unexpected error capturing screen. ", err)
		}
	} else  { 
		codecSelector := mediadevices.NewCodecSelector(
			videoCodecSelector,
			mediadevices.WithAudioEncoders(&opusParams),
		)
		codecSelector.Populate(&mediaEngine)
		
		var videoConstraints = func(constraint *mediadevices.MediaTrackConstraints) {
				constraint.Width = prop.Int(1280)
				constraint.Height = prop.Int(720)
			}

		if(cameraId != ""){
			log.Println("Using camera: ", cameraId)
			videoConstraints = func(constraint *mediadevices.MediaTrackConstraints) {
				constraint.DeviceID = prop.String(cameraId)
				constraint.Width = prop.Int(1280)
				constraint.Height = prop.Int(720)
			}
		} 

		var audioConstraints = func(constraint *mediadevices.MediaTrackConstraints) {}
		if(microphoneId != ""){
			log.Println("Using microphone: ", microphoneId)
			audioConstraints = func(constraint *mediadevices.MediaTrackConstraints) {
				constraint.DeviceID = prop.String(microphoneId)
			}
		}

		stream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: videoConstraints,
			Audio: audioConstraints,
			Codec: codecSelector,
		})
		if err != nil {
			log.Fatal("Unexpected error capturing camera source. ", err)
		}
	}

	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{*iceServer},
		},
	}

	whip.Publish(stream, mediaEngine, iceServers, true)

	fmt.Println("Press 'Enter' to finish...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	whip.Close(true)
}
