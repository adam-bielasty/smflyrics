package main

import (
	"bytes"
	"fmt"
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/meta"
	"github.com/gomidi/midi/smf"
	"github.com/gomidi/midi/smf/smfreader"
	"github.com/metakeule/config"
	"os"
)

var (
	cfg = config.MustNew(
		"smflyrics",
		"1.0.0",
		"extracts lyrics from a SMF file, tracks are separated by an empty line",
	)

	argFile = cfg.NewString(
		"file",
		"the SMF file that is read in",
		config.Shortflag('f'),
		config.Required,
	)

	argTrack = cfg.NewInt32(
		"track",
		"the track from which the lyrics are taken. -1 means all tracks, 0 is the first, 1 the second etc",
		config.Shortflag('t'),
		config.Default(int32(-1)),
	)

	argIncludeText = cfg.NewBool(
		"text",
		"include free text entries in the SMF file. Text is surrounded by doublequotes",
	)
)

func read() (text string, err error) {

	var (
		f            *os.File
		msg          midi.Message
		bf           bytes.Buffer
		rd           smf.Reader
		includeText  bool
		trackno      int32
		currentTrack int32
	)

outer:
	for {
		err = cfg.Run()
		if err != nil {
			fmt.Fprintln(os.Stdout, cfg.Usage())
			break
		}

		includeText = argIncludeText.Get()
		trackno = argTrack.Get()

		f, err = os.Open(argFile.Get())
		if err != nil {
			break
		}

		rd = smfreader.New(f)

		for {
			msg, err = rd.Read()

			if err == smfreader.ErrFinished {
				err = nil
				break outer
			}

			if err != nil {
				break outer
			}

			var shouldWrite bool

			if trackno < 0 || trackno == currentTrack {
				shouldWrite = true
			}

			switch v := msg.(type) {

			case meta.Lyric:
				if shouldWrite {
					bf.WriteString(v.Text() + " ")
				}

			case meta.Text:
				if shouldWrite && includeText {
					bf.WriteString(fmt.Sprintf("%#v", v.Text()))
				}

			case meta.ProgramName:
				if shouldWrite {
					bf.WriteString(fmt.Sprintf("[program: %v]\n", v.Text()))
				}

			case meta.TrackInstrument:
				if shouldWrite {
					bf.WriteString(fmt.Sprintf("[instrument: %v]\n", v.Text()))
				}

			default:
				if msg == meta.EndOfTrack {

					// there is only the option to extract lyrics for a specific track or for
					// all tracks, so more than one track always means all tracks, i.e. trackno == -1
					if trackno < 0 {
						bf.WriteString("\n\n")
					}
					currentTrack++
				}

			}

		}
	}

	return bf.String(), err

}

func main() {
	text, err := read()

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, text)
}
