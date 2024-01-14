/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	// "github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const index = "name~doctype"
const MOBILE = "MOBILE"
const MOBILE_PRIVATE = "MOBILE_PRIVATE"

type Mobile struct {
	DocType string `json:"doctype"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Size    int    `json:"size"`
}

type MobilePrivateData struct {
	DocType string `json:"doctype"`
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Price   int    `json:"price"`
}

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) CreateMobile(ctx contractapi.TransactionContextInterface, publicDescInput string) error {
	var publicData Mobile
	var privateData MobilePrivateData

	//Read public data
	err := json.Unmarshal([]byte(publicDescInput), &publicData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input public description %v", err.Error())
	}
	fmt.Println("Input Data :", publicData)

	//Read private data
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	transientMobileJSON, ok := transientMap["mobile_properties"]
	if !ok {
		return fmt.Errorf("mobile_properties key not found in the transient map")
	}
	err = json.Unmarshal(transientMobileJSON, &privateData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal transientMobileJSON: %v", err)
	}

	//creating composite key for public record
	publicCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{publicData.Name, MOBILE})
	if err != nil {
		return fmt.Errorf("failed to create composite key for public data %v", err.Error())
	}

	//validate mobile public record
	mobilePublicBytes, err := ctx.GetStub().GetState(publicCompositeKey)
	if err != nil {
		return fmt.Errorf("failed to get mobile public data: %v", err)
	}
	if mobilePublicBytes != nil {
		return fmt.Errorf("record already exist for mobile: %v", publicData.Name)
	}

	//insert mobile public record
	// publicData.DocType = MOBILE
	publicDataJSON, err := json.Marshal(publicData)
	if err != nil {
		return fmt.Errorf("failed to marshal public record: %v", publicData.Name)
	}
	err = ctx.GetStub().PutState(publicCompositeKey, publicDataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert mobile public data: %v", err)
	}

	// Get the clientOrgId from the input, will be used for implicit collection, owner, and state-based endorsement policy
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Client org id :", clientOrgID)

	//creating composite key for private record
	privateCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{privateData.Name, MOBILE_PRIVATE})
	if err != nil {
		fmt.Errorf("failed to create composite key for private data %v", err.Error())
	}

	collection := buildCollectionName(clientOrgID)
	fmt.Println("Collection name :", collection)

	// privateData.DocType = MOBILE_PRIVATE
	privateDataJSON, err := json.Marshal(privateData)
	if err != nil {
		return fmt.Errorf("failed to marshal private record: %v", privateData.Name)
	}
	err = ctx.GetStub().PutPrivateData(collection, privateCompositeKey, privateDataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert mobile private data: %v", err)
	}
	return nil
}

func (s *SmartContract) GetMobilePublicData(ctx contractapi.TransactionContextInterface, mobileId string) (*Mobile, error) {
	publicMobileBytes, err := readMobilePublicData(ctx, mobileId)
	if err != nil {
		return nil, err
	}

	var mobilePublicData Mobile
	err = json.Unmarshal(publicMobileBytes, &mobilePublicData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public mobile data bytes file : %v", err.Error())
	}

	return &mobilePublicData, nil
}

func (s *SmartContract) GetMobilePrivateDetails(ctx contractapi.TransactionContextInterface, name string) (*MobilePrivateData, error) {

	mobileAsBytes, err := readMobilePrivateData(ctx, name)
	if err != nil {
		return nil, err
	}

	fmt.Println("mobileAsBytes :", string(mobileAsBytes))

	mobilePrivateData := MobilePrivateData{}
	err = json.Unmarshal(mobileAsBytes, &mobilePrivateData) //unmarshal it aka JSON.parse()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	return &mobilePrivateData, nil
}

func (s *SmartContract) IsMobilePrivateDataExist(ctx contractapi.TransactionContextInterface, clientOrgID string) (bool, error) {
	var privateData MobilePrivateData
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return false, fmt.Errorf("error getting transient: %v", err)
	}

	transientMobileJSON, ok := transientMap["mobile_properties"]
	if !ok {
		return false, fmt.Errorf("mobile_properties key not found in the transient map")
	}

	fmt.Println("transientMobileJSON :", string(transientMobileJSON))

	err = json.Unmarshal(transientMobileJSON, &privateData)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal transientMobileJSON: %v", err)
	}
	// privateData.DocType = MOBILE_PRIVATE
	fmt.Println("privateData :", &privateData)

	//create composite key
	privateCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{privateData.Name, MOBILE_PRIVATE})
	if err != nil {
		return false, fmt.Errorf("failed to create composite key for private data %v", err.Error())
	}
	fmt.Println("Client org id :", clientOrgID)

	collection := buildCollectionName(clientOrgID)
	immutablePropertiesOnChainHash, err := ctx.GetStub().GetPrivateDataHash(collection, privateCompositeKey)
	fmt.Println("immutablePropertiesOnChainHash :", (immutablePropertiesOnChainHash))

	hash := sha256.New()
	hash.Write(transientMobileJSON)
	calculatedPropertiesHash := hash.Sum(nil)
	fmt.Println("calculatedPropertiesHash :", (calculatedPropertiesHash))

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(immutablePropertiesOnChainHash, calculatedPropertiesHash) {
		return false, fmt.Errorf("hash %x for passed immutable properties %s does not match on-chain hash %x",
			calculatedPropertiesHash,
			transientMobileJSON,
			immutablePropertiesOnChainHash,
		)
	}

	return true, nil

}

func (s *SmartContract) UpdateMobilePublicData(ctx contractapi.TransactionContextInterface, inputData string) error {
	dataToUpdate := struct {
		MobileId string `json:"mobileId"`
		NewColor string `json:"newColor"`
	}{}

	err := json.Unmarshal([]byte(inputData), &dataToUpdate)

	var mobilePublicData Mobile
	publicMobileBytes, err := readMobilePublicData(ctx, dataToUpdate.MobileId)
	if err != nil {
		return err
	}

	err = json.Unmarshal(publicMobileBytes, &mobilePublicData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the object bytes: %v", err.Error())
	}

	mobilePublicData.Color = dataToUpdate.NewColor
	fmt.Println("Public Updated Values :", mobilePublicData)

	compositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{dataToUpdate.MobileId, MOBILE})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err.Error())
	}

	publicDataBytes, err := json.Marshal(mobilePublicData)
	if err != nil {
		return fmt.Errorf("failed to marshal public record: %v", dataToUpdate.MobileId)
	}
	err = ctx.GetStub().PutState(compositeKey, publicDataBytes)
	if err != nil {
		return fmt.Errorf("failed to insert mobile public data: %v", err)
	}

	return nil
}

func (s *SmartContract) UpdateMobilePrivateData(ctx contractapi.TransactionContextInterface) error {
	var existingData MobilePrivateData
	var newData MobilePrivateData

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// /**************/
	transientMobileJSON, ok := transientMap["mobile_properties"]
	if !ok {
		return fmt.Errorf("mobile_properties key not found in the transient map")
	}
	fmt.Println("transientMobileJSON :", string(transientMobileJSON))

	err = json.Unmarshal(transientMobileJSON, &existingData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal transientMobileJSON: %v", err)
	}
	fmt.Println("existingData :", &existingData)
	// /***************/

	transientNewDataJSON, ok := transientMap["new_mobile_properties"]
	if !ok {
		return fmt.Errorf("new_mobile_properties key not found in the transient map")
	}
	fmt.Println("transientNewDataJSON :", string(transientNewDataJSON))

	err = json.Unmarshal(transientNewDataJSON, &newData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal transientNewDataJSON: %v", err)
	}
	fmt.Println("newInputData :", &newData)

	if existingData.Name != newData.Name {
		return fmt.Errorf("Mismatched mobile name")
	}

	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}

	isExist, err := s.IsMobilePrivateDataExist(ctx, clientOrgID)
	fmt.Println("isExist :", isExist)
	if !isExist {
		return err
	}

	if newData.Owner != "" {
		existingData.Owner = newData.Owner
	}

	if newData.Price != 0 {
		existingData.Price = newData.Price
	}

	newDataCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{existingData.Name, MOBILE_PRIVATE})
	fmt.Println("newDataCompositeKey :", newDataCompositeKey)
	collection := buildCollectionName(clientOrgID)
	fmt.Println("Collection name :", collection)

	newPrivateDataBytes, err := json.Marshal(existingData)
	if err != nil {
		return fmt.Errorf("failed to marshal private record: %v", existingData.Name)
	}

	fmt.Println("New Private Details :", string(newPrivateDataBytes))

	err = ctx.GetStub().PutPrivateData(collection, newDataCompositeKey, newPrivateDataBytes)
	if err != nil {
		return fmt.Errorf("failed to insert mobile private data: %v", err)
	}

	fmt.Println("*********************************")

	return nil
}

func (s *SmartContract) DeleteMobile(ctx contractapi.TransactionContextInterface, mobileId string) error {

	_, err := readMobilePublicData(ctx, mobileId)
	if err != nil {
		return err
	}

	//delete public data
	publicCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{mobileId, MOBILE})
	if err != nil {
		fmt.Errorf("failed to create composite key for public data %v", err.Error())
	}
	fmt.Println("Public composite key :", publicCompositeKey)

	privateCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{mobileId, MOBILE_PRIVATE})
	if err != nil {
		fmt.Errorf("failed to create composite key for private data %v", err.Error())
	}
	fmt.Println("Private composite key :", privateCompositeKey)

	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Client org id :", clientOrgID)

	collection := buildCollectionName(clientOrgID)
	err = ctx.GetStub().DelPrivateData(collection, privateCompositeKey)
	if err != nil {
		return fmt.Errorf("failed to delete mobile private data: %v", err)
	}
	fmt.Println("Private Data deleted successfully")

	err = ctx.GetStub().DelState(publicCompositeKey)
	if err != nil {
		return fmt.Errorf("failed to delete mobile public data: %v", err)
	}
	fmt.Println("Public Data deleted successfully")

	return nil
}

func buildCollectionName(clientOrgID string) string {
	return fmt.Sprintf("_implicit_org_%s", clientOrgID)
}

func getClientOrgID(ctx contractapi.TransactionContextInterface) (string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	return clientOrgID, nil
}

func readMobilePublicData(ctx contractapi.TransactionContextInterface, mobileId string) ([]byte, error) {

	compositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{mobileId, MOBILE})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err.Error())
	}
	publicMobileBytes, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read public data: %v", err.Error())
	}
	if publicMobileBytes == nil {
		return nil, fmt.Errorf("data not present for : %v", mobileId)
	}

	fmt.Println("Mobile Public Data in Bytes :", string(publicMobileBytes))

	return publicMobileBytes, nil
}

func readMobilePrivateData(ctx contractapi.TransactionContextInterface, mobileId string) ([]byte, error) {
	privateCompositeKey, err := ctx.GetStub().CreateCompositeKey(index, []string{mobileId, MOBILE_PRIVATE})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key for private data %v", err.Error())
	}
	fmt.Println("privateCompositeKey :", privateCompositeKey)

	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("Client org id :", clientOrgID)

	collection := buildCollectionName(clientOrgID)
	fmt.Println("Collection name :", collection)

	mobileAsBytes, err := ctx.GetStub().GetPrivateData(collection, privateCompositeKey)

	fmt.Println("Err for mobileAsBytes : ", err)
	if err != nil {
		return nil, fmt.Errorf("failed to get Mobile 11111: %v", err.Error())
	}

	// No Asset found, return empty response
	if mobileAsBytes == nil {
		// log.Printf("%v does not exist in collection %v", assetID, assetCollection)
		// return nil, nil
		return nil, fmt.Errorf("Mobile does not exist: %v", mobileId)
	}

	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to get Mobile 11111:" + err.Error())
	// } else if mobileAsBytes == nil {
	// 	return nil, fmt.Errorf("Mobile does not exist: " + mobileId)
	// }
	fmt.Println("Mobile Private Data in Bytes :", string(mobileAsBytes))

	return mobileAsBytes, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
