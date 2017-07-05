package main

import (
	"github.com/apganesh/anomaly_detection/src/utils"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Wrong number of arguments: anomaly_detect batch_log.json stream_log.json flagged_purchases.json")
		return
	}
	batchFileName := os.Args[1]

	var deg int
	var trans int

	ad := anomaly.NewAnomalyDetection()

	if len(os.Args) == 6 {
		deg, _ = strconv.Atoi(os.Args[4])
		trans, _ = strconv.Atoi(os.Args[5])
	}

	if len(os.Args) == 6 {
		ad.Degree = uint32(deg)
		ad.Transactions = uint32(trans)
	}

	fmt.Println("Reading the batch file")

	//defer profile.Start(profile.MemProfile).Stop()
	//defer profile.Start(profile.CPUProfile).Stop()
	readbatch := ad.ReadBatchFile(batchFileName)
	if readbatch == false {
		fmt.Println("Error while reading batch file: ", batchFileName)
		return
	}
	ad.PrintStats()

	fmt.Println("Reading the stream file")
	streamFileName := os.Args[2]
	flagFileName := os.Args[3]
	readstream := ad.ReadStreamFile(streamFileName, flagFileName)
	if readstream == false {
		fmt.Println("Error which reading stream file: ", streamFileName)
		return
	}
	ad.PrintStats()
}
