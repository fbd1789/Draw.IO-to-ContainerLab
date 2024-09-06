package ceosConfig

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"log"
	"gopkg.in/yaml.v3"
)

	// Go structures to map the YAML file	
	type Node struct {
		Kind     string            `yaml:"kind"`
	}
	
	type Topology struct {
		Nodes map[string]Node  `yaml:"nodes"`
	}
	
	type Config struct {
		Topology Topology   `yaml:"topology"`
	}

// readYMLFile generates a list of nodes
func readYMLFile(LabName string) []string {
	// Lire le fichier YAML
	data, err := os.ReadFile(LabName)
	if err != nil {
		log.Fatalf("Error while reading the YAML file: %v", err)
	}

	// Decode the YAML file into a Go structure
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error while decoding the YAML file: %v", err)
	}

	// Retrieve the names of the switches (keys from the 'Nodes' map)
	var switchNames []string
	for nodeName := range config.Topology.Nodes {
		switchNames = append(switchNames, nodeName)
	}

	// Display the list of switch names
	return switchNames
}

// generateRandomMacPart génère une partie aléatoire pour l'adresse MAC
func generateRandomMacPart(rng *rand.Rand) string {
	return fmt.Sprintf("%02x:%02x:%02x", rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

// GenerateConfigFiles reads a file containing the names of the routers and generates configuration files
func GenerateConfigFiles(LabName,FileName string) error {
	// Create the configuration directory if it does not exist
	fullPath := fmt.Sprintf("%s/%s",LabName, "configs/ceos-config")
	
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return fmt.Errorf("error while creating the directory: %v", err)
	}

	// Open the file containing the router names (config.yaml)
	switchNames := readYMLFile(fmt.Sprintf("%s/%s",LabName, "config.yaml"))

	// Initialize a counter for the SERIALNUMBER
	serialNumber := 1000

	// Create a random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _,switchName := range switchNames {
		// Generate the full name of the configuration file
		cfgFileName := filepath.Join(fullPath, fmt.Sprintf("%s.cfg", switchName))

		// Generate the file configuration
		config := fmt.Sprintf(
			"SERIALNUMBER=FDDATACENTER%d\nSYSTEMMACADDR=00:1c:73:%s\n",
			serialNumber, generateRandomMacPart(rng),
		)

		// Create and write to the configuration file
		err = os.WriteFile(cfgFileName, []byte(config), 0644)
		if err != nil {
			return fmt.Errorf("error while creating the file %s : %v", cfgFileName, err)
		}

		fmt.Printf("file %s generated successfuly.\n", cfgFileName)

		// Increment the serial number for the next router
		serialNumber++
	}
	return nil
}
