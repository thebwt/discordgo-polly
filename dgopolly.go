package discordgo_polly

import (
	"encoding/binary"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	session2 "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/bitly/go-simplejson"
	"github.com/bwmarrin/discordgo"
	"io"
	"io/ioutil"
	"layeh.com/gopus"
	"log"
	"sync"
)

func relax(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// To keep this module if you don't want to use the config structure I use, this just returns a static credential
func LoadConfig() aws.Config {
	//load our bot config
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	fileLoad, err := simplejson.NewJson(file)
	if err != nil {
		log.Fatal(err)
	}

	//AWS credentials for polly
	awsid := fileLoad.Get("awsapiid")
	if awsid == nil {
		log.Println("No AWS APIID")
	}
	awsapiid, err := awsid.String()
	relax(err)
	awssecret := fileLoad.Get("awsapisecret")
	if awssecret == nil {
		log.Println("no AWS APISECRET")
	}
	awsapisecret, err := awssecret.String()
	relax(err)
	awsregion := fileLoad.Get("awsapiregion")
	if awssecret == nil {
		log.Println("no AWS APISREGION")
	}
	awsapiregion, err := awsregion.String()
	relax(err)

	return aws.Config{
		Region:      aws.String(awsapiregion),
		Credentials: credentials.NewStaticCredentials(awsapiid, awsapisecret, ""),
	}
}

//ultimate goal is to have a function that we can just say
type DgoPolly struct {
	pollySession polly.Polly
	pollyConfig  polly.SynthesizeSpeechInput
}

func newPolly(creds credentials.Credentials, aws aws.Config) (ourPolly DgoPolly) {
	session, err := session2.NewSession(&aws)
	relax(err)
	ourPolly.pollySession = *polly.New(session)

	testToSynth := "Default Message"
	outputFormat := polly.OutputFormatPcm

	pollyconfig := polly.SynthesizeSpeechInput{
		Text:         &testToSynth,
		OutputFormat: &outputFormat,
	}

	ourPolly.pollyConfig = pollyconfig
	return
}

func (p *DgoPolly) say(vc *discordgo.VoiceConnection, text string) {
	opusChannel := p.getSynthSpeech(text)
	for {
		opus, ok := <-*opusChannel
		if !ok {
			return
		}

		vc.OpusSend <- opus
	}
}

//Returns an Opus channel for play at 48000Hz, what discordGo needs for vc.OpusSend
func (p *DgoPolly) getSynthSpeech(text string) *chan []byte {
	p.pollyConfig.Text = &text

	request, response := p.pollySession.SynthesizeSpeechRequest(&p.pollyConfig)
	err := request.Send()
	relax(err)

	outputChan := make(chan []byte, 1000)
	encodeChan := make(chan []int16, 1000)

	opusEncoder, err := gopus.NewEncoder(16000, 1, gopus.Voip)
	relax(err)
	opusEncoder.SetBitrate(64 * 1000)

	wg := sync.WaitGroup{}

	loadBuffer := func(input io.Reader) {

		defer func() {
			fmt.Println("done loading frames")
			wg.Done()
			close(encodeChan)
		}()

		for {
			//960 is the framesize, 1 is the number of channels the input has
			inBuf := make([]int16, 960*1)
			err := binary.Read(input, binary.LittleEndian, &inBuf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return
			}
			//Upsamping.. apparently
			// polly does audio in 16KHz, discord needs 48KHz, so we add every frame 3 times.
			encodeChan <- inBuf
			encodeChan <- inBuf
			encodeChan <- inBuf

		}
	}

	encodeBuffer := func() {
		defer func() {
			fmt.Println("Done encoding new frames")
			wg.Done()
			close(outputChan)
		}()

		for {
			pcm, ok := <-encodeChan
			if !ok {
				return
			}
			//we want audio in two channels. So the normal 960*2, plus the additional *2 because the source has more
			// channels
			opus, err := opusEncoder.Encode(pcm, 960, 960*2*2)
			relax(err)

			outputChan <- opus
		}
	}

	wg.Add(2)
	go loadBuffer(response.AudioStream)
	go encodeBuffer()
	wg.Wait()

	return &outputChan

}
