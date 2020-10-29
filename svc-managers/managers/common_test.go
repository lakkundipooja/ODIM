//(C) Copyright [2020] Hewlett Packard Enterprise Development LP
//
//Licensed under the Apache License, Version 2.0 (the "License"); you may
//not use this file except in compliance with the License. You may obtain
//a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//License for the specific language governing permissions and limitations
// under the License.

package managers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ODIM-Project/ODIM/lib-utilities/common"
	"github.com/ODIM-Project/ODIM/lib-utilities/config"
	"github.com/ODIM-Project/ODIM/lib-utilities/errors"
	"github.com/ODIM-Project/ODIM/svc-managers/mgrcommon"
	"github.com/ODIM-Project/ODIM/svc-managers/mgrmodel"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetExternalInterface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExternalInterface(); got == nil {
				t.Errorf("Result of GetExternalInterface() should not be equal to nil")
			}
		})
	}
}

func mockGetExternalInterface() *ExternalInterface {
	return &ExternalInterface{
		Device: Device{
			GetDeviceInfo: mockGetDeviceInfo,
			ContactClient: mockContactClient,
		},
		DB: DB{
			GetAllKeysFromTable: mockGetAllKeysFromTable,
			GetManagerByURL:     mockGetManagerByURL,
			GetPluginData:       mockGetPluginData,
			UpdateManagersData:  mockUpdateManagersData,
			GetResource:         mockGetResource,
		},
	}
}

func mockGetAllKeysFromTable(table string) ([]string, error) {
	return []string{"/redfish/v1/Managers/uuid:1"}, nil
}

func mockGetManagerData(id string) (mgrmodel.RAManager, error) {
	if id == "nonExistingUUID" {
		return mgrmodel.RAManager{}, fmt.Errorf("not found")
	} else if id == "noDevice" {
		return mgrmodel.RAManager{
			Name:            "odimra",
			ManagerType:     "Service",
			FirmwareVersion: "1.0",
			ID:              "noDevice",
			UUID:            "noDevice",
			State:           "Absent",
		}, nil
	}
	return mgrmodel.RAManager{
		Name:            "odimra",
		ManagerType:     "Service",
		FirmwareVersion: "1.0",
		ID:              config.Data.RootServiceUUID,
		UUID:            config.Data.RootServiceUUID,
		State:           "Enabled",
	}, nil
}

func mockGetManagerByURL(url string) (string, *errors.Error) {
	if url == "/redfish/v1/Managers/invalidURL:1" || url == "/redfish/v1/Managers/invalidURL" || url == "/redfish/v1/Managers/invalidID" {
		return "", errors.PackError(errors.DBKeyNotFound, "not found")
	}
	managerData := make(map[string]interface{})
	managerData["ManagerType"] = "BMC"
	managerData["Status"] = `{"State":"Enabled"}}`
	managerData["Name"] = "somePlugin"
	if url == "/redfish/v1/Managers/uuid" {
		managerData["Name"] = "someOtherID"
	} else if url == "/redfish/v1/Managers/noPlugin" {
		managerData["Name"] = "noPlugin"
	} else if url == "/redfish/v1/Managers/noToken" {
		managerData["Name"] = "noToken"
	}
	data, _ := json.Marshal(managerData)
	return string(data), nil
}

func mockGetPluginData(pluginID string) (mgrmodel.Plugin, *errors.Error) {
	if pluginID == "someOtherID" {
		return mgrmodel.Plugin{
			IP:                "localhost",
			Port:              "9091",
			Username:          "admin",
			Password:          []byte("password"),
			ID:                "CFM",
			PreferredAuthType: "XAuthToken",
		}, nil
	} else if pluginID == "noToken" {
		return mgrmodel.Plugin{
			IP:                "localhost",
			Port:              "9092",
			Username:          "admin",
			Password:          []byte("password"),
			ID:                "noToken",
			PreferredAuthType: "XAuthToken",
		}, nil
	} else if pluginID == "noPlugin" {
		return mgrmodel.Plugin{}, errors.PackError(errors.DBKeyNotFound, "not found")
	}
	return mgrmodel.Plugin{
		IP:                "localhost",
		Port:              "9093",
		Username:          "admin",
		Password:          []byte("password"),
		ID:                "somePlugin",
		PreferredAuthType: "BasicAuth",
	}, nil
}

func mockUpdateManagersData(key string, managerData map[string]interface{}) error {
	return nil
}

func mockGetResource(table, key string) (string, *errors.Error) {
	if key == "/redfish/v1/Managers/uuid1:1/Ethernet" {
		return "", errors.PackError(errors.DBKeyNotFound, "not found")
	}
	return "body", nil
}

func mockGetDeviceInfo(req mgrcommon.ResourceInfoRequest) (string, error) {
	if req.URL == "/redfish/v1/Managers/deviceAbsent:1" || req.URL == "/redfish/v1/Managers/uuid1:1/Ethernet" {
		return "", fmt.Errorf("error")
	}
	manager := mgrmodel.Manager{
		Status: &mgrmodel.Status{
			State: "Enabled",
		},
	}
	dataByte, err := json.Marshal(manager)
	return string(dataByte), err
}

func mockContactClient(url, method, token string, odataID string, body interface{}, loginCredential map[string]string) (*http.Response, error) {

	if url == "https://localhost:9091/ODIM/v1/Sessions" {
		body := `{"Token": "12345"}`
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
			Header: http.Header{
				"X-Auth-Token": []string{"12345"},
			},
		}, nil
	} else if url == "https://localhost:9092/ODIM/v1/Sessions" {
		body := `{"Token": ""}`
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}, nil
	}
	if url == "https://localhost:9091/ODIM/v1/Managers/uuid/EthernetInterfaces" && token == "12345" {
		body := `{"data": "/ODIM/v1/Managers/uuid/EthernetInterfaces"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}, nil
	} else if url == "https://localhost:9093/ODIM/v1/Managers/uuid1/EthernetInterfaces" {
		body := `{"data": "/ODIM/v1/Managers/uuid/EthernetInterfaces"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}, nil
	} else if url == "https://localhost:9092/ODIM/v1/Managers/uuid/EthernetInterfaces" && token == "23456" {
		body := `{"data": "/ODIM/v1/Managers/uuid/EthernetInterfaces"}`
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}, nil
	}
	return nil, fmt.Errorf("InvalidRequest")
}

func getEncryptedKey(t *testing.T, key []byte) []byte {
	cryptedKey, err := common.EncryptWithPublicKey(key)
	if err != nil {
		t.Fatalf("error: failed to encrypt data: %v", err)
	}
	return cryptedKey
}
