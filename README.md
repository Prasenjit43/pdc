# Mobile Chaincode for Hyperledger Fabric

This is a Hyperledger Fabric chaincode written in Go that manages mobile device records. It uses both public and private data to maintain information about mobile devices.

## Chaincode Functions

### 1. `CreateMobile`

This function is used to create a new mobile device record. It takes a JSON input containing public data (name, color, size) and transient private data (owner, price).

**Peer Command for CreateMobile:**

```bash
export MOBILE=$(echo -n "{\"doctype\":\"MOBILE_PRIVATE\",\"name\":\"mobile05\",\"owner\":\"parth\",\"price\":100}" | base64 | tr -d \\n)

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n pdc --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"CreateMobile","Args":["{\"doctype\":\"MOBILE\",\"name\":\"mobile05\",\"color\":\"black\",\"size\":35}"]}' --transient "{\"mobile_properties\":\"$MOBILE\"}"
```

### 2. `GetMobilePublicData`

**Peer Command for GetMobilePublicData:**

```bash
peer chaincode query -C mychannel -n pdc -c '{"Args":["GetMobilePublicData", "mobile04"]}' | jq .
```

Retrieves the public data of a mobile device based on the provided mobile ID.

### 3. `GetMobilePrivateDetails`

**Peer Command for GetMobilePrivateDetails:**

```bash
peer chaincode query -C mychannel -n pdc -c '{"Args":["GetMobilePrivateDetails", "mobile04"]}' | jq .
```

Retrieves the private details of a mobile device based on the provided name.

### 4. `IsMobilePrivateDataExist`

**Peer Command for IsMobilePrivateDataExist:**

```bash
export MOBILE=$(echo -n "{\"doctype\":\"MOBILE_PRIVATE\",\"name\":\"mobile04\",\"owner\":\"parth\",\"price\":100}" | base64 | tr -d \\n)

peer chaincode query -C mychannel -n pdc -c '{"Args":["IsMobilePrivateDataExist","Org1MSP"]}' --transient "{\"mobile_properties\":\"$MOBILE\"}"
```

Checks if private data for a mobile device already exists based on the client organization ID and transient data.

### 5. `UpdateMobilePublicData`

**Peer Command for UpdateMobilePublicData:**

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n pdc --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"UpdateMobilePublicData","Args":["{\"mobileId\":\"mobile04\",\"newColor\":\"red\"}"]}'
```

Updates the color of a mobile device based on the provided mobile ID.

### 6. `UpdateMobilePrivateData`

**Peer Command for UpdateMobilePrivateData:**

```bash
export MOBILE=$(echo -n "{\"doctype\":\"MOBILE_PRIVATE\",\"name\":\"mobile05\",\"owner\":\"parth\",\"price\":100}" | base64 | tr -d \\n)
export NEW_MOBILE=$(echo -n "{\"name\":\"mobile05\",\"owner\":\"parth\",\"price\":300}" | base64 | tr -d \\n)

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n pdc --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"UpdateMobilePrivateData","Args":[]}' --transient "{\"mobile_properties\":\"$MOBILE\",\"new_mobile_properties\":\"$NEW_MOBILE\"}"
```

Updates the private data of a mobile device based on the transient data provided. It ensures the existence of the mobile device's private data before making updates.

### 7. `DeleteMobile`

**Peer Command for DeleteMobile:**

```bash
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n pdc --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"DeleteMobile","Args":["mobile05"]}'
```

Deletes both public and private data of a mobile device based on the provided mobile ID.

## Transient Data

The chaincode utilizes transient data to pass private information during transactions. It ensures the confidentiality of sensitive data.

## Implicit Collections

Private data is stored in implicit collections, with collection names based on the client organization ID.

## Endorsement Policy

The chaincode ensures that private data is endorsed by the required parties.
