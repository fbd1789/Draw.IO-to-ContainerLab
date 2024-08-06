package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"gopkg.in/yaml.v3"
	"gopkg.in/ini.v1"
)

type Mxfile struct {
	XMLName  xml.Name `xml:"mxfile"`
	Host     string   `xml:"host,attr"`
	Modified string   `xml:"modified,attr"`
	Agent    string   `xml:"agent,attr"`
	Version  string   `xml:"version,attr"`
	Etag     string   `xml:"etag,attr"`
	Type     string   `xml:"type,attr"`
	Diagram  Diagram  `xml:"diagram"`
}

type Diagram struct {
	Name         string       `xml:"name,attr"`
	ID           string       `xml:"id,attr"`
	MxGraphModel MxGraphModel `xml:"mxGraphModel"`
}

type MxGraphModel struct {
	Dx         string  `xml:"dx,attr"`
	Dy         string  `xml:"dy,attr"`
	Grid       string  `xml:"grid,attr"`
	GridSize   string  `xml:"gridSize,attr"`
	Guides     string  `xml:"guides,attr"`
	Tooltips   string  `xml:"tooltips,attr"`
	Connect    string  `xml:"connect,attr"`
	Arrows     string  `xml:"arrows,attr"`
	Fold       string  `xml:"fold,attr"`
	Page       string  `xml:"page,attr"`
	PageScale  string  `xml:"pageScale,attr"`
	PageWidth  string  `xml:"pageWidth,attr"`
	PageHeight string  `xml:"pageHeight,attr"`
	Math       string  `xml:"math,attr"`
	Shadow     string  `xml:"shadow,attr"`
	Root       Root    `xml:"root"`
}

type Root struct {
	MxCell []MxCell `xml:"mxCell"`
}

type MxCell struct {
	ID       string      `xml:"id,attr"`
	Parent   string      `xml:"parent,attr,omitempty"`
	Style    string      `xml:"style,attr,omitempty"`
	Edge     string      `xml:"edge,attr,omitempty"`
	Source   string      `xml:"source,attr,omitempty"`
	Target   string      `xml:"target,attr,omitempty"`
	Vertex   string      `xml:"vertex,attr,omitempty"`
	Value    string      `xml:"value,attr,omitempty"`
	MxGeometry MxGeometry `xml:"mxGeometry"`
}

type MxGeometry struct {
	Relative string `xml:"relative,attr,omitempty"`
	X        string `xml:"x,attr,omitempty"`
	Y        string `xml:"y,attr,omitempty"`
	Width    string `xml:"width,attr,omitempty"`
	Height   string `xml:"height,attr,omitempty"`
}

