package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/trento-project/trento/internal/cluster"
	"github.com/trento-project/trento/internal/cluster/crmmon"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/consul/mocks"
)

func TestCreateInventory(t *testing.T) {
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "trentotest")
	destination := path.Join(tmpDir, "ansible_hosts")

	content := &InventoryContent{
		Nodes: []*Node{
			&Node{
				Name:        "node1",
				AnsibleHost: "192.168.10.1",
				AnsibleUser: "trento",
				Variables: map[string]interface{}{
					"key1": "value1",
					"key2": []string{"value2", "value3"},
				},
			},
			&Node{
				Name: "node2",
			},
		},
		Groups: []*Group{
			&Group{
				Name: "group1",
				Nodes: []*Node{
					{
						Name:        "node3",
						AnsibleHost: "192.168.11.1",
						AnsibleUser: "trento",
						Variables: map[string]interface{}{
							"key1": 1,
							"key2": []string{"value2", "value3"},
						},
					},
					&Node{
						Name: "node4",
					},
				},
			},
			&Group{
				Name: "group2",
				Nodes: []*Node{
					{
						Name: "node5",
					},
					&Node{
						Name: "node6",
					},
				},
			},
		},
	}

	err := CreateInventory(destination, content)

	assert.NoError(t, err)
	assert.FileExists(t, destination)

	// Cannot use backticks as the lines have a final space in many lines
	expectedContent := "\n" +
		"node1 ansible_host=192.168.10.1 ansible_user=trento key1=value1 key2=[value2 value3] \n" +
		"node2 ansible_host= ansible_user= \n" +
		"[group1]\n" +
		"node3 ansible_host=192.168.11.1 ansible_user=trento key1=1 key2=[value2 value3] \n" +
		"node4 ansible_host= ansible_user= \n" +
		"[group2]\n" +
		"node5 ansible_host= ansible_user= \n" +
		"node6 ansible_host= ansible_user= \n"

	data, err := ioutil.ReadFile(destination)
	if err == nil {
		assert.Equal(t, expectedContent, string(data))
	}
}

func mockGetCluster(client consul.Client) (map[string]*cluster.Cluster, error) {
	return map[string]*cluster.Cluster{
		"cluster1": &cluster.Cluster{
			Crmmon: crmmon.Root{
				Nodes: []crmmon.Node{
					crmmon.Node{
						Name: "node1",
					},
					crmmon.Node{
						Name: "node2",
					},
				},
			},
		},
		"cluster2": &cluster.Cluster{
			Crmmon: crmmon.Root{
				Nodes: []crmmon.Node{
					crmmon.Node{
						Name: "node3",
					},
					crmmon.Node{
						Name: "node4",
					},
				},
			},
		},
	}, nil
}

func mockGetCheckSelection(client consul.Client, clusterId string) (string, error) {
	switch clusterId {
	case "cluster1":
		return "check1,check2", nil
	case "cluster2":
		return "check3,check4", nil
	}
	return "", nil
}

func mockGetNodeAddress(client consul.Client, node string) (string, error) {
	switch node {
	case "node1":
		return "192.168.10.1", nil
	case "node2":
		return "192.168.10.2", nil
	case "node3":
		return "192.168.10.3", nil
	case "node4":
		return "", fmt.Errorf("Error getting node address")
	}
	return "", nil
}

func mockGetConnectionName(client consul.Client, clusterId string, node string) (string, error) {
	switch node {
	case "node1":
		return "user1", nil
	case "node2":
		return "user2", nil
	case "node3":
		return "", nil
	case "node4":
		return "", fmt.Errorf("Error getting node user")
	}

	return "", nil
}

func mockGetCloudUserName(client consul.Client, node string) (string, error) {
	switch node {
	case "node3":
		return "clouduser", nil
	case "node4":
		return "", fmt.Errorf("Error getting cloud user")
	}
	return "", nil
}

func TestNewClusterInventoryContent(t *testing.T) {
	consulInst := new(mocks.Client)

	getClusters = mockGetCluster
	getCheckSelection = mockGetCheckSelection
	getNodeAddress = mockGetNodeAddress
	getConnectionName = mockGetConnectionName
	getCloudUserName = mockGetCloudUserName

	content, err := NewClusterInventoryContent(consulInst)

	expectedContent := &InventoryContent{
		Groups: []*Group{
			&Group{
				Name: "cluster1",
				Nodes: []*Node{
					&Node{
						Name: "node1",
						Variables: map[string]interface{}{
							"cluster_selected_checks": "check1,check2",
						},
						AnsibleHost: "192.168.10.1",
						AnsibleUser: "user1",
					},
					&Node{
						Name: "node2",
						Variables: map[string]interface{}{
							"cluster_selected_checks": "check1,check2",
						},
						AnsibleHost: "192.168.10.2",
						AnsibleUser: "user2",
					},
				},
			},
			&Group{
				Name: "cluster2",
				Nodes: []*Node{
					&Node{
						Name: "node3",
						Variables: map[string]interface{}{
							"cluster_selected_checks": "check3,check4",
						},
						AnsibleHost: "192.168.10.3",
						AnsibleUser: "clouduser",
					},
					&Node{
						Name: "node4",
						Variables: map[string]interface{}{
							"cluster_selected_checks": "check3,check4",
						},
						AnsibleHost: "",
						AnsibleUser: "root",
					},
				},
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}
