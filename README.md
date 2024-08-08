# Draw.IO-to-ContainerLab

## Introduction
Draw a network diagram on DrawIO and generate a yml file for ContainerLab

## Installation

### Simple usage
Step 1: Draw your network

![Alt text](DrawIOExemple.png)

Step 2: Export your network
Export from DrawIO the schemas in XML

![Alt text](drawIOExemple1.png)

Step 3: Choose your binary
- Windows : versionWindows.exe
- Linux : versionLinux

Step 4: Use it
```
Define the information that needed in the config.ini file
[global]
nameLab = MonLab

[mgmt]
ipv4Subnet = 172.20.20.0/24

[topolgy]
image = 4.30.3M

[nodes]
vrf = MGMT
```

Exemple: versionWindows.exe
A config.yaml is generated for the containerLab
```
name: MonLab
mgmt:
    network: MonLab-mgmt
    ipv4-subnet: 172.20.20.0/24
topology:
    kinds:
        ceos:
            image: arista/ceos:4.30.3M
    nodes:
        Leaf1a:
            kind: ceos
            mgmt-ipv4: 172.20.20.4/24
            env:
                CLAB_MGMT_VRF: MGMT
        Leaf1b:
            kind: ceos
            mgmt-ipv4: 172.20.20.5/24
            env:
                CLAB_MGMT_VRF: MGMT
        Leaf2a:
            kind: ceos
            mgmt-ipv4: 172.20.20.6/24
            env:
                CLAB_MGMT_VRF: MGMT
        Leaf2b:
            kind: ceos
            mgmt-ipv4: 172.20.20.7/24
            env:
                CLAB_MGMT_VRF: MGMT
        Spine1:
            kind: ceos
            mgmt-ipv4: 172.20.20.1/24
            env:
                CLAB_MGMT_VRF: MGMT
        Spine2:
            kind: ceos
            mgmt-ipv4: 172.20.20.2/24
            env:
                CLAB_MGMT_VRF: MGMT
        Spine3:
            kind: ceos
            mgmt-ipv4: 172.20.20.3/24
            env:
                CLAB_MGMT_VRF: MGMT
    links:
        - endpoints: ['Spine1:eth1', 'Leaf1a:eth1']
        - endpoints: ['Spine1:eth2', 'Leaf1b:eth1']
        - endpoints: ['Spine1:eth3', 'Leaf2a:eth1']
        - endpoints: ['Spine1:eth4', 'Leaf2b:eth1']
        - endpoints: ['Spine2:eth1', 'Leaf1a:eth2']
        - endpoints: ['Spine2:eth2', 'Leaf1b:eth2']
        - endpoints: ['Spine2:eth3', 'Leaf2b:eth2']
        - endpoints: ['Spine3:eth1', 'Leaf1b:eth3']
        - endpoints: ['Spine3:eth2', 'Leaf1a:eth3']
        - endpoints: ['Leaf1a:eth4', 'Leaf1b:eth4']
        - endpoints: ['Leaf1b:eth5', 'Leaf1a:eth5']
        - endpoints: ['Leaf2a:eth2', 'Leaf2b:eth3']
        - endpoints: ['Leaf2b:eth4', 'Leaf2a:eth3']
        - endpoints: ['Spine2:eth4', 'Leaf2a:eth4']
        - endpoints: ['Spine3:eth3', 'Leaf2b:eth5']
        - endpoints: ['Spine3:eth4', 'Leaf2a:eth5']

```

### Fun usage
#### Prerequisite
It is my first software in Go !!

```
go version
go version go1.21.5 linux/amd64
```

```
git clone https://github.com/fbd1789/Draw.IO-to-ContainerLab.git
```