package ganeti

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Job struct {
	Status string `json:"status"`
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Explain string `json:"explain"`
}

type Client struct {
	Username   string
	Password   string
	Connection string
	SSLVerify  bool
}

type Disk struct {
	Size string `json:"size"`
}

type NIC struct {
	Link string `json:"link"`
}

type BackendParams struct {
	Memory string `json:"memory"`
	VCPUs  int    `json:"vcpus"`
}

type BackendParamsRead struct {
	Memory int `json:"memory"`
	VCPUs  int `json:"vcpus"`
}

type Instance struct {
	BackendParams BackendParams `json:"beparams"`
	Disks         []Disk        `json:"disks"`
	DiskTemplate  string        `json:"disk_template"`
	GroupName     string        `json:"group_name,omitempty"`
	Hypervisor    string        `json:"hypervisor,omitempty"`
	Name          string        `json:"instance_name"`
	Mode          string        `json:"mode"`
	NICs          []NIC         `json:"nics"`
	Node          string        `json:"pnode,omitempty"`
	OSType        string        `json:"os_type,omitempty"`
	Version       int           `json:"__version__"`
}

type InstanceRead struct {
	AdminState    string            `json:"admin_state"`
	BackendParams BackendParamsRead `json:"beparams"`
	DiskTemplate  string            `json:"disk_template"`
	Name          string            `json:"name"`
	Node          string            `json:"pnode"`
	OSType        string            `json:"os"`
	Status        string            `json:"status"`
	UUID          string            `json:"uuid"`
}

func NewClient(host string, port int, username string, password string, apiVersion int, useSsl bool, sslVerify bool) (*Client, error) {
	var c Client

	proto := "http"

	if useSsl {
		proto = "https"
	}

	connStr := fmt.Sprintf("%s://%s:%d/%d", proto, host, port, apiVersion)

	c.Username = username
	c.Password = password
	c.Connection = connStr
	c.SSLVerify = sslVerify

	return &c, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.SSLVerify},
	}

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: httpTransport,
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(c.Username, c.Password))
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var e Error

		err = json.Unmarshal(body, &e)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("[%d] %s - %s", e.Code, e.Message, e.Explain)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) ReadInstance(id string) (*InstanceRead, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/instances/%s", c.Connection, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var instance InstanceRead

	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (c *Client) CreateInstance(instance Instance) (*Instance, error) {
	rb, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/instances", c.Connection), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	jobByte, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	jobStr := strings.TrimSuffix(string(jobByte), "\n")

	job := Job{
		Status: "pending",
	}

	for {
		switch job.Status {
		case "error":
			return nil, fmt.Errorf("Job ID %s failed to create instance", jobStr)
		case "success":
			return &instance, nil
		}

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/jobs/%s", c.Connection, jobStr), nil)
		if err != nil {
			return nil, err
		}

		res, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(res, &job)
		if err != nil {
			return nil, err
		}

		time.Sleep(time.Second * 10)
	}
}

func (c *Client) DeleteInstance(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/instances/%s", c.Connection, id), nil)
	if err != nil {
		return err
	}

	jobByte, err := c.doRequest(req)
	if err != nil {
		return err
	}

	jobStr := strings.TrimSuffix(string(jobByte), "\n")

	job := Job{
		Status: "pending",
	}

	for {
		switch job.Status {
		case "error":
			return fmt.Errorf("Job ID %s failed to delete instance", jobStr)
		case "success":
			return nil
		}

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/jobs/%s", c.Connection, jobStr), nil)
		if err != nil {
			return err
		}

		res, err := c.doRequest(req)
		if err != nil {
			return err
		}

		err = json.Unmarshal(res, &job)
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 10)
	}
}
