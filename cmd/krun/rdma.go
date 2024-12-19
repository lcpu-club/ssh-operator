package main

// import "os"

// func determineGIDOffset() (string, error) {
// 	// Determine the GID offset
// 	// FI_VERBS_GID_IDX=9
// 	baseDir := "/sys/class/infiniband/"
// 	cards, err := os.ReadDir(baseDir)
// 	if err != nil {
// 		return "", err
// 	}

// 	for _, card := range cards {
// 		if !card.IsDir() {
// 			continue
// 		}
// 		portsDir := baseDir + card.Name() + "/ports/"
// 		ports, err := os.ReadDir(portsDir)
// 		if err != nil {
// 			return "", err
// 		}

// 		for _, port := range ports {
// 			if !port.IsDir() {
// 				continue
// 			}
// 			gidDevBase := portsDir + port.Name() + "/gid_attrs/ndevs/"
// 			gidTypBase := portsDir + port.Name() + "/gid_attrs/types/"
// 			gidBase := portsDir + port.Name() + "/gids/"

// 			gids, err := os.ReadDir(gidBase)
// 			if err != nil {
// 				return "", err
// 			}

// 			for _, gid := range gids {
// 				if !gid.IsDir() {
// 					continue
// 				}
// 				gidIdx := gid.Name()
// 				gidDev, err := os.ReadFile(gidDevBase + gidIdx)
// 				if err != nil {
// 					continue
// 				}
// 				gidTyp, err := os.ReadFile(gidTypBase + gidIdx)
// 				if err != nil {
// 					continue
// 				}
// 				if string(gidTyp) == "RoCE" {
// 					return string(gidDev), nil
// 				}
// 			}
// 		}
// 	}
// }
