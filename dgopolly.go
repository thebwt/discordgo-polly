package discordgo_polly

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	session2 "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/bitly/go-simplejson"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
)

func relax(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// To keep this module if you don't want to use the config structure I use, this just returns a static credential
func loadConfig() aws.Config {
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
type dgopolly struct {
	pollySession polly.Polly
	pollyConfig  polly.SynthesizeSpeechInput
}

func newPolly(creds credentials.Credentials, aws aws.Config) (ourPolly dgopolly) {
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

func (*dgopolly) say(vc *discordgo.VoiceConnection, text string) {

}
