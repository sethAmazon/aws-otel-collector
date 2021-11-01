/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"strconv"
)

type documentVersionTuple struct {
	str *string
	num int
}

func main() {
	//set up aws go sdk config
	testSession, err := session.NewSession()
	if err != nil {
		log.Fatalf("Error creating session %v", err)
	}
	// set up aws go sdk ssm client
	ssmClient := ssm.New(testSession)

	//get list of documents with versions
	documentName := "testAWSDistroOTel-Collector"
	listDocumentVersionInput := ssm.ListDocumentVersionsInput{Name: &documentName}
	listDocumentVersionsOutput, err := ssmClient.ListDocumentVersions(&listDocumentVersionInput)
	if err != nil {
		log.Fatalf("Error getting document list %v", err)
	} else if len(listDocumentVersionsOutput.DocumentVersions) < 1{
		log.Fatalf("At least one document version must be found")
	}

	versionTuples := make([]documentVersionTuple, 0, len(listDocumentVersionsOutput.DocumentVersions))
	largestVersion := documentVersionTuple{num: -1}
	for _, documentVersion := range listDocumentVersionsOutput.DocumentVersions {
		versionTuple := toVersionTuple(documentVersion)
		if versionTuple.num > largestVersion.num {
			largestVersion.num = versionTuple.num
			largestVersion.str = versionTuple.str
		}
		versionTuples = append(versionTuples, versionTuple)
	}

	//do not delete latest version and make it default
	updateDocumentDefaultVersionInput := ssm.UpdateDocumentDefaultVersionInput{Name: &documentName, DocumentVersion: largestVersion.str}
	_, err = ssmClient.UpdateDocumentDefaultVersion(&updateDocumentDefaultVersionInput)
	if err != nil {
		log.Fatalf("Error updating default document %v", err)
	}
	log.Printf("Updated default version to %d", largestVersion.num)

	for _, version := range versionTuples {
		if largestVersion.num == version.num {
			continue
		}
		deleteDocumentInput := ssm.DeleteDocumentInput{Name: &documentName, DocumentVersion: version.str}
		_, err := ssmClient.DeleteDocument(&deleteDocumentInput)
		if err != nil {
			log.Fatalf("Error deleting document version %d %v", version.num, err)
		}
		log.Printf("Document version deleted %d", version.num)
	}
}

func toVersionTuple(info *ssm.DocumentVersionInfo) documentVersionTuple  {
	versionNumber, err := strconv.Atoi(*info.DocumentVersion)
	if err != nil {
		log.Fatalf("Error parsing version number %v", err)
	}
	return documentVersionTuple{
		str: info.DocumentVersion,
		num: versionNumber,
	}
}
