package main

import (
	// "encoding/json"
	// "bytes"
	"bytes"
	"encoding/xml"
	"html/template"
	"io"
	"strconv"

	// "flag"
	"fmt"
	// "io"
	"os"
	// "strconv"
	// "text/template"
	// "github.com/3th1nk/cidr"
	"github.com/3th1nk/cidr"
	"gopkg.in/ini.v1"
)

type Nodes struct {
	XMLName xml.Name `xml:"mxfile"`
	Diagram Diagram  `xml:"diagram"`
}

type Diagram struct {
	XMLNAME      xml.Name     `xml:"diagram"`
	MxGraphModel MxGraphModel `xml:"mxGraphModel"`
}

type MxGraphModel struct {
	XMLNAME xml.Name `xml:"mxGraphModel"`
	Root    Root     `xml:"root"`
}

type Root struct {
	XMLNAME xml.Name `xml:"root"`
	MxCell  []MxCell `xml:"mxCell"`
}

type MxCell struct {
	ID     string `xml:"id,attr"`
	Value  string `xml:"value,attr"`
	Parent string `xml:"parent,attr"`
	Edge   string `xml:"edge,attr"`
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

type DeviceItem struct {
	Id              string `json:"Id"`
	Name            string `json:"Name"`
	InterfaceNumber int    `json:"InterfaceNumber"`
}

type Devices struct {
	ItemsNode []DeviceItem
}

type LinkItem struct {
	SourceName string
	SourcePort string
	TargetName string
	TargetPort string
}

type Links struct {
	ItemsLink []LinkItem
}

const VersionCode = "0.1"

type Environment struct {
	Image     string
	LabName   string
	IpAddress string
	Network   string
}

func (d *Devices) AddItem(item DeviceItem) {
	d.ItemsNode = append(d.ItemsNode, item)
}

func (l *Links) AddItem(item LinkItem) {
	l.ItemsLink = append(l.ItemsLink, item)
}

func main() {
	cfg, err := ini.Load("myConfig.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	// fmt.Println("sourceFile:", cfg.Section("").Key("sourceFile").String())

	sourceFile := cfg.Section("").Key("sourceFile").String()
	targetFile := cfg.Section("").Key("targetFile").String()
	ceosImage := cfg.Section("").Key("ceosImage").String()
	management := cfg.Section("").Key("management").String()
	labName := cfg.Section("").Key("labName").String()

	d := Devices{}
	l := Links{}

	// Verify the IP address
	NetworkName := ""
	if management != "" {
		_, err0 := cidr.Parse(management)
		if err0 != nil {
			fmt.Println(err0)
			return
		}
		NetworkName = labName + "-mgnt"
	}

	environment := Environment{Image: "arista/ceos:" + ceosImage, LabName: labName, IpAddress: management, Network: NetworkName}

	file, err := os.Open(sourceFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var nodes Nodes
	xml.Unmarshal(byteValue, &nodes)
	for i := 0; i < len(nodes.Diagram.MxGraphModel.Root.MxCell); i++ {
		// fmt.Println(nodes.Diagram.MxGraphModel.Root.MxCell[i].ID,
		// 	nodes.Diagram.MxGraphModel.Root.MxCell[i].Value,
		// 	nodes.Diagram.MxGraphModel.Root.MxCell[i].Parent,
		// 	nodes.Diagram.MxGraphModel.Root.MxCell[i].Edge,
		// 	nodes.Diagram.MxGraphModel.Root.MxCell[i].Source,
		// 	nodes.Diagram.MxGraphModel.Root.MxCell[i].Target)
		if nodes.Diagram.MxGraphModel.Root.MxCell[i].Value != "" {
			item1 := DeviceItem{Id: nodes.Diagram.MxGraphModel.Root.MxCell[i].ID,
				Name:            nodes.Diagram.MxGraphModel.Root.MxCell[i].Value,
				InterfaceNumber: 1}
			d.AddItem(item1)
		}
		if nodes.Diagram.MxGraphModel.Root.MxCell[i].Source != "" && nodes.Diagram.MxGraphModel.Root.MxCell[i].Target != "" {
			item1 := LinkItem{SourceName: nodes.Diagram.MxGraphModel.Root.MxCell[i].Source,
				SourcePort: "empty",
				TargetName: nodes.Diagram.MxGraphModel.Root.MxCell[i].Target,
				TargetPort: "empty"}
			l.AddItem(item1)
		}
	}

	// Mise en place des noms des devices et des interfacess
	for i, link := range l.ItemsLink {
		for j, node := range d.ItemsNode {
			if link.SourceName == node.Id {
				l.ItemsLink[i].SourceName = node.Name
				d.ItemsNode[j].InterfaceNumber++
				l.ItemsLink[i].SourcePort = "eth" + strconv.Itoa((node.InterfaceNumber))
			}
			if link.TargetName == node.Id {
				l.ItemsLink[i].TargetName = node.Name
				d.ItemsNode[j].InterfaceNumber++
				l.ItemsLink[i].TargetPort = "eth" + strconv.Itoa((node.InterfaceNumber))
			}
		}
	}

	const (
		headerTemplate = `
	name: {{.LabName}}
	mgmt:
	  network: {{.Network}}
	  ipv4-subnet: {{.IpAddress}}
	topology:
	  kinds:
	    ceos:
	      image: {{.Image -}}
	`
	)

	const (
		nodeTemplate = `
	  nodes:
	    {{- range .ItemsNode}}
	    {{.Name}}:
	      kind: ceos
		{{- end -}}
	`
	)

	const (
		linkTemplate = `
	  links:
	    {{- range .ItemsLink}}
	    - endpoints: ["{{.SourceName}}:{{.SourcePort}}","{{.TargetName}}:{{.TargetPort}}"]
		{{- end}}
	`
	)

	var tpl bytes.Buffer
	tmpl, err := template.New("test").Parse(headerTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&tpl, environment)
	if err != nil {
		panic(err)
	}

	tmpl, err = template.New("test").Parse(nodeTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&tpl, d)
	if err != nil {
		panic(err)
	}

	tmpl, err = template.New("test").Parse(linkTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&tpl, l)
	if err != nil {
		panic(err)
	}

	str := tpl.String()

	// Sauvegarde de la donnee dans un fichier
	f, _ := os.Create(targetFile)
	defer f.Close()
	f.Write([]byte(str))
}
