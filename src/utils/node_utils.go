package anomaly

import (
	"math"
	"time"

	"github.com/apganesh/anomaly_detection/src/priorityqueue"
)

//struct to keep track of the purchases made ...
type Purchase struct {
	Amount   float64
	Sequence uint64
	Id       uint32
}

// This is the data which node keeps track for book keeping
type PurchaseData struct {
	id         uint32
	sum        float64
	sum2       float64
	mypur      []*Purchase
	frpur      *priority_queue.PQ
	curfriends []uint32
	dirty      bool
}

// LESS function for Priority Queue (min priority queue)
func (this *Purchase) Less(other interface{}) bool {
	return this.Sequence < other.(*Purchase).Sequence
}

func NewPurchaseData() *PurchaseData {
	pd := &PurchaseData{}
	pd.sum = 0.0
	pd.sum2 = 0.0
	pd.dirty = false
	pd.frpur = priority_queue.New()
	return pd
}

func (pd *PurchaseData) updateFriendPurchase(userId uint32, timestamp time.Time, amount float64, ad *AnomalyDetection) {

	if (pd.frpur).Len() == int(ad.Transactions) {
		top := (pd.frpur).Pop().(*Purchase)
		pd.sum = pd.sum - top.Amount
		pd.sum2 = pd.sum2 - (top.Amount * top.Amount)
	}
	(pd.frpur).Push(&Purchase{amount, ad.globalseq, userId})

	pd.sum = pd.sum + amount
	pd.sum2 = pd.sum2 + (amount * amount)
}

/*
func (pd *PurchaseData) printFriendsPurchases(ad *AnomalyDetection) {
	var purchases []*Purchase
	fmt.Println("------------------------------------")
	for (pd.frpur).Len() > 0 {
		x := (pd.frpur).Pop().(*Purchase)
		purchases = append(purchases, x)
		fmt.Println(x.Sequence, x.Id, x.Amount)
	}
	for _, purch := range purchases {
		(pd.frpur).Push(purch)
	}
}
*/
/*
func (pd *PurchaseData) updateSumAndSum2(ad *AnomalyDetection) {
	var purchases []*Purchase
	pd.sum = 0.0
	pd.sum2 = 0.0

	for (pd.frpur).Len() > 0 {
		x := (pd.frpur).Pop().(*Purchase)
		purchases = append(purchases, x)
		pd.sum += x.Amount
		pd.sum2 += (x.Amount * x.Amount)
	}
	for _, purch := range purchases {
		(pd.frpur).Push(purch)
	}
}
*/

// This updates the latest purchases from all the nodes' friends
func (pd *PurchaseData) updateNodeFriendsPurchases(ad *AnomalyDetection) {
	friends := pd.curfriends

	// Clear the current priority queue
	for (pd.frpur).Len() > 0 {
		(pd.frpur).Pop()
	}

	pd.sum = 0.0
	pd.sum2 = 0.0

	mypq := pd.frpur

	fimap := make(map[uint32]int)

	// Initialize the priority queue
	for _, fid := range friends {
		if len(ad.trans[fid].mypur) > 0 {
			fimap[fid] = len(ad.trans[fid].mypur) - 1
		}
	}

	// number of friends purchases to process
	toprocess := len(fimap)

	// Do a selective k-way merge all the purchases from the friends
	for toprocess > 0 {
		for fid, _ := range fimap {

			if fimap[fid] == -1 {
				continue
			}

			// Initial objects into priority queue
			if mypq.Len() < int(ad.Transactions) {
				curobj := ad.trans[fid].mypur[fimap[fid]]
				pd.sum += curobj.Amount
				pd.sum2 += (curobj.Amount * curobj.Amount)
				mypq.Push(curobj)
				fimap[fid] -= 1
				if fimap[fid] == -1 {
					toprocess -= 1
				}
				continue
			}

			topval := (pd.frpur).Top().(*Purchase)
			if ad.trans[fid].mypur[fimap[fid]].Sequence < topval.Sequence {
				fimap[fid] = -1
				toprocess -= 1
			} else {
				// Pop the top element
				pd.sum -= topval.Amount
				pd.sum2 -= (topval.Amount * topval.Amount)
				(pd.frpur).Pop()
				fimap[topval.Id] = -1
				toprocess -= 1

				// Add the new one
				curobj := ad.trans[fid].mypur[fimap[fid]]
				mypq.Push(ad.trans[fid].mypur[fimap[fid]])
				pd.sum += curobj.Amount
				pd.sum2 += (curobj.Amount * curobj.Amount)
				fimap[fid] -= 1
				if fimap[fid] == -1 {
					toprocess -= 1
				}
			}
		}
	}
	//pd.updateSumAndSum2(ad)
}

func (pd *PurchaseData) addPurchase(timestamp time.Time, amount float64, ad *AnomalyDetection) {

	if len(pd.mypur) == int(ad.Transactions) {
		pd.mypur = pd.mypur[1:]
	}
	pd.mypur = append(pd.mypur, &Purchase{amount, ad.globalseq, pd.id})

	// update the friends at given degree
	if pd.dirty == true && ad.readingstream {
		friends := ad.GetFriends(pd.id, ad.Degree)
		pd.curfriends = friends
	}

	if ad.readingstream == true {
		for _, fid := range pd.curfriends {
			ad.trans[fid].updateFriendPurchase(pd.id, timestamp, amount, ad)
		}
	}

	// If the node is dirty need to update the friends purchases
	if pd.dirty == true && ad.readingstream == true {
		pd.updateNodeFriendsPurchases(ad)
		pd.dirty = false
	}
	ad.globalseq++
}

func (pd *PurchaseData) getMeanStd() (mean float64, std float64) {
	N := float64((pd.frpur).Len())
	mean = pd.sum / N
	variance := ((N * mean * mean) + pd.sum2 - (2 * mean * pd.sum)) / N
	std = math.Sqrt(variance)
	return
}
