package main

import (
	// "encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/template"
	"bytes"
	"flag"
)

type Nodes struct{
	XMLName xml.Name `xml:"mxfile"`
	Diagram	Diagram		`xml:"diagram"`
}

type Diagram struct {
	XMLNAME		xml.Name	`xml:"diagram"`
	MxGraphModel	MxGraphModel 	`xml:"mxGraphModel"`
}

type MxGraphModel struct {
	XMLNAME 	xml.Name	`xml:"mxGraphModel"`
	Root 		Root 		`xml:"root"`
}

type Root struct {
	XMLNAME 	xml.Name	`xml:"root"`
	MxCell 		[]MxCell 		`xml:"mxCell"`
}

type MxCell struct {
	ID		string		`xml:"id,attr"`
	Value 	string		`xml:"value,attr"`
	Parent 	string		`xml:"parent,attr"`
	Edge	string		`xml:"edge,attr"`
	Source	string		`xml:"source,attr"`
	Target	string		`xml:"target,attr"`
}


type DeviceItem struct {
	Id string `json:"Id"`
	Name string `json:"Name"`
	InterfaceNumber int `json:"InterfaceNumber"`
}

type Devices struct {
	ItemsNode []DeviceItem
}

type LinkItem struct{
	SourceName string
	SourcePort string
	TargetName string
	TargetPort string
}

type Links struct {
	ItemsLink []LinkItem
}

type Environment struct {
	Image  string
}


func (d *Devices) AddItem(item DeviceItem) {
	d.ItemsNode = append(d.ItemsNode, item)
}

func (l *Links) AddItem(item LinkItem) {
	l.ItemsLink = append(l.ItemsLink, item )
}

func main () {
	sourceFile := flag.String("s","default.xml","source file name")
	targetFile := flag.String("t","default.yml","target file name")
	ceosImage := flag.String("i","4.30.3M","image version of the code")
	flag.Parse()

	d :=Devices{}
	l :=Links{}
	// environment := Environment{Image: "arista/ceos:4.30.3M"}
	environment := Environment{Image: "arista/ceos:"+ *ceosImage}

	file, err := os.Open(*sourceFile)
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()
	
	byteValue,_ := io.ReadAll(file)

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
			item1 :=DeviceItem{Id: nodes.Diagram.MxGraphModel.Root.MxCell[i].ID,
			Name: nodes.Diagram.MxGraphModel.Root.MxCell[i].Value,
			InterfaceNumber: 1,}
			d.AddItem(item1)
		}
		if nodes.Diagram.MxGraphModel.Root.MxCell[i].Source != "" && nodes.Diagram.MxGraphModel.Root.MxCell[i].Target != "" {
			item1 :=LinkItem{SourceName: nodes.Diagram.MxGraphModel.Root.MxCell[i].Source,
			SourcePort: "empty",
			TargetName: nodes.Diagram.MxGraphModel.Root.MxCell[i].Target,
			TargetPort:"empty",}
			l.AddItem(item1)
		}
	}

	// Mise en place des noms des devices et des interfacess
	for i, link :=range l.ItemsLink{
		for j, node :=range d.ItemsNode {
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


	const (headerTemplate = `
name: lab
topology:
  kinds:
    ceos:
      image: {{.Image -}}
`)


	const (nodeTemplate = `
  nodes:
    {{- range .ItemsNode}}
    {{.Name}}:
      kinds: ceos
	{{- end -}}
`)

	const (linkTemplate = `
  links:
    {{- range .ItemsLink}}
	  - endpoints: ['{{.SourceName}}:{{.SourcePort}}','{{.TargetName}}:{{.TargetPort}}']
	{{- end}}
`)



	var tpl bytes.Buffer
	tmpl, err := template.New("test").Parse(headerTemplate)
		if err != nil { panic(err) }
	err = tmpl.Execute(&tpl, environment)
		if err != nil { panic(err)}
	
	tmpl, err = template.New("test").Parse(nodeTemplate)
		if err != nil { panic(err) }
	err = tmpl.Execute(&tpl, d)
		if err != nil { panic(err)}

	tmpl, err = template.New("test").Parse(linkTemplate)
		if err != nil { panic(err) }
	err = tmpl.Execute(&tpl, l)
		if err != nil { panic(err)}

		str := tpl.String()
	
	
	// Sauvegarde de la donnee dans un fichier
	f, _ := os.Create(*targetFile)
	defer f.Close()
	f.Write([]byte(str))
}
