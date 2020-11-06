package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

type TableXML struct {
	XMLName xml.Name `xml:"table"`
	Text    string   `xml:",chardata"`
	Name    string   `xml:"name,attr"`
	G0      string   `xml:"g0,attr"`
	G1      string   `xml:"g1,attr"`
	G2      string   `xml:"g2,attr"`
	Desc    struct {
		Text string `xml:",chardata"`
		Lang string `xml:"lang,attr"`
	} `xml:"desc"`
	Tag []TagXML `xml:"tag"`
}

type TagXML struct {
	XMLName  xml.Name `xml:"tag"`
	Text     string   `xml:",chardata"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Type     string   `xml:"type,attr"`
	Writable string   `xml:"writable,attr"`
	G2       string   `xml:"g2,attr"`
	Desc     []struct {
		Text string `xml:",chardata"`
		Lang string `xml:"lang,attr"`
	} `xml:"desc"`
}

type TagJSON struct {
	Writable    bool              `json:"writable"`
	Path        string            `json:"path"`
	Group       string            `json:"group"`
	Description map[string]string `json:"description"`
	Type        string            `json:"type"`
}

func DecodeXML(in io.Reader, out io.Writer) error {
	xmlDecoder := xml.NewDecoder(in)

	out.Write([]byte("{\"tags\": [{")) // Write start of answer for tags array

	jsonEncoder := json.NewEncoder(out)

	var finishedFirst bool // variable used to check if we need to wirte a comma before our tag
DecoderLoop:
	for {
		token, err := xmlDecoder.Token()
		if err != nil {
			return fmt.Errorf("xml decoding failed: %w", err)
			break
		}

		switch tokenType := token.(type) {
		case xml.StartElement:
			switch tokenType.Name.Local {
			case "table":
				if finishedFirst { // write comma for every array element except for the first
					out.Write([]byte(","))
				} else {
					finishedFirst = true // we can set this here, because even if the first one fails we would not continue
				}

				// Decode XML from XML Decoder
				var table TableXML
				err := xmlDecoder.DecodeElement(&table, &tokenType)
				if err != nil {
					return fmt.Errorf("xml decoding failed on tag: %w", err)
					break DecoderLoop
				}

				// Encode all tags from table to the Writer
				for i := range table.Tag {
					tagJSON, err := tagXMLtoJson(table.Tag[i], table.Name)
					if err != nil {
						return fmt.Errorf("xml to json transformation failed: %w", err)
						break DecoderLoop
					}
					err = jsonEncoder.Encode(tagJSON)
					if err != nil {
						return fmt.Errorf("json encoding of tags failed: %w", err)
						break DecoderLoop
					}
				}

			}

		}

	}
	return nil

}

func tagXMLtoJson(xmlTag TagXML, tableName string) (jsonTag TagJSON, err error) {
	var description = make(map[string]string)
	for i := range xmlTag.Desc {
		description[xmlTag.Desc[i].Lang] = xmlTag.Desc[i].Text
	}

	writable, err := strconv.ParseBool(xmlTag.Writable)
	if err != nil {
		return TagJSON{}, fmt.Errorf("transforming xml attribute 'writeable' to boolean failed: %w", err)
	}

	return TagJSON{
		Writable:    writable,
		Path:        fmt.Sprintf("%s:%s", tableName, xmlTag.Name),
		Group:       tableName,
		Description: description,
		Type:        xmlTag.Type,
	}, nil
}
