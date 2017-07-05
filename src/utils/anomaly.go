package anomaly

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// struct to read data from the batch/stream json file
type Event struct {
	Name      string  `json:"event_type"`
	Timestamp string  `json:"timestamp"`
	Id        uint32  `json:"id,string"`
	Amount    float64 `json:"amount,string"`
	Id1       uint32  `json:"id1,string"`
	Id2       uint32  `json:"id2,string"`
}

// struct used to output the flagged entry to the Json File
type FlagEntry struct {
	Name      string  `json:"event_type"`
	Timestamp string  `json:"timestamp"`
	Id        uint32  `json:"id,string"`
	Amount    float64 `json:"amount,string"`
	Mean      float64 `json:"mean,string"`
	Sd        float64 `json:"sd,string"`
}

// struct for collecting stats
type Stats struct {
	numpurchases  uint32
	numbefriends  uint32
	numunfriends  uint32
	numgetfriends uint32
}

type PurchaseMap map[uint32]*PurchaseData

// top level object which hold the data for the whole flow
type AnomalyDetection struct {
	graph         Graph
	trans         PurchaseMap
	globalseq     uint64
	Degree        uint32
	Transactions  uint32
	readingstream bool
	stats         Stats
}

// keep the date values the same otherwise the Parse will not work.  you can change the format though
const timeLayout = "2006-01-02 15:04:05"

func NewAnomalyDetection() *AnomalyDetection {
	ad := &AnomalyDetection{}
	ad.globalseq = 1
	ad.Degree = 1
	ad.Transactions = 2
	ad.graph = make(Graph)
	ad.trans = make(PurchaseMap)
	ad.readingstream = false

	ad.stats.numbefriends = 0
	ad.stats.numgetfriends = 0
	ad.stats.numpurchases = 0
	ad.stats.numunfriends = 0

	return ad
}

/////////////////////////////////////////////////////////////////////////////
// Anomaly detection platform Utilities
/////////////////////////////////////////////////////////////////////////////

func (ad *AnomalyDetection) PrintStats() {
	fmt.Println("------------------------------")
	fmt.Println("Total purchases: ", ad.stats.numpurchases)
	fmt.Println("Total befriends: ", ad.stats.numbefriends)
	fmt.Println("Total unfriends: ", ad.stats.numunfriends)
	fmt.Println("Total getfriends: ", ad.stats.numgetfriends)
	fmt.Println("------------------------------")

	ad.stats.numbefriends = 0
	ad.stats.numgetfriends = 0
	ad.stats.numpurchases = 0
	ad.stats.numunfriends = 0
}
func (ad *AnomalyDetection) GetFriends(id uint32, degree uint32) []uint32 {
	ad.stats.numgetfriends += 1
	return ad.graph.GetFriends(id, degree)
}

func (ad *AnomalyDetection) addVertex(id uint32) {
	ad.graph.addVertex(id)
	_, ok := ad.trans[id]
	if ok == false {
		ad.trans[id] = NewPurchaseData()
		ad.trans[id].id = id
	}
}

func set_union(slice1 []uint32, slice2 []uint32) []uint32 {
	res := []uint32{}
	m := map[uint32]uint32{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			res = append(res, mKey)
		}
	}
	return res
}

func (ad *AnomalyDetection) addEdge(id1, id2 uint32) {
	ad.addVertex(id1)
	ad.addVertex(id2)

	ad.stats.numbefriends += 1
	if ad.readingstream == true {
		// Update all the nodes which are affected
		friends1 := ad.GetFriends(id1, ad.Degree-1)
		friends2 := ad.GetFriends(id2, ad.Degree-1)

		ad.graph.AddUndirectedEdge(id1, id2)

		ad.trans[id1].dirty = true
		ad.trans[id2].dirty = true

		commonfriends := set_union(friends1, friends2)
		for _, cfid := range commonfriends {
			ad.trans[cfid].dirty = true
		}
	} else {
		ad.graph.AddUndirectedEdge(id1, id2)
		ad.trans[id1].dirty = true
		ad.trans[id2].dirty = true

	}
}

func (ad *AnomalyDetection) remEdge(id1, id2 uint32) {
	ad.stats.numunfriends += 1
	if ad.readingstream == true {
		friends1 := ad.GetFriends(id1, ad.Degree-1)
		friends2 := ad.GetFriends(id2, ad.Degree-1)

		ad.graph.RemUndirectedEdge(id1, id2)
		ad.trans[id1].dirty = true
		ad.trans[id2].dirty = true

		friendstoupdate := set_union(friends1, friends2)
		for _, nodetoupdate := range friendstoupdate {
			ad.trans[nodetoupdate].dirty = true
		}

	} else {
		ad.trans[id1].dirty = true
		ad.trans[id2].dirty = true

		ad.graph.RemUndirectedEdge(id1, id2)
	}
}

