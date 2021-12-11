package main

import (
	"flag"
	"fmt"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	dsp "github.com/eripe970/go-dsp-utils"
)

func main() {

	// NOTE: All of the below fields are required for this example to work correctly.
	var (
		Token1     = flag.String("t1", "", "Discord bot 1 token.")
		Token2     = flag.String("t2", "", "Discord bot 2 token.")
		GuildID    = flag.String("g", "", "Guild ID")
		ChannelID1 = flag.String("c1", "", "Channel ID 1")
		ChannelID2 = flag.String("c2", "", "Channel ID 2")
		err        error
	)
	flag.Parse()

	// Initialize Bot 1
	// Connect to Discord
	discord1, err := discordgo.New("Bot " + *Token1)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open Websocket
	err = discord1.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Connect to voice channel.
	// NOTE: Setting mute to false, deaf to true.
	dgv1, err := discord1.ChannelVoiceJoin(*GuildID, *ChannelID1, false, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Initialize Bot 2
	// Connect to Discord
	discord2, err := discordgo.New("Bot " + *Token2)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open Websocket
	err = discord2.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Connect to voice channel.
	// NOTE: Setting mute to false, deaf to true.
	dgv2, err := discord2.ChannelVoiceJoin(*GuildID, *ChannelID2, false, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Starts echo
	echo(dgv1, dgv2)

	// Close connections
	dgv1.Close()
	discord1.Close()

	dgv2.Close()
	discord2.Close()
}

// Takes inbound audio and sends it right back out.
func echo(v1 *discordgo.VoiceConnection, v2 *discordgo.VoiceConnection) {

	recv1 := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v1, recv1)

	send1 := make(chan []int16, 2)
	go dgvoice.SendPCM(v1, send1)

	v1.Speaking(true)
	defer v1.Speaking(false)

	recv2 := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v2, recv2)

	send2 := make(chan []int16, 2)
	go dgvoice.SendPCM(v2, send2)

	v2.Speaking(true)
	defer v2.Speaking(false)

	for {

		p, ok := <-recv1
		if !ok {
			return
		}

		normalizedFloats := []float64{}

		for _, i := range p.PCM {
			newFloat := float64(i) / 32768.0
			if newFloat > 1 {
				newFloat = 1
			}
			if newFloat < -1 {
				newFloat = -1
			}
			normalizedFloats = append(normalizedFloats, newFloat)
		}

		inputSignal := &dsp.Signal{
			SampleRate: 48000,
			Signal:     normalizedFloats,
		}

		signalFiltered, err := inputSignal.LowPassFilter(80)
		if err != nil {
			fmt.Printf("Failed to low-pass: %v", err.Error())
		}

		outs := []int16{}

		for _, i := range signalFiltered.Signal {
			scaled := int16(i * 32768)
			if scaled > 32767 || scaled < -32767 {
				fmt.Printf("Got weird value: %v\n", scaled)
			}
			outs = append(outs, scaled)
		}

		send2 <- outs

	}
}
