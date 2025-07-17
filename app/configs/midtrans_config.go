package configs

import (
	"log"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var MidtransClient snap.Client

func InitMidtransClient() {
	MidtransClient.New(LoadENV.MIDTRANS_SERVER_KEY, midtrans.Sandbox)
	midtrans.ClientKey = LoadENV.MIDTRANS_CLIENT_KEY
	midtrans.ServerKey = LoadENV.MIDTRANS_SERVER_KEY
	midtrans.Environment = midtrans.Sandbox
	log.Println("✅ Midtrans Snap Client initialized.")
}