func (ad *AnomalyDetection) addPurchase(userId uint32, timestamp time.Time, amount float64) {
	ad.stats.numpurchases += 1
	ad.addVertex(userId)
	ad.trans[userId].addPurchase(timestamp, amount, ad)
}

func (ad *AnomalyDetection) isAnomalyPurchase(userId uint32, amount float64) (float64, float64, bool) {
	if (ad.trans[userId].frpur).Len() < 2 {
		return math.NaN(), math.NaN(), false
	}
	mean, std := ad.trans[userId].getMeanStd()
	threshold := mean + 3*std

	if amount > threshold {
		return mean, std, true
	}
	return mean, std, false
}

/////////////////////////////////////////////////////////////////////////////
// File Utilities
/////////////////////////////////////////////////////////////////////////////

func (ad *AnomalyDetection) ReadStreamFile(fileName string, flagFileName string) bool {
	fmt.Println("STARTED READING THE STREAM FILE: ", fileName)
	streamFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Cannot open file at location: ", fileName)
		return false
	}

	defer streamFile.Close()

	flaggedFile, err := os.Create(flagFileName)
	if err != nil {
		fmt.Println("Cannot create file at location: ", flagFileName)
	}
	defer flaggedFile.Close()

	// This flag suggests that the Stream file is being read
	ad.readingstream = true
	jsonDec := json.NewDecoder(streamFile)
	jsonEnc := json.NewEncoder(flaggedFile)

	fmt.Println("Degree and Transactions are: ", ad.Degree, ad.Transactions)

	for jsonDec.More() {
		var evt Event
		err := jsonDec.Decode(&evt)
		if err != nil {

			fmt.Println("Error parsing the json file: ", err)
			return false
		}
		tt, _ := time.Parse(timeLayout, strings.Trim(string(evt.Timestamp), `"`))

		if evt.Name == "befriend" {
			ad.addEdge(evt.Id1, evt.Id2)
		} else if evt.Name == "unfriend" {
			ad.remEdge(evt.Id1, evt.Id2)
		} else if evt.Name == "purchase" {
			ad.addPurchase(evt.Id, tt, evt.Amount)
			mean, sd, res := ad.isAnomalyPurchase(evt.Id, evt.Amount)
			if res == true {
				ms := fmt.Sprintf("%.2f", mean)
				ss := fmt.Sprintf("%.2f", sd)
				mf, _ := strconv.ParseFloat(ms, 64)
				sf, _ := strconv.ParseFloat(ss, 64)
				flagEntry := FlagEntry{evt.Name, evt.Timestamp, evt.Id, evt.Amount, mf, sf}
				jsonEnc.Encode(flagEntry)
			}
		} else {
			fmt.Println("Found error reading stream file")
			return false
		}
	}

	return true
}

func (ad *AnomalyDetection) ReadBatchFile(fileName string) bool {
	batchFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Cannot open file at location: ", fileName)
		return false
	}
	defer batchFile.Close()

	type DT struct {
		D uint32 `json:"D,string"`
		T uint32 `json:"T,string"`
	}

	jsonDec := json.NewDecoder(batchFile)

	var dt DT
	err = jsonDec.Decode(&dt)
	if err != nil {
		fmt.Println("Error parsing D and T information from file: ", err)
		return false
	}
	/*
		ad.Degree = dt.D
		ad.Transactions = dt.T
	*/
	var evt Event
	ad.readingstream = false

	for jsonDec.More() {
		err := jsonDec.Decode(&evt)
		if err != nil {
			fmt.Println("Error parsing json file: ", err)
			return false
		}
		tt, _ := time.Parse(timeLayout, strings.Trim(string(evt.Timestamp), `"`))
		if evt.Name == "befriend" {
			ad.addEdge(evt.Id1, evt.Id2)
		} else if evt.Name == "unfriend" {
			ad.remEdge(evt.Id1, evt.Id2)
		} else if evt.Name == "purchase" {
			ad.addPurchase(evt.Id, tt, evt.Amount)
		} else {
			fmt.Println("Unsuspected error while reading the json file:", fileName)
			return false
		}
	}
	return true
}
