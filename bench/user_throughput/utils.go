package user_throughput

import (
	"math/rand"
)

type Operation struct {
	Type      int
	ID        int
	PastEpoch int
}

type WorkLoadPerSecond struct {
	Ops []*Operation
}

type WorkLoad struct {
	Work []*WorkLoadPerSecond
}

const AppendUserDay = 0.01728
const LookUpUserDay = 28
const SecondsPerDay = 24 * 60 * 60

var Debug = false

func GenerateWorkLoad(numUser int, epochInterval int, MonitoringInterval int, numID int, testLength int) (int, *WorkLoad) {
	workload := new(WorkLoad)
	workload.Work = make([]*WorkLoadPerSecond, testLength)

	for i := 0; i < testLength; i++ {
		tmp := new(WorkLoadPerSecond)
		tmp.Ops = make([]*Operation, 0)

		workload.Work[i] = tmp
	}

	// rand.Seed(time.Now().UnixNano())
	tot := 0

	numAppend := int(float64(numUser) * AppendUserDay)
	for i := 0; i < numAppend; i++ {
		sec := rand.Intn(SecondsPerDay)
		if sec >= testLength {
			continue
		}
		tot++
		op := &Operation{
			Type:      0, //Append
			ID:        rand.Intn(numID),
			PastEpoch: 0,
		}
		workload.Work[sec].Ops = append(workload.Work[sec].Ops, op)
	}

	for i := 0; i < numUser; i++ {
		id := rand.Intn(numID)
		for j := 0; j < LookUpUserDay; j++ {
			sec := rand.Intn(SecondsPerDay)
			if sec >= testLength {
				continue
			}
			tot++
			op := &Operation{
				Type:      1, //LookUp
				ID:        id,
				PastEpoch: 0,
			}
			workload.Work[sec].Ops = append(workload.Work[sec].Ops, op)
		}
	}

	//NOTE: remove this
	if Debug {
		return tot, workload
	}

	pastEpoch := MonitoringInterval / epochInterval

	if SecondsPerDay%MonitoringInterval != 0 {
		panic("Wrong MonitoringInterval")
	}

	if MonitoringInterval%epochInterval != 0 {
		panic("Wrong EpochInterval")
	}

	monitoringPerSecond := numUser/MonitoringInterval + 1

	for i := 0; i < testLength; i++ {
		for j := 0; j < monitoringPerSecond; j++ {
			id := rand.Intn(numID)
			tot++
			op := &Operation{
				Type:      2, //Monitoring
				ID:        id,
				PastEpoch: pastEpoch,
			}
			workload.Work[i].Ops = append(workload.Work[i].Ops, op)
		}
	}

	// for i := 0; i < numUser; i++ {
	// 	id := rand.Intn(numID)
	//
	// 	st := rand.Intn(SecondsPerDay)
	// 	sec := st
	// 	for ; sec < SecondsPerDay; sec = sec + MonitoringInterval {
	// 		if sec >= testLength {
	// 			continue
	// 		}
	// 		tot++
	// 		op := &Operation{
	// 			Type:      2, //Monitoring
	// 			ID:        id,
	// 			PastEpoch: pastEpoch,
	// 		}
	// 		workload.Work[sec].Ops = append(workload.Work[sec].Ops, op)
	// 	}
	//
	// 	sec = sec - SecondsPerDay
	// 	for ; sec < st; sec = sec + MonitoringInterval {
	// 		if sec >= testLength {
	// 			continue
	// 		}
	// 		tot++
	// 		op := &Operation{
	// 			Type:      2, //Monitoring
	// 			ID:        id,
	// 			PastEpoch: pastEpoch,
	// 		}
	// 		workload.Work[sec].Ops = append(workload.Work[sec].Ops, op)
	// 	}
	// }

	return tot, workload
}
