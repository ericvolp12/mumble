package main

import (
	"encoding/binary"
	"flag"
	"fmt"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/mattetti/audio"
	"github.com/moutend/go-equalizer/pkg/equalizer"
)

func main() {

	// NOTE: All of the below fields are required for this example to work correctly.
	var (
		Token     = flag.String("t", "", "Discord bot token.")
		GuildID   = flag.String("g", "", "Guild ID")
		ChannelID = flag.String("c", "", "Channel ID")
		err       error
	)
	flag.Parse()

	// Connect to Discord
	discord, err := discordgo.New("Bot " + *Token)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open Websocket
	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Connect to voice channel.
	// NOTE: Setting mute to false, deaf to true.
	dgv, err := discord.ChannelVoiceJoin(*GuildID, *ChannelID, false, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Starts echo
	echo(dgv)

	// Close connections
	dgv.Close()
	discord.Close()

	return
}

// Takes inbound audio and sends it right back out.
func echo(v *discordgo.VoiceConnection) {

	recv := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v, recv)

	send := make(chan []int16, 2)
	go dgvoice.SendPCM(v, send)

	v.Speaking(true)
	defer v.Speaking(false)

	for {

		p, ok := <-recv
		if !ok {
			return
		}

		rawInts := []int{}

		for _, i := range p.PCM {
			rawInts = append(rawInts, int(i))
		}

		buf := audio.NewPCMIntBuffer(rawInts, &audio.Format{
			NumChannels: 1,
			SampleRate:  48000,
			BitDepth:    16,
			Endianness:  binary.LittleEndian,
		})

		buf.SwitchPrimaryType(audio.Float)

		outs := []float64{}

		bpf := equalizer.NewLowPass(48000, 80, 0.5)
		for _, i := range buf.AsFloat64s() {
			outs = append(outs, bpf.Apply(i))
		}

		buf.Floats = outs

		buf.SwitchPrimaryType(audio.Integer)

		send <- buf.AsInt16s()
	}
}
