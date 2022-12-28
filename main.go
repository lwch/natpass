package main

import (
	"context"
	"fmt"
	"image"
	"io"
	"time"

	"github.com/lwch/rdesktop"
	"github.com/lwch/runtime"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"

	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"

	// If you don't like x264, you can also use vpx by importing as below
	// "github.com/pion/mediadevices/pkg/codec/vpx" // This is required to use VP8/VP9 video encoder
	// or you can also use openh264 for alternative h264 implementation
	"github.com/pion/mediadevices/pkg/codec/openh264"
	// or if you use a raspberry pi like, you can use mmal for using its hardware encoder
	// "github.com/pion/mediadevices/pkg/codec/mmal"
	// "github.com/pion/mediadevices/pkg/codec/x264" // This is required to use h264 video encoder
	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
)

type screen struct {
	ctx    context.Context
	cancel context.CancelFunc
	cli    *rdesktop.Client
}

func newScreen() *screen {
	cli, err := rdesktop.New()
	runtime.Assert(err)
	return &screen{cli: cli}
}

func (s *screen) Open() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return nil
}

func (s *screen) Close() error {
	s.cancel()
	s.cli.Close()
	return nil
}
func (s *screen) VideoRecord(selectedProp prop.Media) (video.Reader, error) {
	r := video.ReaderFunc(func() (img image.Image, release func(), err error) {
		select {
		case <-s.ctx.Done():
			return nil, nil, io.EOF
		default:
		}

		for {
			img, err = s.cli.Screenshot()
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		release = func() {}
		return
	})
	return r, nil
}

func (s *screen) Properties() []prop.Media {
	size, err := s.cli.Size()
	runtime.Assert(err)
	supportedProp := prop.Media{
		Video: prop.Video{
			Width:       size.X,
			Height:      size.Y,
			FrameFormat: frame.FormatRGBA,
		},
	}
	return []prop.Media{supportedProp}
}

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun1.l.google.com:19302"},
			},
		},
	}
	driver.GetManager().Register(
		newScreen(),
		driver.Info{Label: "screen", DeviceType: driver.Camera, Priority: driver.PriorityNormal},
	)
	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	Decode(MustReadStdin(), &offer)

	// Create a new RTCPeerConnection
	x264Params, err := openh264.NewParams()
	if err != nil {
		panic(err)
	}
	x264Params.BitRate = 1_000_000 // 500kbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {

		},
		Codec: codecSelector,
	})
	if err != nil {
		panic(err)
	}

	for _, track := range s.GetTracks() {
		track.OnEnded(func(err error) {
			fmt.Printf("Track (ID: %s) ended with error: %v\n",
				track.ID(), err)
		})

		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
