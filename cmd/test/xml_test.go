package test

import (
	"encoding/xml"
	"testing"
)

var expect = `
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
	xmlns:tem="http://tempuri.org/" xmlns:gxw="http://schemas.datacontract.org/2004/07/GXWCF2">
	<soapenv:Body>
		<tem:GetRecord>
			<tem:LD>
				<gxw:LogonType>0</gxw:LogonType>
				<gxw:Password></gxw:Password>
				<gxw:UserName>admin</gxw:UserName>
			</tem:LD>
			<tem:nTableID>107</tem:nTableID>
			<tem:nParentID>1</tem:nParentID>
			<tem:nRecordID>0</tem:nRecordID>
			<tem:strXML></tem:strXML>
			<tem:nErrorCode>0</tem:nErrorCode>
			<tem:strErrorXML></tem:strErrorXML>
		</tem:GetRecord>
	</soapenv:Body>
</soapenv:Envelope>
`

var expect2 = `
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
	<Body>
		<GetRecord xmlns="http://tempuri.org/">
			<LD>
				<LogonType xmlns="http://schemas.datacontract.org/2004/07/GXWCF2">0</LogonType>
				<Password xmlns="http://schemas.datacontract.org/2004/07/GXWCF2"></Password>
				<UserName xmlns="http://schemas.datacontract.org/2004/07/GXWCF2">admin</UserName>
			</LD>
			<nTableID>107</nTableID>
			<nParentID>1</nParentID>
			<nRecordID>0</nRecordID>
			<strXML></strXML>
			<nErrorCode>0</nErrorCode>
			<strErrorXML></strErrorXML>
		</GetRecord>
	</Body>
</Envelope>
`

type Envelope struct {
	XMLName xml.Name `xml:"soapenv:Envelope"`
	NS0     string   `xml:"xmlns:soapenv,attr"`
	NS1     string   `xml:"xmlns:ns1,attr"`
	NS2     string   `xml:"xmlns:ns2,attr"`
	NS3     string   `xml:"xmlns:ns3,attr"`
	NS4     string   `xml:"xmlns:ns4,attr"`
	NS5     string   `xml:"xmlns:ns5,attr"`
	Body    any      `xml:"soapenv:Body"`
}

func TestName(t *testing.T) {
	r := Envelope{
		NS0: "http://schemas.xmlsoap.org/soap/envelope/",
		NS1: "http://tempuri.org/",
		NS2: "http://schemas.microsoft.com/2003/10/Serialization/",
		NS3: "http://schemas.datacontract.org/2004/07/GXWCF2",
		NS4: "http://schemas.datacontract.org/2004/07/GXSV",
		NS5: "http://schemas.microsoft.com/2003/10/Serialization/Arrays",
		Body: AddRecord{
			LD: &Logon{
				LogonType: 1,
			},
			NTableID: 2,
			NSiteID:  3,
		},
	}

	out, err := xml.MarshalIndent(&r, "", "    ")
	if err != nil {
		t.Error(err)
	}

	t.Logf("\n" + string(out))
}