//  Node Extraction
type Nodes struct {
	ID			string
	Name		string
	MgmtIPv4 	string
	Env      	map[string]string
}
func extractNodes(mxfile Mxfile, VrfMgmt string) []Nodes {
	var nodes []Nodes
	for _, value := range mxfile.Diagram.MxGraphModel.Root.MxCell {
		if len(value.Value) != 0 {
			node := Nodes{
				ID:   value.ID,
				Name: value.Value,
				Env: map[string]string{"CLAB_MGMT_VRF": VrfMgmt},
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Line Extraction
type Lines struct {
	Source	string
	Target	string
	PortSource string
	PortTarget string
}
func extractLines (mxfile Mxfile, result  map[string]string) []Lines {
	var lines []Lines
	increment := func (port string) string{
		newPort,_ := strconv.Atoi(port)
		newPort++
		return strconv.Itoa(newPort)
	}
	for _, value := range mxfile.Diagram.MxGraphModel.Root.MxCell {
		if len(value.Source)!=0 && len(value.Target)!=0 {
			deviceSource := result[value.Source]
			interfaceSource := increment(result[deviceSource])
			result[deviceSource] = interfaceSource

			deviceTarget := result[value.Target]
			interfaceTarget := increment(result[deviceTarget])
			result[deviceTarget] = interfaceTarget
			line := Lines{
				Source: result[value.Source],
				Target: result[value.Target],
				PortSource: "eth" + interfaceSource,
				PortTarget: "eth" + interfaceTarget,
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// YML structure for the export
type Management struct {
	Network    string `yaml:"network"`
	IPv4Subnet string `yaml:"ipv4-subnet"`
}

type Kind struct {
	Image string `yaml:"image"`
}

type Kinds struct {
	Ceos Kind `yaml:"ceos"`
}

type Config struct {
	Name     string     `yaml:"name"`
	Mgmt     Management `yaml:"mgmt"`
	Topology Topology   `yaml:"topology"`
}

type Topology struct {
	Kinds Kinds           `yaml:"kinds"`
    Nodes map[string]Node `yaml:"nodes"`
	Links  []Link          `yaml:"links"`
}
type Node struct {
    Kind string `yaml:"kind"`
	MgmtIPv4 string `yaml:"mgmt-ipv4"`
	Env      map[string]string `yaml:"env"`
}

type Link struct {
    Endpoints []string `yaml:"endpoints,flow"`
}

func main() {
	// Read the config.ini file
	inidata, err := ini.Load("config.ini")
	if err != nil {
	   fmt.Printf("Fail to read file: %v", err)
	   os.Exit(1)
	 }
	LabName := inidata.Section("global").Key("nameLab").String()
	Ipv4Subnet := inidata.Section("mgmt").Key("ipv4Subnet").String()
	NetworkMgmt := LabName + "-mgmt"
	ImageCeos := inidata.Section("topolgy").Key("image").String()
	VrfMgmt := inidata.Section("nodes").Key("vrf").String()

	// Read the XML file
	byteValue, err := os.ReadFile("test3.xml")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Unmarshal the XML data into the struct
	var mxfile Mxfile
	err = xml.Unmarshal(byteValue, &mxfile)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return
	}

	// Extract nodes
	nodes := extractNodes(mxfile, VrfMgmt)

	// Create json
	// ID is the Key, Name is the Value
	// Name is the Key, "1" is the Value representing the number of interface (eth)
	result := make(map[string]string)
    for _, node := range nodes {
        result[node.ID] = node.Name
		result[node.Name] = "1"
    }

	// Extract lines
	lines := extractLines(mxfile, result)

	// Build the YML for the links section
	var links []Link
    for _, ld := range lines {
        link := Link{
            Endpoints: []string{fmt.Sprintf("%s:%s", ld.Source, ld.PortSource), fmt.Sprintf("%s:%s", ld.Target, ld.PortTarget)},
        }
        links = append(links, link)
    }

	// Build yml structure
	configTest := Config{
		Name: LabName,
		Mgmt: Management{
			Network:    NetworkMgmt,
			IPv4Subnet: Ipv4Subnet,
		},
		Topology: Topology{
			Kinds: Kinds{
				Ceos: Kind{
					// Image: "arista/ceos:4.30.3M",
					Image: fmt.Sprintf("arista/ceos:%s", ImageCeos),
				},
			},
			Nodes: make(map[string]Node),
			Links: links,
		},
	} 

	// Build the YML for the nodes section
	addNode := func(name, kind string, mgmtIPv4 string, env map[string]string) {
		configTest.Topology.Nodes[name] = Node{Kind: kind, MgmtIPv4: mgmtIPv4, Env: env,}
    }
	for _, node := range nodes {
		addNode(node.Name,"ceos", node.MgmtIPv4, node.Env)
	}

	// Marshal the configuration to YAML
    yamlData, err := yaml.Marshal(&configTest)
    if err != nil {
        log.Fatalf("error: %v", err)
    }

    // Print the YAML
    fmt.Println(string(yamlData))
	// Save the YAML data to a file
	fileName := "config.yaml"
	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
	}

	fmt.Printf("YAML data has been written to %s\n", fileName)

}