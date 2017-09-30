package discordgo_polly

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
)

func relax(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// To keep this module if you don't want to use the config structure I use, this just returns a static credential
func loadConfig() credentials.Credentials {
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

	return *credentials.NewStaticCredentials(awsapiid, awsapisecret, "")
}
