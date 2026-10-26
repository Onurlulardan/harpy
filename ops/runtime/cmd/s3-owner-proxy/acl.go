package main

import (
	"encoding/xml"
	"fmt"
)

type minIOACL struct {
	XMLName xml.Name `xml:"AccessControlPolicy"`
	Owner   struct {
		ID string `xml:"ID"`
	} `xml:"Owner"`
	Grants []struct {
		Grantee struct {
			Type string `xml:"Type"`
		} `xml:"Grantee"`
		Permission string `xml:"Permission"`
	} `xml:"AccessControlList>Grant"`
}

func validateMinIOACL(document []byte) error {
	var value minIOACL
	if xml.Unmarshal(document, &value) != nil || value.XMLName.Local != "AccessControlPolicy" ||
		len(value.Grants) != 1 || value.Grants[0].Grantee.Type != "CanonicalUser" ||
		value.Grants[0].Permission != "FULL_CONTROL" {
		return fmt.Errorf("unexpected S3 ACL response")
	}
	return nil
}

func privateACL(owner string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` +
		`<Owner><ID>` + owner + `</ID><DisplayName>harpy-host</DisplayName></Owner>` +
		`<AccessControlList><Grant>` +
		`<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" ` +
		`xsi:type="CanonicalUser"><ID>` + owner + `</ID>` +
		`<DisplayName>harpy-host</DisplayName></Grantee>` +
		`<Permission>FULL_CONTROL</Permission></Grant></AccessControlList>` +
		`</AccessControlPolicy>`)
}
