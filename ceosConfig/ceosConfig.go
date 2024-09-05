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

	// Structures Go pour mapper le fichier YAML	
	type Node struct {
		Kind     string            `yaml:"kind"`
	}
	
	type Topology struct {
		Nodes map[string]Node  `yaml:"nodes"`
	}
	
	type Config struct {
		Topology Topology   `yaml:"topology"`
	}

// readYMLFile genere une liste de node
func readYMLFile(LabName string) []string {
	// Lire le fichier YAML
	data, err := os.ReadFile(LabName) // Remplace "config.yml" par le nom de ton fichier
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du fichier YAML : %v", err)
	}

	// Décoder le fichier YAML dans une structure Go
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Erreur lors du décodage du fichier YAML : %v", err)
	}

	// Récupérer les noms des switches (clés de la map 'Nodes')
	var switchNames []string
	for nodeName := range config.Topology.Nodes {
		switchNames = append(switchNames, nodeName)
	}

	// Afficher la liste des noms des switches
	return switchNames
}

// generateRandomMacPart génère une partie aléatoire pour l'adresse MAC
func generateRandomMacPart(rng *rand.Rand) string {
	return fmt.Sprintf("%02x:%02x:%02x", rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

// GenerateConfigFiles lit un fichier contenant les noms des routeurs et génère des fichiers de configuration
func GenerateConfigFiles(LabName,FileName string) error {
	// Créer le répertoire de configuration s'il n'existe pas
	fullPath := fmt.Sprintf("%s/%s",LabName, "configs/ceos-config")
	
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du répertoire : %v", err)
	}

	// Ouvrir le fichier contenant les noms des routeurs (switch.txt)
	switchNames := readYMLFile(fmt.Sprintf("%s/%s",LabName, "config.yaml"))

	// Initialiser un compteur pour le SERIALNUMBER
	serialNumber := 1000

	// Créer un générateur de nombres aléatoires
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _,switchName := range switchNames {
		// Générer le nom complet du fichier de configuration
		cfgFileName := filepath.Join(fullPath, fmt.Sprintf("%s.cfg", switchName))

		// Générer la configuration du fichier
		config := fmt.Sprintf(
			"SERIALNUMBER=FDDATACENTER%d\nSYSTEMMACADDR=00:1c:73:%s\n",
			serialNumber, generateRandomMacPart(rng),
		)

		// Créer et écrire dans le fichier de configuration
		err = os.WriteFile(cfgFileName, []byte(config), 0644)
		if err != nil {
			return fmt.Errorf("erreur lors de la création du fichier %s : %v", cfgFileName, err)
		}

		fmt.Printf("Fichier %s généré avec succès.\n", cfgFileName)

		// Incrémenter le numéro de série pour le prochain routeur
		serialNumber++
	}

	return nil
}
