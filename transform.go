package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

// TableXML is the XML structure for <table>
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

// TagXML is the XML structure for <tag>
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

// TagJSON describes the array elements we want to output in our stream
type TagJSON struct {
	Writable    bool              `json:"writable"`
	Path        string            `json:"path"`
	Group       string            `json:"group"`
	Description map[string]string `json:"description"`
	Type        string            `json:"type"`
}

// DecodeXML will take an XML stream from in and write converted JSON to out table by table.
// If anything goes wrong an error will be returned and nothing else will be written.
func DecodeXML(in io.Reader, out io.Writer) error {
	xmlDecoder := xml.NewDecoder(in)

	out.Write([]byte("{\"tags\": [")) // Write start of answer for tags array

	jsonEncoder := json.NewEncoder(out)

	var finishedFirst bool // variable used to check if we need to wirte a comma before our tag
DecoderLoop: // We need a named loop to break out of inner switch cases
	for {
		token, err := xmlDecoder.Token()
		if err != nil && err == io.EOF {
			_, err := out.Write([]byte("]}")) // Write end of answer for tags array
			if err != nil {
				return fmt.Errorf("xml decoding failed: %w", err)
				break
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("xml decoding failed: %w", err)
			break
		}

		// Check every start token "<" if we have a table element and decode and stream the single table.
		switch tokenType := token.(type) {
		case xml.StartElement:
			switch tokenType.Name.Local {
			case "table":
				// Decode XML from XML Decoder
				var table TableXML
				err := xmlDecoder.DecodeElement(&table, &tokenType)
				if err != nil {
					return fmt.Errorf("xml decoding failed on tag: %w", err)
					break DecoderLoop
				}

				// Encode all tags from table to the Writer
				for i := range table.Tag {
					if finishedFirst { // write comma for every array element except for the first to produce valid JSON
						out.Write([]byte(","))
					} else {
						finishedFirst = true // we can set this here, because even if the first one fails we would not continue
					}
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
		default:

		}

	}
	return nil

}

// tagXMLtoJson converts a single tag from XML to JSON according to the given specifications.
func tagXMLtoJson(xmlTag TagXML, tableName string) (jsonTag TagJSON, err error) {
	var description = make(map[string]string) // Description is just a json object with string fields, which is basically a map.
	for i := range xmlTag.Desc {
		description[xmlTag.Desc[i].Lang] = xmlTag.Desc[i].Text
	}

	writable, err := strconv.ParseBool(xmlTag.Writable) // The only thing that could go wrong is the Writeable field not being true or false.
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
